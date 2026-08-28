package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rushikeshg25/coolDb/internal/core"
	"github.com/rushikeshg25/coolDb/internal/database"
	"github.com/rushikeshg25/coolDb/internal/httpapi"
)

type Config struct {
	Host         string
	Port         int
	HTTPPort     int
	DatabasePath string
	Output       io.Writer
}

func Run(ctx context.Context, config Config) error {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	printBanner(config.Output)
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 3040
	}

	if config.DatabasePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		config.DatabasePath = filepath.Join(homeDir, "cooldb", "default.cooldb")
	}
	databasePath, err := filepath.Abs(config.DatabasePath)
	if err != nil {
		return fmt.Errorf("resolve database path: %w", err)
	}
	engine, err := database.Open(databasePath)
	if err != nil {
		return err
	}

	coreServer := core.NewCoreServer(config.Host, config.Port, engine)

	slog.Info("Server starting", "host", config.Host, "port", config.Port, "database", databasePath)
	if config.HTTPPort <= 0 {
		if err := core.BindAndListen(ctx, coreServer); err != nil {
			return err
		}
		slog.Info("Server shut down gracefully")
		return nil
	}
	if err := runWithHTTP(ctx, coreServer, engine, config.Host, config.HTTPPort); err != nil {
		return err
	}
	slog.Info("Server shut down gracefully")
	return nil
}

func runWithHTTP(ctx context.Context, coreServer *core.CoreServer, engine *database.Engine, host string, port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return fmt.Errorf("start demo HTTP server: %w", err)
	}
	httpServer := &http.Server{
		Handler:           httpapi.NewHandler(engine),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	runContext, cancel := context.WithCancel(ctx)
	defer cancel()

	httpDone := make(chan error, 1)
	go func() {
		slog.Info("Local demo API starting", "address", listener.Addr().String())
		httpDone <- normalizeHTTPError(httpServer.Serve(listener))
	}()
	go func() {
		<-runContext.Done()
		shutdownContext, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_ = httpServer.Shutdown(shutdownContext)
	}()

	grpcDone := make(chan error, 1)
	go func() { grpcDone <- core.BindAndListen(runContext, coreServer) }()

	var firstError error
	select {
	case firstError = <-grpcDone:
		cancel()
		if httpError := <-httpDone; firstError == nil {
			firstError = httpError
		}
	case firstError = <-httpDone:
		cancel()
		if grpcError := <-grpcDone; firstError == nil {
			firstError = grpcError
		}
	case <-ctx.Done():
		cancel()
		grpcError := <-grpcDone
		httpError := <-httpDone
		if grpcError != nil {
			firstError = grpcError
		} else {
			firstError = httpError
		}
	}
	return firstError
}

func normalizeHTTPError(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return fmt.Errorf("demo HTTP server: %w", err)
}

func printBanner(writer io.Writer) {
	fmt.Fprint(writer, `
 ██████╗ ██████╗  ██████╗ ██╗     ██████╗ ██████╗
██╔════╝██╔═══██╗██╔═══██╗██║     ██╔══██╗██╔══██╗
██║     ██║   ██║██║   ██║██║     ██║  ██║██████╔╝
██║     ██║   ██║██║   ██║██║     ██║  ██║██╔══██╗
╚██████╗╚██████╔╝╚██████╔╝███████╗██████╔╝██████╔╝
 ╚═════╝ ╚═════╝  ╚═════╝ ╚══════╝╚═════╝ ╚═════╝
`)
}
