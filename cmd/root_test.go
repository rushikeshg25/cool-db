package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var output bytes.Buffer
	rootCmd.SetOut(&output)
	rootCmd.SetErr(&output)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return output.String(), err
}

func TestRootShowsHelp(t *testing.T) {
	output, err := executeCommand(t)
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

	output, err := executeCommand(t, "version")
	if err != nil {
		t.Fatalf("version command error = %v", err)
	}
	if got, want := output, "CoolDB 1.2.3 (built test-time)\n"; got != want {
		t.Errorf("version output = %q, want %q", got, want)
	}
}

func TestStartRejectsWAL(t *testing.T) {
	output, err := executeCommand(t, "start", "--wal")
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
