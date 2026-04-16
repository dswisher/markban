package server

import (
	"bufio"
	"context"
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

// TestServer_Rebuild verifies that rebuild creates .build/index.html with
// rendered board content, without starting a network listener.
func TestServer_Rebuild(t *testing.T) {
	boardDir := makeTestBoardDir(t)
	s := New(boardDir, true)

	require.NoError(t, s.rebuild())

	data, err := os.ReadFile(filepath.Join(boardDir, ".build", "index.html"))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(strings.TrimSpace(string(data)), "<!DOCTYPE html>"))
	assert.Contains(t, string(data), "Test Task")
}

// TestServer_HandleIndex verifies that handleIndex serves the rendered HTML.
func TestServer_HandleIndex(t *testing.T) {
	boardDir := makeTestBoardDir(t)
	s := New(boardDir, true)

	require.NoError(t, s.rebuild())

	ts := httptest.NewServer(http.HandlerFunc(s.handleIndex))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))
}
