package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "0.1.0"
)

func init() {
	rootCmd.AddCommand(VersionCmd)
	rootCmd.AddCommand(RunCmd)
}

var rootCmd = &cobra.Command{
	Use:   "cool",
	Short: "cooldb is a SQLite based database for storing cool stuff.",
	Long:  `cooldb is a SQLite based database for storing cool stuff built with Go available at https://github.com/rushikeshg25/cool-db.`,
	Run:   nil,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command:%v", err)
		os.Exit(1)
	}
}
