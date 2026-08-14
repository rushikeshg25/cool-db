package cmd

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/rushikeshg25/coolDb/internal/client"
	"github.com/rushikeshg25/coolDb/internal/shell"
	"github.com/spf13/cobra"
)

type connectionFlags struct {
	host              string
	port              int
	connectionTimeout time.Duration
	queryTimeout      time.Duration
}

func (f *connectionFlags) bind(command *cobra.Command) {
	command.Flags().StringVar(&f.host, "host", "localhost", "Hostname or IP address of the CoolDB server")
	command.Flags().IntVarP(&f.port, "port", "p", 3040, "Port number of the CoolDB server")
	command.Flags().DurationVar(&f.connectionTimeout, "connect-timeout", 5*time.Second, "Time allowed to connect to the server")
	command.Flags().DurationVar(&f.queryTimeout, "query-timeout", 30*time.Second, "Maximum duration of each query")
}

func (f connectionFlags) config() client.Config {
	return client.Config{
		Host:              f.host,
		Port:              f.port,
		ConnectionTimeout: f.connectionTimeout,
		QueryTimeout:      f.queryTimeout,
	}
}

func newShellCommand(connectClient clientConnector, newInput inputFactory) *cobra.Command {
	var connection connectionFlags
	command := &cobra.Command{
		Use:   "shell",
		Short: "Open an interactive SQL shell",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			databaseClient, err := connectClient(command.Context(), connection.config())
			if err != nil {
				return err
			}
			defer databaseClient.Close()

			prompt := fmt.Sprintf("cool@%s:%d> ", connection.host, connection.port)
			input, err := newInput(prompt)
			if err != nil {
				return err
			}
			defer input.Close()

			return (shell.Runner{
				Executor: databaseClient,
				Input:    input,
				Output:   command.OutOrStdout(),
				Errors:   command.ErrOrStderr(),
			}).Run(command.Context())
		},
	}
	connection.bind(command)
	return command
}

func newExecCommand(connectClient clientConnector) *cobra.Command {
	var connection connectionFlags
	command := &cobra.Command{
		Use:   "exec [query]",
		Short: "Execute one SQL query",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			query, err := queryFromInput(command, args)
			if err != nil {
				return err
			}
			databaseClient, err := connectClient(command.Context(), connection.config())
			if err != nil {
				return err
			}
			defer databaseClient.Close()

			result, err := databaseClient.Execute(command.Context(), query)
			if err != nil {
				return err
			}
			shell.RenderResult(command.OutOrStdout(), result)
			return nil
		},
	}
	connection.bind(command)
	return command
}

func queryFromInput(command *cobra.Command, args []string) (string, error) {
	if len(args) == 1 {
		query := strings.TrimSpace(args[0])
		if query == "" {
			return "", fmt.Errorf("query cannot be empty")
		}
		return query, nil
	}
	payload, err := io.ReadAll(command.InOrStdin())
	if err != nil {
		return "", fmt.Errorf("read query from stdin: %w", err)
	}
	query := strings.TrimSpace(string(payload))
	if query == "" {
		return "", fmt.Errorf("provide a query argument or pipe a query to stdin")
	}
	return query, nil
}
