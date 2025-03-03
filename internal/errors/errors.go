package errors

import "fmt"

var (
	InvalidFileErr = func(fileName string) error {
		return fmt.Errorf("invalid file: not a valid coolDb file %s", fileName)
	}
	ErrUnknownCmd = func(cmd string) error {
		return fmt.Errorf("ERROR unknown command '%v'", cmd)
	}
	ErrUnknownCmdArg = func() error {
		return fmt.Errorf("ERROR unknown command argument")
	}
	ErrFileParse = func(fileName string) error {
		return fmt.Errorf("ERROR unable to parse file %s", fileName)
	}
)
