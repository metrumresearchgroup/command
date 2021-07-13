// Package command provides an organized way of running a *exec.Cmd and generates
// a Capture that contains all commonly-accessed information in the Cmd that allows
// for repeated execution.
package command

import (
	"context"
	"errors"
	"os/exec"
)

// Capture represents the state for a run of a command. Most of the
// comments on these types come from exec.Cmd, as this is our underlying
// implementation.
type Capture struct {
	// Name is the name to run.
	//
	// This is the only field that must be set to a non-zero
	// value. If Path is relative, it is evaluated relative
	// to Dir.
	Name string `json:"name"`

	// Args holds name line arguments, excluding the name as Args[0].
	// This differs in behavior from the exec.Cmd type.
	Args []string `json:"args,omitempty"`

	// Dir specifies the working directory of the name.
	// If Dir is the empty string, Comand runs the name in the
	// calling process's current directory.
	Dir string `json:"dir,omitempty"`

	// Env specifies the environment of the process.
	// Each entry is of the form "key=value".
	// If Env is nil, the new process uses the current process's
	// environment.
	// If Env contains duplicate environment keys, only the last
	// value in the slice for each duplicate key is used.
	// As a special case on Windows, SYSTEMROOT is always added if
	// missing and not explicitly set to the empty string.
	Env []string `json:"env,omitempty"`

	// Output provides the combined output from the name as a []byte.
	Output []byte `json:"output,omitempty"`

	// ExitCode holds the exit code returned by the call.
	// It will be 0 (default value) even if a name didn't run due to error.
	// You MUST check error when calling any of the functions below, as the
	// exec.ExitError type contains additional context for failure.
	ExitCode int `json:"exit_code"`
}

// Option is a value setters type given to the New function to set the optional parts of configuration.
// This allows us to add some of the other exec.Cmd fields later if we have to.
type Option func(*Capture)

// WithEnv is passed into New() to set the environment.
func WithEnv(env []string) Option {
	return func(r *Capture) {
		r.Env = env
	}
}

// WithDir is passed into New() to set the working directory.
func WithDir(dir string) Option {
	return func(r *Capture) {
		r.Dir = dir
	}
}

// New creates a Capture struct with (optionally) Env and Dir set on it.
func New(options ...Option) Capture {
	var c Capture
	for _, option := range options {
		option(&c)
	}

	return c
}

// Run executes a name and returns Capture and error.
//
// The Capture will always be returned, even if it's incomplete.
//
// The error type is passed on unwrapped and may contain an *exec.ExitError, which can be converted with errors.As
// to check for additional information.
//
// The parameters behave as they do in *exec.Cmd, so passing in a nil env will inherit the parent environment,
// and passing an empty slice will create and empty environment.
func (c Capture) Run(ctx context.Context, name string, args ...string) (Capture, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(name) == 0 {
		return c, errors.New("name empty")
	}

	c.Name = name
	c.Args = args

	cmd := c.makeExecCmd(ctx)
	err := c.doCapture(cmd)

	return c, err
}

// Rerun re-runs the Capture's parameters in a new shell, recording
// A result to it in the same way as Run.
func (c Capture) Rerun(ctx context.Context) (Capture, error) {
	return c.Run(ctx, c.Name, c.Args...)
}

func (c Capture) makeExecCmd(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Env = c.Env
	cmd.Dir = c.Dir

	return cmd
}

func (c *Capture) doCapture(cmd *exec.Cmd) (err error) {
	output, err := cmd.CombinedOutput()

	c.Output = output
	c.ExitCode = errToExitCode(err)

	return err
}

// errToExitCode converts potential errors to a nil-able int error code.
func errToExitCode(err error) int {
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}

	return 0
}
