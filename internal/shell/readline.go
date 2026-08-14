package shell

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chzyer/readline"
)

type Readline struct {
	instance *readline.Instance
}

func NewReadline(prompt string) (*Readline, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home directory: %w", err)
	}
	instance, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:     filepath.Join(homeDirectory, ".cooldb_history"),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, fmt.Errorf("initialize terminal: %w", err)
	}
	return &Readline{instance: instance}, nil
}

func (r *Readline) Readline() (string, error) {
	line, err := r.instance.Readline()
	if errors.Is(err, readline.ErrInterrupt) {
		return line, ErrInterrupted
	}
	return line, err
}

func (r *Readline) Close() error {
	return r.instance.Close()
}
