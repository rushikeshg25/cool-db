package cmd

import (
	"encoding/binary"
	"fmt"
	"os"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rushikeshg25/coolDb/internal"
	"github.com/rushikeshg25/coolDb/server"
	"github.com/spf13/cobra"
)

const (
	defaultDbFilePrefix = "cooldb"
)

var VersionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("CoolDB version %s\n", Version)
	},
}

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the cooldb server",
	Run: func(cmd *cobra.Command, args []string) {
		var file string
		var filePath string
		if len(args) == 1 {
			file = args[0]
			filePath = os.Getenv("PWD") + "/" + file
			if !doesFileExist(file) {
				fmt.Printf("File %s does not exist\n", file)
				fmt.Printf("Creating new CoolDB file %s\n", file)
				if err := createNewDbFile(file); err != nil {
					fmt.Printf("Error creating new file %s\n", err)
					os.Exit(1)
				}
			}
		} else if len(args) == 0 {
			id, err := gonanoid.New()
			if err != nil {
				fmt.Println("Error Creating CoolDB file")
				os.Exit(1)
			}
			file = fmt.Sprintf("%s_%s", defaultDbFilePrefix, id)
			filePath = os.Getenv("PWD") + "/" + file
			if err := createNewDbFile(file); err != nil {
				fmt.Println("Error Creating CoolDB file")
				os.Exit(1)
			}
		} else {
			fmt.Println("Invalid arguments")
			os.Exit(1)
		}
		server.Start(file, filePath)
	},
}

func doesFileExist(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}

func createNewDbFile(fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	DbFileHeader := internal.InitFileConfig()
	if err := binary.Write(f, binary.BigEndian, DbFileHeader); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}
