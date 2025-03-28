package server

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

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
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}

	coolDir := filepath.Join(homeDir, "cooldb")

	if _, err := os.Stat(coolDir); os.IsNotExist(err) {
		err = os.Mkdir(coolDir, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			os.Exit(1)
		}
	}

	dirFD, err := os.Open(coolDir)
	if err != nil {
		fmt.Println("Error opening directory:", err)
		os.Exit(1)
	}
	slog.Info("Waiting for connections...")
	server := core.NewCoreServer(Host, Port)
	fmt.Println(server.Host, server.Port)
	defer dirFD.Close()
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
