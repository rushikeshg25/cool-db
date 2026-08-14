package cmd

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/rushikeshg25/coolDb/internal/client"
	"github.com/rushikeshg25/coolDb/server"
	"github.com/spf13/cobra"
)

type fakeManagedClient struct {
	queries []string
	result  client.Result
	err     error
	closed  bool
}

func (c *fakeManagedClient) Execute(ctx context.Context, query string) (client.Result, error) {
	c.queries = append(c.queries, query)
	return c.result, c.err
}

func (c *fakeManagedClient) Close() error {
	c.closed = true
	return nil
}

type fakeInteractiveInput struct {
	lines  []string
	index  int
	closed bool
}

func (i *fakeInteractiveInput) Readline() (string, error) {
	if i.index >= len(i.lines) {
		return "", io.EOF
	}
	line := i.lines[i.index]
	i.index++
	return line, nil
}

func (i *fakeInteractiveInput) Close() error {
	i.closed = true
	return nil
}

func commandFactory(databaseClient *fakeManagedClient, input *fakeInteractiveInput, gotConfig *client.Config) func() *cobra.Command {
	return func() *cobra.Command {
		return newRootCommand(dependencies{
			runServer: func(context.Context, server.Config) error { return nil },
			connectClient: func(ctx context.Context, config client.Config) (managedClient, error) {
				if gotConfig != nil {
					*gotConfig = config
				}
				return databaseClient, nil
			},
			newInput: func(prompt string) (interactiveInput, error) { return input, nil },
		})
	}
}

func TestShellExecutesInteractiveQueries(t *testing.T) {
	databaseClient := &fakeManagedClient{result: client.Result{Text: "one"}}
	input := &fakeInteractiveInput{lines: []string{"SELECT 1", "quit"}}
	var config client.Config
	output, err := executeCommand(t, commandFactory(databaseClient, input, &config), "shell", "--host", "db.local", "--port", "4040")
	if err != nil {
		t.Fatalf("shell command error = %v", err)
	}
	if got, want := output, "one\n"; got != want {
		t.Errorf("shell output = %q, want %q", got, want)
	}
	if len(databaseClient.queries) != 1 || databaseClient.queries[0] != "SELECT 1" {
		t.Errorf("queries = %#v", databaseClient.queries)
	}
	if config.Host != "db.local" || config.Port != 4040 {
		t.Errorf("client config = %#v", config)
	}
	if !databaseClient.closed || !input.closed {
		t.Errorf("resources not closed: client=%t input=%t", databaseClient.closed, input.closed)
	}
}

func TestExecAcceptsArgument(t *testing.T) {
	databaseClient := &fakeManagedClient{result: client.Result{Text: "row\n"}}
	output, err := executeCommand(t, commandFactory(databaseClient, &fakeInteractiveInput{}, nil), "exec", "SELECT * FROM users")
	if err != nil {
		t.Fatalf("exec command error = %v", err)
	}
	if got, want := output, "row\n"; got != want {
		t.Errorf("exec output = %q, want %q", got, want)
	}
}

func TestExecAcceptsStdin(t *testing.T) {
	databaseClient := &fakeManagedClient{result: client.Result{Text: "created"}}
	factory := commandFactory(databaseClient, &fakeInteractiveInput{}, nil)
	root := factory()
	root.SetIn(strings.NewReader("CREATE TABLE users (id INTEGER)\n"))
	root.SetArgs([]string{"exec"})
	if err := root.Execute(); err != nil {
		t.Fatalf("exec command error = %v", err)
	}
	if len(databaseClient.queries) != 1 || databaseClient.queries[0] != "CREATE TABLE users (id INTEGER)" {
		t.Errorf("queries = %#v", databaseClient.queries)
	}
}
