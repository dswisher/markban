package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/dswisher/markban/internal/board"
	"github.com/dswisher/markban/internal/render"
)

const debounceDuration = 200 * time.Millisecond

// Server is a live-reload HTTP server for a Markban board.
type Server struct {
	boardDir string
	buildDir string
	port     int

	mu      sync.Mutex
	clients []chan struct{}
}

// New creates a new Server for the given board directory.
func New(boardDir string) *Server {
	return &Server{
		boardDir: boardDir,
		buildDir: filepath.Join(boardDir, ".build"),
	}
}

// Port returns the port the server is listening on. Only valid after Run has
// started the listener.
func (s *Server) Port() int {
	return s.port
}

// Run starts the HTTP server and file watcher. It blocks until ctx is
// cancelled, then shuts down cleanly.
func (s *Server) Run(ctx context.Context) error {
	// Create build directory and do initial render.
	if err := s.rebuild(); err != nil {
		return fmt.Errorf("initial render: %w", err)
	}

	// Set up file watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	if err := s.watchDir(watcher); err != nil {
		return fmt.Errorf("setting up watcher: %w", err)
	}

	// Set up HTTP handlers.
	mux := http.NewServeMux()
	mux.HandleFunc("/events", s.handleSSE)
	mux.HandleFunc("/", s.handleIndex)

	// Find a free port.
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		// Fallback to any free port.
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			return fmt.Errorf("starting listener: %w", err)
		}
	}
	s.port = listener.Addr().(*net.TCPAddr).Port

	srv := &http.Server{Handler: mux}

	// Start file watcher goroutine.
	go s.watchLoop(ctx, watcher)

	// Start HTTP server in background.
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	log.Printf("Serving board at http://localhost:%d", s.port)

	// Block until context is cancelled or the server fails.
	select {
	case <-ctx.Done():
	case err := <-serverErr:
		return err
	}

	// Graceful shutdown.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// rebuild loads the board from disk and renders it to the build directory.
func (s *Server) rebuild() error {
	b, err := board.LoadBoard(s.boardDir)
	if err != nil {
		return fmt.Errorf("loading board: %w", err)
	}

	if err := render.RenderToDir(b, s.buildDir); err != nil {
		return fmt.Errorf("rendering board: %w", err)
	}

	return nil
}

// watchDir adds the board directory and all immediate subdirectories (columns)
// to the watcher. The .build directory is ignored.
func (s *Server) watchDir(watcher *fsnotify.Watcher) error {
	if err := watcher.Add(s.boardDir); err != nil {
		return err
	}

	entries, err := os.ReadDir(s.boardDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue // skip hidden dirs including .build
		}
		if err := watcher.Add(filepath.Join(s.boardDir, name)); err != nil {
			return err
		}
	}

	return nil
}

// watchLoop listens for fsnotify events and triggers rebuilds with debounce.
func (s *Server) watchLoop(ctx context.Context, watcher *fsnotify.Watcher) {
	var timer *time.Timer

	for {
		select {
		case <-ctx.Done():
			if timer != nil {
				timer.Stop()
			}
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Ignore events inside the .build directory.
			if strings.Contains(filepath.ToSlash(event.Name), "/.build/") ||
				strings.HasSuffix(filepath.ToSlash(event.Name), "/.build") {
				continue
			}

			// Only care about .md files and directory additions/removals.
			if !isRelevantEvent(event) {
				continue
			}

			// Debounce: reset the timer on every relevant event.
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(debounceDuration, func() {
				s.triggerRebuild(watcher)
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("watcher error: %v", err)
		}
	}
}

// isRelevantEvent returns true for events we care about: .md files or
// directory create/remove.
func isRelevantEvent(event fsnotify.Event) bool {
	if strings.HasSuffix(event.Name, ".md") {
		return true
	}
	// Directory-level creates and removes (new or deleted columns).
	if event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
		return true
	}
	return false
}

// triggerRebuild re-renders the board and notifies all SSE clients.
func (s *Server) triggerRebuild(watcher *fsnotify.Watcher) {
	// Re-add any new subdirectories to the watcher.
	_ = s.watchDir(watcher)

	if err := s.rebuild(); err != nil {
		log.Printf("rebuild error: %v", err)
		return
	}

	s.notifyClients()
}

// notifyClients sends a reload signal to all connected SSE clients.
func (s *Server) notifyClients() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ch := range s.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// addClient registers a new SSE client channel and returns it.
func (s *Server) addClient() chan struct{} {
	ch := make(chan struct{}, 1)
	s.mu.Lock()
	s.clients = append(s.clients, ch)
	s.mu.Unlock()
	return ch
}

// removeClient unregisters an SSE client channel.
func (s *Server) removeClient(ch chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, c := range s.clients {
		if c == ch {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return
		}
	}
}

// handleIndex serves index.html directly for all non-event requests,
// avoiding http.FileServer's redirect of /index.html → ./ which some
// browsers (e.g. Chrome) mishandle.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.buildDir, "index.html"))
}

// handleSSE is the HTTP handler for the /events SSE endpoint.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch := s.addClient()
	defer s.removeClient(ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			_, err := fmt.Fprintf(w, "event: reload\ndata: {}\n\n")
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
