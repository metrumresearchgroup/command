// Package command provides an organized way of running a *exec.Cmd and generates
// a Capture that contains all commonly-accessed information in the Cmd that allows
// for repeated execution.
package command

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"

	"github.com/metrumresearchgroup/command/pipes"
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

	// ExitCode holds the exit code returned by the call.
	// It will be 0 (default value) even if a name didn't run due to error.
	// You MUST check error when calling any of the functions below, as the
	// exec.ExitError type contains additional context for failure.
	ExitCode int `json:"exit_code"`

	// tmpCommand stores the running command in the interactive Start style.
	tmpCmd *exec.Cmd

	// tmpCancel is the cancel func for the context. This will cancel a clone
	// of the original context, which gives us upper cancels a chance to run out
	// their lifecycle.
	tmpCancel func()
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
func New(options ...Option) *Capture {
	var c Capture
	for _, option := range options {
		option(&c)
	}

	return &c
}

// Run executes a name with args and returns an error.
//
// The error type is passed on unwrapped and may contain an *exec.ExitError, which can be converted with errors.As
// to check for additional information.
//
// The parameters behave as they do in *exec.Cmd, so passing in a nil env will inherit the parent environment,
// and passing an empty slice will create and empty environment.
func (c *Capture) Run(ctx context.Context, name string, args ...string) (*pipes.Pipes, error) {
	c.Name = name
	c.Args = args

	cmd := c.makeExecCmd(ctx)

	return c.doRun(cmd)
}

func (c *Capture) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	c.Name = name
	c.Args = args

	cmd := c.makeExecCmd(ctx)

	return cmd.CombinedOutput()
}

// Start starts a exec.Cmd with stdin/out/err set to the values in io, a name, and args, and returns Capture and Error
// indicating whether the start was successful, and to provide with re-runnable operation.
//
// The results are similar to Run, where an *exec.ExitError will fire in the case of failure to run.
func (c *Capture) Start(ctx context.Context, name string, args ...string) (*pipes.Pipes, error) {
	c.Name = name
	c.Args = args

	// This inner cancel seems redundant but it doesn't notify upward. This
	// allows upper lifecycles to finish.
	ctx, c.tmpCancel = context.WithCancel(ctx)

	cmd := c.makeExecCmd(ctx)
	p, err := c.doStart(cmd)
	if err != nil {
		return p, err
	}

	c.tmpCmd = cmd

	return p, nil
}

func (c *Capture) doStart(cmd *exec.Cmd) (*pipes.Pipes, error) {
	p, err := pipes.Attach(cmd)
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		c.ExitCode = errToExitCode(err)
	}

	return p, err
}

// Restart runs a previously run command, binding pipes before operation.
func (c *Capture) Restart(ctx context.Context) (*pipes.Pipes, error) {
	return c.Start(ctx, c.Name, c.Args...)
}

// Stop ends a Started capture.
func (c *Capture) Stop() error {
	if c.tmpCancel == nil || c.tmpCmd == nil {
		return errors.New("command was not started")
	}

	err := c.doStop()

	c.ExitCode = errToExitCode(err)

	return err
}

func (c *Capture) doStop() error {
	if c.tmpCancel != nil {
		c.tmpCancel()
	}

	timeout := time.NewTimer(time.Second * 10)
	ticker := time.NewTicker(time.Second)

	if c.tmpCmd == nil {
		return errors.New("wasn't running")
	}

	if c.tmpCmd.ProcessState == nil || c.tmpCmd.ProcessState.Exited() {
		if c.tmpCmd.ProcessState != nil {
			c.ExitCode = c.tmpCmd.ProcessState.ExitCode()
		}

		c.tmpCmd = nil

		return nil // process is stopped
	}

	if c.tmpCmd.ProcessState.Exited() {
		return nil
	} else {
		err := c.tmpCmd.Process.Kill()
		if err != nil {
			return err
		}
	}

	for {
		select {
		case <-timeout.C:
			ticker.Stop()

			return errors.New("timeout reached")
		case <-ticker.C:
			if !c.tmpCmd.ProcessState.Exited() {
				ticker.Stop()
				timeout.Stop()

				continue
			}

			return c.tmpCmd.Wait()
		}
	}
}

// Rerun re-runs the Capture's parameters in a new shell, recording
// A result to it in the same way as Run.
func (c *Capture) Rerun(ctx context.Context) (*pipes.Pipes, error) {
	return c.Run(ctx, c.Name, c.Args...)
}

func (c *Capture) makeExecCmd(ctx context.Context) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Env = c.Env
	cmd.Dir = c.Dir

	return cmd
}

func (c *Capture) doRun(cmd *exec.Cmd) (*pipes.Pipes, error) {
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	p := pipes.Output(outBuf, errBuf)

	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	err := cmd.Run()
	c.ExitCode = errToExitCode(err)

	return p, err
}

// errToExitCode converts potential errors to a nil-able int error code.
func errToExitCode(err error) int {
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}

	return 0
}
