package cmd

import (
	"fmt"
	"os"

	gonanoid "github.com/matoous/go-nanoid/v2"
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
		if len(args) == 1 {
			file = args[0]
			if !doesFileExist(file) {
				fmt.Printf("File %s does not exist", file)
				fmt.Println()
				fmt.Printf("Creating new CoolDB file %s", file)
				if err := createNewDbFile(file + ".db"); err != nil {
					fmt.Printf("Error creating new file %s", err)
					os.Exit(1)
				}
			}

		} else if len(args) == 0 {
			id, err := gonanoid.New()
			if err != nil {
				fmt.Println("Error Creating CoolDB file")
			}
			fileName := fmt.Sprintf("%s_%s.db", defaultDbFilePrefix, id)
			if err := createNewDbFile(fileName); err != nil {
				fmt.Println("Error Creating CoolDB file")
			}
		} else {
			fmt.Println("Invalid arguments")
		}
		server.Start(file)
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
	f.Close()
	return nil
}
