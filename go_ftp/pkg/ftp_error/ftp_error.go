package ftp_error

import (
	"fmt"
)

type NotImplementedError struct {
	Cmd string
}

func (e *NotImplementedError) Error() string {
	return fmt.Sprintf("Command: %s not implemented", e.Cmd)
}

type ExitError struct{}

func (e *ExitError) Error() string {
	return fmt.Sprintf("Good Bye")
}

type InvalidCommandError struct {
	Cmd string
}

func (e *InvalidCommandError) Error() string {
	return fmt.Sprintf("Invalid Command: %s", e.Cmd)
}

type NoArgumentError struct {
	Cmd string
}

func (e *NoArgumentError) Error() string {
	return fmt.Sprintf("No argument for command %s", e.Cmd)
}

type FileNotFoundError struct {
	File string
}

func (e *FileNotFoundError) Error() string {
	return fmt.Sprintf("No argument for command %s", e.File)
}
