package cmd

import (
	"fmt"
	"os"

	"github.com/rushikeshg25/coolDb/server"
	"github.com/spf13/cobra"
)

var (
	Version   = "0.1.0"
	BuildTime = "unknown"
)

type serverStarter func(host string, port int, databasePath string) error

// NewRootCommand constructs an isolated command tree for the CoolDB binary.
func NewRootCommand() *cobra.Command {
	return newRootCommand(server.Start)
}

func newRootCommand(startServer serverStarter) *cobra.Command {
	root := &cobra.Command{
		Use:   "cool",
		Short: "cooldb is a SQL-based database for storing cool stuff.",
		Long:  `cooldb is a SQL-based database for storing cool stuff, built with Go. Available at https://github.com/rushikeshg25/cool-db.`,
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			return command.Help()
		},
	}
	root.AddCommand(newStartCommand(startServer), newVersionCommand())
	return root
}

func newStartCommand(startServer serverStarter) *cobra.Command {
	var (
		port         int
		host         string
		databasePath string
		wal          bool
	)
	command := &cobra.Command{
		Use:   "start",
		Short: "Starts CoolDB server",
		Long:  `Starts CoolDB server`,
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			if wal {
				return fmt.Errorf("write-ahead logging is not implemented yet; omit --wal")
			}
			return startServer(host, port, databasePath)
		},
	}
	command.Flags().IntVarP(&port, "port", "p", 3040, "Port to run CoolDB server on")
	command.Flags().StringVar(&host, "host", "localhost", "Host to run CoolDB server on")
	command.Flags().StringVar(&databasePath, "db", "", "Database file (default: ~/cooldb/default.cooldb)")
	command.Flags().BoolVarP(&wal, "wal", "w", false, "Enable write-ahead logging (not available in v0.1)")
	return command
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CoolDB version",
		Args:  cobra.NoArgs,
		Run: func(command *cobra.Command, args []string) {
			fmt.Fprintf(command.OutOrStdout(), "CoolDB %s (built %s)\n", Version, BuildTime)
		},
	}
}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
