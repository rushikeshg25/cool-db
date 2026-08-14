package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/rushikeshg25/coolDb/internal/core"
	"github.com/rushikeshg25/coolDb/internal/database"
)

func Start(host string, port int, databasePath string) error {
	printBanner()
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 3040
	}

	if databasePath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home directory: %w", err)
		}
		databasePath = filepath.Join(homeDir, "cooldb", "default.cooldb")
	}
	databasePath, err := filepath.Abs(databasePath)
	if err != nil {
		return fmt.Errorf("resolve database path: %w", err)
	}
	engine, err := database.Open(databasePath)
	if err != nil {
		return err
	}

	coreServer := core.NewCoreServer(host, port, engine)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("Server starting", "host", host, "port", port, "database", databasePath)
	if err := core.BindAndListen(ctx, coreServer); err != nil {
		return err
	}
	slog.Info("Server shut down gracefully")
	return nil
}

func printBanner() {
	fmt.Print(`
 ██████╗ ██████╗  ██████╗ ██╗     ██████╗ ██████╗
██╔════╝██╔═══██╗██╔═══██╗██║     ██╔══██╗██╔══██╗
██║     ██║   ██║██║   ██║██║     ██║  ██║██████╔╝
██║     ██║   ██║██║   ██║██║     ██║  ██║██╔══██╗
╚██████╗╚██████╔╝╚██████╔╝███████╗██████╔╝██████╔╝
 ╚═════╝ ╚═════╝  ╚═════╝ ╚══════╝╚═════╝ ╚═════╝
`)
}
