package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

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
	for _, expected := range []string{"cooldb is a SQL-based database", "start", "version"} {
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
	output, err := executeCommand(t, NewRootCommand, "start", "--wal")
	if err == nil {
		t.Fatal("start --wal error = nil")
	}
	if !strings.Contains(err.Error(), "write-ahead logging is not implemented") {
		t.Errorf("error = %q, want WAL guidance", err)
	}
	if !strings.Contains(output, "Usage:") {
		t.Errorf("output = %q, want command usage", output)
	}
}

func TestStartPassesConfigurationToServer(t *testing.T) {
	var gotHost, gotPath string
	var gotPort int
	factory := func() *cobra.Command {
		return newRootCommand(func(ctx context.Context, config server.Config) error {
			gotHost, gotPort, gotPath = config.Host, config.Port, config.DatabasePath
			return nil
		})
	}

	if _, err := executeCommand(t, factory, "start", "--host", "127.0.0.1", "--port", "4040", "--db", "test.cooldb"); err != nil {
		t.Fatalf("start command error = %v", err)
	}
	if gotHost != "127.0.0.1" || gotPort != 4040 || gotPath != "test.cooldb" {
		t.Errorf("server config = (%q, %d, %q)", gotHost, gotPort, gotPath)
	}
}
