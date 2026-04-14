package server

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTestBoardDir creates a minimal board directory for testing.
func makeTestBoardDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	colDir := filepath.Join(dir, "1-todo")
	require.NoError(t, os.MkdirAll(colDir, 0o755))

	taskContent := "# Test Task\n\nA test blurb.\n"
	require.NoError(t, os.WriteFile(filepath.Join(colDir, "task.md"), []byte(taskContent), 0o644))

	return dir
}

// TestSSEHandler_ReceivesReloadEvent starts an SSE handler in a test HTTP
// server, triggers a reload notification via a separate request, and verifies
// the client receives the "reload" event.
func TestSSEHandler_ReceivesReloadEvent(t *testing.T) {
	s := &Server{ctx: context.Background()}

	ts := httptest.NewServer(http.HandlerFunc(s.handleSSE))
	defer ts.Close()

	// Open SSE connection with a context so we can cancel it when done.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	require.NoError(t, err)

	// Start the request in a goroutine; it will block reading the stream.
	events := make(chan string, 4)
	connEstablished := make(chan struct{})

	go func() {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		close(connEstablished)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				events <- line
			}
		}
	}()

	// Wait for the SSE connection to be established, then notify.
	select {
	case <-connEstablished:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for SSE connection to establish")
	}

	// Give handleSSE time to register the client after sending headers.
	time.Sleep(50 * time.Millisecond)
	s.notifyClients()

	// Expect to receive the "event: reload" line within 2 seconds.
	select {
	case line := <-events:
		assert.Equal(t, "event: reload", line)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for reload event")
	}
}

// TestServer_InitialRender verifies that Run creates the .build/index.html
// file during startup.
func TestServer_InitialRender(t *testing.T) {
	boardDir := makeTestBoardDir(t)
	s := New(boardDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = s.Run(ctx)
	}()

	// Poll until index.html appears (max 2 seconds).
	indexPath := filepath.Join(boardDir, ".build", "index.html")
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(indexPath); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	data, err := os.ReadFile(indexPath)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(strings.TrimSpace(string(data)), "<!DOCTYPE html>"))
	assert.Contains(t, string(data), "Test Task")

	// Poll until the server port is set, then verify it's responding.
	var port int
	for time.Now().Before(deadline) {
		if p := s.Port(); p != 0 {
			port = p
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	require.NotZero(t, port, "server port should be set")

	url := fmt.Sprintf("http://localhost:%d/", port)
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
