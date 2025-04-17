package cmd

import (
	"log"
	"os"

	"github.com/rushikeshg25/coolDb/server"
	"github.com/spf13/cobra"
)

var (
	Version = "0.1.0"
	port    int
	host    string
	WAL     string
)

func init() {
	startCmd.Flags().IntVarP(&port, "port", "p", 3040, "Port to run CoolDB server on")
	startCmd.Flags().StringVarP(&host, "host", "", "localhost", "Host to run CoolDB server on")
	startCmd.Flags().StringVarP(&WAL, "wal", "w", "false", "Enable Work ahead logs")
	rootCmd.AddCommand(startCmd)
}

var rootCmd = &cobra.Command{
	Use:   "cool",
	Short: "cooldb is a SQL-based database for storing cool stuff.",
	Long:  `cooldb is a SQL-based database for storing cool stuff, built with Go. Available at https://github.com/rushikeshg25/cool-db.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
		os.Exit(1)
	}
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts CoolDB server",
	Long:  `Starts CoolDB server`,
	Run: func(cmd *cobra.Command, args []string) {
		server.Start(host, port, WAL)
	},
}
