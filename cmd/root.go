package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/rushikeshg25/coolDb/internal/client"
	"github.com/rushikeshg25/coolDb/internal/shell"
	"github.com/rushikeshg25/coolDb/server"
	"github.com/spf13/cobra"
)

var (
	Version   = "0.1.0"
	BuildTime = "unknown"
)

type serverRunner func(context.Context, server.Config) error

type managedClient interface {
	shell.Executor
	Close() error
}

type clientConnector func(context.Context, client.Config) (managedClient, error)

type interactiveInput interface {
	shell.LineReader
	Close() error
}

type inputFactory func(prompt string) (interactiveInput, error)

type dependencies struct {
	runServer     serverRunner
	connectClient clientConnector
	newInput      inputFactory
}

func productionDependencies() dependencies {
	return dependencies{
		runServer: server.Run,
		connectClient: func(ctx context.Context, config client.Config) (managedClient, error) {
			return client.Connect(ctx, config)
		},
		newInput: func(prompt string) (interactiveInput, error) {
			return shell.NewReadline(prompt)
		},
	}
}

// NewRootCommand constructs an isolated command tree for the CoolDB binary.
func NewRootCommand() *cobra.Command {
	return newRootCommand(productionDependencies())
}

func newRootCommand(deps dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:          "cool",
		Short:        "cooldb is a SQL-based database for storing cool stuff.",
		Long:         `cooldb is a SQL-based database for storing cool stuff, built with Go. Available at https://github.com/rushikeshg25/cool-db.`,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			return command.Help()
		},
	}
	root.AddCommand(
		newServerCommand(deps.runServer),
		newShellCommand(deps.connectClient, deps.newInput),
		newExecCommand(deps.connectClient),
		newVersionCommand(),
	)
	return root
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
