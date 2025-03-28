package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/rushikeshg25/coolDb/internal/core"
)

func Start(Host string, Port int) {
	printBanner()
	if Host == "" {
		Host = "localhost"
	}
	if Port == 0 {
		Port = 3040
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Error getting home directory", slog.String("error", err.Error()))
		os.Exit(1)
	}

	coolDir := filepath.Join(homeDir, "cooldb")
	if _, err := os.Stat(coolDir); os.IsNotExist(err) {
		err = os.Mkdir(coolDir, 0755)
		if err != nil {
			slog.Error("Error creating directory", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	dirFD, err := os.Open(coolDir)
	if err != nil {
		slog.Error("Error opening directory", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dirFD.Close()

	server := core.NewCoreServer(Host, Port, dirFD)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := core.BindAndListen(ctx, server); err != nil {
			slog.Error("Server encountered an error", slog.String("error", err.Error()))
		}
	}()

	slog.Info("Server started", slog.String("host", server.Host), slog.Int("port", server.Port))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	slog.Info("Shutting down server...")
	cancel()
	wg.Wait()
	slog.Info("Server shut down gracefully")
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
