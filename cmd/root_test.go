package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/rushikeshg25/coolDb/internal/client"
	"github.com/rushikeshg25/coolDb/server"
	"github.com/spf13/cobra"
)

func executeCommand(t *testing.T, rootCommandFactory func() *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var output bytes.Buffer
	rootCommand := rootCommandFactory()
	rootCommand.SetOut(&output)
	rootCommand.SetErr(&output)
	rootCommand.SetArgs(args)
	err := rootCommand.Execute()
	return output.String(), err
}

func TestRootShowsHelp(t *testing.T) {
	output, err := executeCommand(t, NewRootCommand)
	if err != nil {
		t.Fatalf("root command error = %v", err)
	}
	for _, expected := range []string{"cooldb is a SQL-based database", "server", "shell", "exec", "version"} {
		if !strings.Contains(output, expected) {
			t.Errorf("root output = %q, want it to contain %q", output, expected)
		}
	}
}

func TestVersionCommand(t *testing.T) {
	oldVersion, oldBuildTime := Version, BuildTime
	Version, BuildTime = "1.2.3", "test-time"
	t.Cleanup(func() { Version, BuildTime = oldVersion, oldBuildTime })

	output, err := executeCommand(t, NewRootCommand, "version")
	if err != nil {
		t.Fatalf("version command error = %v", err)
	}
	if got, want := output, "CoolDB 1.2.3 (built test-time)\n"; got != want {
		t.Errorf("version output = %q, want %q", got, want)
	}
}

func TestStartRejectsWAL(t *testing.T) {
	_, err := executeCommand(t, NewRootCommand, "start", "--wal")
	if err == nil {
		t.Fatal("start --wal error = nil")
	}
	if !strings.Contains(err.Error(), "write-ahead logging is not implemented") {
		t.Errorf("error = %q, want WAL guidance", err)
	}
}

func TestStartPassesConfigurationToServer(t *testing.T) {
	var gotHost, gotPath string
	var gotPort, gotHTTPPort int
	factory := func() *cobra.Command {
		return newRootCommand(dependencies{
			runServer: func(ctx context.Context, config server.Config) error {
				gotHost, gotPort, gotHTTPPort, gotPath = config.Host, config.Port, config.HTTPPort, config.DatabasePath
				return nil
			},
			connectClient: func(context.Context, client.Config) (managedClient, error) {
				return nil, nil
			},
		})
	}

	for _, commandName := range []string{"server", "start"} {
		if _, err := executeCommand(t, factory, commandName, "--host", "127.0.0.1", "--port", "4040", "--http-port", "4041", "--db", "test.cooldb"); err != nil {
			t.Fatalf("%s command error = %v", commandName, err)
		}
	}
	if gotHost != "127.0.0.1" || gotPort != 4040 || gotHTTPPort != 4041 || gotPath != "test.cooldb" {
		t.Errorf("server config = (%q, %d, %d, %q)", gotHost, gotPort, gotHTTPPort, gotPath)
	}
}
