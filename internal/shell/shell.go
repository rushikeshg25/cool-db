package shell

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rushikeshg25/coolDb/internal/client"
)

var ErrInterrupted = errors.New("input interrupted")

type Executor interface {
	Execute(context.Context, string) (client.Result, error)
}

type LineReader interface {
	Readline() (string, error)
}

type InputKind uint8

const (
	InputEmpty InputKind = iota
	InputQuery
	InputExit
	InputMetaCommand
)

type Input struct {
	Kind InputKind
	Text string
}

func Parse(input string) Input {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return Input{Kind: InputEmpty}
	}
	switch strings.ToLower(trimmed) {
	case "exit", "quit", ".exit", ".quit", "\\q":
		return Input{Kind: InputExit}
	}
	if strings.HasPrefix(trimmed, ".") || strings.HasPrefix(trimmed, "\\") {
		return Input{Kind: InputMetaCommand, Text: trimmed}
	}
	return Input{Kind: InputQuery, Text: trimmed}
}

type Runner struct {
	Executor Executor
	Input    LineReader
	Output   io.Writer
	Errors   io.Writer
}

func (r Runner) Run(ctx context.Context) error {
	if r.Executor == nil || r.Input == nil {
		return fmt.Errorf("shell requires an executor and input reader")
	}
	if r.Output == nil {
		r.Output = io.Discard
	}
	if r.Errors == nil {
		r.Errors = io.Discard
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		line, err := r.Input.Readline()
		if errors.Is(err, ErrInterrupted) {
			continue
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}

		input := Parse(line)
		switch input.Kind {
		case InputEmpty:
			continue
		case InputExit:
			return nil
		case InputMetaCommand:
			fmt.Fprintf(r.Errors, "ERR unknown meta-command %q\n", input.Text)
			continue
		case InputQuery:
			queryContext, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			result, err := r.Executor.Execute(queryContext, input.Text)
			stop()
			if err != nil {
				RenderError(r.Errors, err)
				continue
			}
			RenderResult(r.Output, result)
		}
	}
}

func RenderResult(writer io.Writer, result client.Result) {
	if writer == nil || result.Text == "" {
		return
	}
	fmt.Fprint(writer, result.Text)
	if !strings.HasSuffix(result.Text, "\n") {
		fmt.Fprintln(writer)
	}
}

func RenderError(writer io.Writer, err error) {
	if writer == nil || err == nil {
		return
	}
	fmt.Fprintf(writer, "ERR %s\n", err)
}
