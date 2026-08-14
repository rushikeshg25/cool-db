package shell

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/rushikeshg25/coolDb/internal/client"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  Input
	}{
		{input: "", want: Input{Kind: InputEmpty}},
		{input: "  SELECT * FROM users;  ", want: Input{Kind: InputQuery, Text: "SELECT * FROM users;"}},
		{input: "exit", want: Input{Kind: InputExit}},
		{input: ".QUIT", want: Input{Kind: InputExit}},
		{input: "\\q", want: Input{Kind: InputExit}},
		{input: ".tables", want: Input{Kind: InputMetaCommand, Text: ".tables"}},
	}
	for _, test := range tests {
		if got := Parse(test.input); got != test.want {
			t.Errorf("Parse(%q) = %#v, want %#v", test.input, got, test.want)
		}
	}
}

type fakeReader struct {
	lines []string
	errs  []error
	index int
}

func (r *fakeReader) Readline() (string, error) {
	if r.index >= len(r.lines) {
		return "", io.EOF
	}
	line := r.lines[r.index]
	var err error
	if r.index < len(r.errs) {
		err = r.errs[r.index]
	}
	r.index++
	return line, err
}

type fakeExecutor struct {
	queries []string
	results map[string]client.Result
	errors  map[string]error
}

func (e *fakeExecutor) Execute(ctx context.Context, query string) (client.Result, error) {
	e.queries = append(e.queries, query)
	return e.results[query], e.errors[query]
}

func TestRunnerExecutesQueriesAndRendersVisibleBehavior(t *testing.T) {
	reader := &fakeReader{lines: []string{"", "SELECT 1", "bad query", ".tables", "quit"}}
	executor := &fakeExecutor{
		results: map[string]client.Result{"SELECT 1": {Text: "one"}},
		errors:  map[string]error{"bad query": errors.New("syntax error")},
	}
	var output, errorOutput bytes.Buffer
	runner := Runner{Executor: executor, Input: reader, Output: &output, Errors: &errorOutput}

	if err := runner.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !reflect.DeepEqual(executor.queries, []string{"SELECT 1", "bad query"}) {
		t.Errorf("queries = %#v", executor.queries)
	}
	if got, want := output.String(), "one\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
	if got, want := errorOutput.String(), "ERR syntax error\nERR unknown meta-command \".tables\"\n"; got != want {
		t.Errorf("error output = %q, want %q", got, want)
	}
}

func TestRunnerContinuesAfterInterruptedInput(t *testing.T) {
	reader := &fakeReader{lines: []string{"", "quit"}, errs: []error{ErrInterrupted}}
	runner := Runner{Executor: &fakeExecutor{}, Input: reader}
	if err := runner.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestRenderResultPreservesExistingNewline(t *testing.T) {
	var output bytes.Buffer
	RenderResult(&output, client.Result{Text: "row\n"})
	if got, want := output.String(), "row\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}
