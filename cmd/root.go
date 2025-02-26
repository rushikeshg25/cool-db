package cmd

import (
	"log"
	"os"

	"github.com/rushikeshg25/coolDb/server"
	"github.com/spf13/cobra"
)

// func init() {
// 	rootCmd.AddCommand(startCmd)
// 	rootCmd.AddCommand(versionCmd)
// 	rootCmd.AddCommand(helpCmd)
// 	rootCmd.AddCommand(quitCmd)
// 	rootCmd.AddCommand(dbinfoCmd)
// 	rootCmd.AddCommand(openCmd)
// 	rootCmd.AddCommand(closeCmd)
// }

var rootCmd = &cobra.Command{
	Use:   "cooldb",
	Short: "cooldb is a SQLite based database for storing cool stuff.",
	Long:  `cooldb is a SQLite based database for storing cool stuff built with Go available at https://github.com/rushikeshg25/cool-db.`,
	Run: func(cmd *cobra.Command, args []string) {
		server.Start()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command:%v", err)
		os.Exit(1)
	}
}
