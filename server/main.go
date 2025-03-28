package server

import (
	"fmt"
	"os"
	"path/filepath"
)

func Start() {
	printBanner()
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
