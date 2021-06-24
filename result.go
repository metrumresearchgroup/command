// Package command provides an organized way of running a *exec.Cmd and generates a Result that contains all
// commonly-accessed information in the Cmd that allows for repeated execution.
package command

import (
	"context"
	"errors"
	"os/exec"
	"path"
)

// Result represents a single execution of a *exec.Cmd
type Result struct {
	// Name is the command to run, similar to exec.Command's first parameter
	Name string `json:"name"`

	// Args contains the command line called, without argv[0].
	Args []string `json:"args,omitempty"`

	// Env contains the environment passed into CaptureContext.
	Env []string `json:"env,omitempty"`

	// Output provides the combined output from the command as a string.
	Output string `json:"output,omitempty"`

	// ExitCode holds the exit code returned by the call.
	// It will be 0 (default value) even if a command didn't run due to error.
	// You MUST check error when calling any of the functions below, as the
	// exec.ExitError type contains additional context for failure.
	ExitCode int `json:"exitCode"`
}

// CaptureContext executes an exec.CommandContext and returns Result and error.
//
// The Result will always be returned, even if it's incomplete.
//
// The error type is passed on unwrapped and may contain an *exec.ExitError, which can be converted with errors.As
// to check for additional information.
//
// The parameters behave as they do in *exec.Cmd, so passing in a nil env will inherit the parent environment,
// and passing an empty slice will create and empty environment.
func CaptureContext(ctx context.Context, env []string, name string, args ...string) (cr Result, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = env

	return capture(cmd)
}

// Capture executes an exec.CommandContext and captures the output as a Result and an error. It has no control over
// the context.Context. If you need to inject context.Context, use CaptureContext
func Capture(env []string, name string, args ...string) (cr Result, err error) {
	return CaptureContext(context.Background(), env, name, args...)
}

func capture(cmd *exec.Cmd) (cr Result, err error) {
	output, err := cmd.CombinedOutput()

	cr = Result{
		Name:     path.Base(cmd.Path),
		Args:     cmd.Args[1:],
		Env:      cmd.Env,
		Output:   string(output),
		ExitCode: errToExitCode(err),
	}

	return cr, err
}

// errToExitCode converts potential errors to a nil-able int error code.
func errToExitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}

	return 0
}

// CaptureContext re-runs the Result's parameters in a new shell, recording
// A result in the same way as CaptureContext.
func (cr Result) CaptureContext(ctx context.Context) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// suppress:G204
	cmd := exec.CommandContext(ctx, cr.Name, cr.Args...)
	cmd.Env = cr.Env

	return capture(cmd)
}

// Capture is the same as CaptureContext without regard for controlling context.Context.
func (cr Result) Capture() (Result, error) {
	return cr.CaptureContext(context.Background())
}
