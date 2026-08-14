package cmd

import (
	"context"
	"io"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rushikeshg25/coolDb/server"
)

func TestUnifiedExecWorkflowPersistsAcrossRestart(t *testing.T) {
	port := availablePort(t)
	databasePath := filepath.Join(t.TempDir(), "integration.cooldb")

	stopServer := startTestServer(t, port, databasePath)
	commands := []string{
		"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)",
		"INSERT INTO users VALUES (1, 'Ada')",
	}
	for _, query := range commands {
		if _, err := executeCommand(t, NewRootCommand, "exec", "--host", "127.0.0.1", "--port", strconv.Itoa(port), query); err != nil {
			stopServer()
			t.Fatalf("exec %q error = %v", query, err)
		}
	}
	stopServer()

	stopServer = startTestServer(t, port, databasePath)
	defer stopServer()
	output, err := executeCommand(t, NewRootCommand, "exec", "--host", "127.0.0.1", "--port", strconv.Itoa(port), "SELECT * FROM users")
	if err != nil {
		t.Fatalf("SELECT after restart error = %v", err)
	}
	for _, expected := range []string{"id", "name", "1", "Ada", "(1 row(s))"} {
		if !strings.Contains(output, expected) {
			t.Errorf("SELECT output = %q, want it to contain %q", output, expected)
		}
	}

	_, err = executeCommand(t, NewRootCommand, "exec", "--host", "127.0.0.1", "--port", strconv.Itoa(port), "SELECT * FROM missing")
	if err == nil || !strings.Contains(err.Error(), "NotFound") {
		t.Errorf("missing table error = %v, want NotFound", err)
	}
}

func availablePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func startTestServer(t *testing.T, port int, databasePath string) func() {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- server.Run(ctx, server.Config{
			Host:         "127.0.0.1",
			Port:         port,
			DatabasePath: databasePath,
			Output:       io.Discard,
		})
	}()

	return func() {
		cancel()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("server shutdown error = %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Errorf("server on port %d did not stop", port)
		}
	}
}
