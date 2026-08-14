package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/rushikeshg25/coolDb/internal/core"
	"github.com/rushikeshg25/coolDb/internal/database"
)

type Config struct {
	Host         string
	Port         int
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
	if err := core.BindAndListen(ctx, coreServer); err != nil {
		return err
	}
	slog.Info("Server shut down gracefully")
	return nil
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
