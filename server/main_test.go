package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"
)

func TestRunStartsAndStopsDemoHTTPServer(t *testing.T) {
	grpcPort := availablePort(t)
	httpPort := availablePort(t)
	databasePath := filepath.Join(t.TempDir(), "server-test.cooldb")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, Config{
			Host:         "127.0.0.1",
			Port:         grpcPort,
			HTTPPort:     httpPort,
			DatabasePath: databasePath,
			Output:       io.Discard,
		})
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d/api/health", httpPort)
	var response *http.Response
	var err error
	for deadline := time.Now().Add(3 * time.Second); time.Now().Before(deadline); {
		response, err = http.Get(url) // #nosec G107 -- test URL is a loopback address.
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		cancel()
		t.Fatalf("GET health: %v", err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Errorf("health status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() shutdown error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not stop after context cancellation")
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
