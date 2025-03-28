package errors

import "fmt"

var (
	ErrUnknownCmd = func(cmd string) error {
		return fmt.Errorf("ERROR unknown command '%v'", cmd)
	}
	ErrUnknownCmdArg = func() error {
		return fmt.Errorf("ERROR unknown command argument")
	}
)
