package cmd

import (
	"fmt"
	"os"

	"github.com/rushikeshg25/coolDb/server"
	"github.com/spf13/cobra"
)

var (
	Version      = "0.1.0"
	BuildTime    = "unknown"
	port         int
	host         string
	databasePath string
	wal          bool
)

func init() {
	startCmd.Flags().IntVarP(&port, "port", "p", 3040, "Port to run CoolDB server on")
	startCmd.Flags().StringVarP(&host, "host", "", "localhost", "Host to run CoolDB server on")
	startCmd.Flags().StringVar(&databasePath, "db", "", "Database file (default: ~/cooldb/default.cooldb)")
	startCmd.Flags().BoolVarP(&wal, "wal", "w", false, "Enable write-ahead logging (not available in v0.1)")
	rootCmd.AddCommand(startCmd, versionCmd)
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
		os.Exit(1)
	}
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts CoolDB server",
	Long:  `Starts CoolDB server`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if wal {
			return fmt.Errorf("write-ahead logging is not implemented yet; omit --wal")
		}
		return server.Start(host, port, databasePath)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CoolDB version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "CoolDB %s (built %s)\n", Version, BuildTime)
	},
}
