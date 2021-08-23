// Package command provides an organized way of running a *exec.Cmd
// and generates a Capture that contains all commonly-accessed information
// in the Cmd that allows for repeated execution.
package command

import (
	"errors"
	"os/exec"
)

// ErrToExitCode converts potential errors to a nil-able int error code.
func ErrToExitCode(err error) int {
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}

	return 0
}
