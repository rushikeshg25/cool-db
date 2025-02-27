package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/rushikeshg25/coolDb/server"
	"github.com/spf13/cobra"
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
				log.Fatalf("File %s does not exist", file)
				os.Exit(1)
			}

		} else if len(args) == 0 {
			//Create a new Db file and start interactive cli
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
