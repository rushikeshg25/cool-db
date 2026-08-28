package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rushikeshg25/coolDb/server"
	"github.com/spf13/cobra"
)

func newServerCommand(runServer serverRunner) *cobra.Command {
	var (
		port         int
		httpPort     int
		host         string
		databasePath string
		wal          bool
	)
	command := &cobra.Command{
		Use:     "server",
		Aliases: []string{"start"},
		Short:   "Start the CoolDB server",
		Args:    cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			if wal {
				return fmt.Errorf("write-ahead logging is not implemented yet; omit --wal")
			}
			ctx, stop := signal.NotifyContext(command.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runServer(ctx, server.Config{
				Host:         host,
				Port:         port,
				HTTPPort:     httpPort,
				DatabasePath: databasePath,
				Output:       command.OutOrStdout(),
			})
		},
	}
	command.Flags().IntVarP(&port, "port", "p", 3040, "Port to run CoolDB server on")
	command.Flags().IntVar(&httpPort, "http-port", 0, "Local demo HTTP port (0 disables the demo API)")
	command.Flags().StringVar(&host, "host", "localhost", "Host to run CoolDB server on")
	command.Flags().StringVar(&databasePath, "db", "", "Database file (default: ~/cooldb/default.cooldb)")
	command.Flags().BoolVarP(&wal, "wal", "w", false, "Enable write-ahead logging (not available in v0.1)")
	return command
}
