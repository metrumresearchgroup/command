// Package command provides an organized way of running a *exec.Cmd
// and generates a Capture that contains all commonly-accessed information
// in the Cmd that allows for repeated execution.
package command

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"
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
	// If Dir is the empty string, Command runs the name in the
	// calling process's current directory.
	Dir string `json:"dir,omitempty"`

	// Env specifies the environment of the process.
	// Each entry is of the form "key=value".
	//
	// If Env is nil, the new process uses the current process's
	// environment.
	//
	// If Env contains duplicate environment keys, only the last
	// value in the slice for each duplicate key is used.
	//
	// As a special case on Windows, SYSTEMROOT is always added if
	// missing and not explicitly set to the empty string.
	Env []string `json:"env,omitempty"`

	// ExitCode holds the exit code returned by the call.
	// It will be 0 (default value) even if a name didn't run due to
	// error. You MUST check error when calling any of the functions
	// below, as the exec.ExitError type contains additional context
	// for failure.
	ExitCode int `json:"exit_code"`

	// cmdModifierFunc is for a user-supplied function to change the
	// properties of the *exec.Cmd before it is called. This allows
	// the end user to manipulate inputs or set properties otherwise
	// hidden by this library.
	cmdModifierFunc func(*exec.Cmd) error

	// tmpCommand stores the running command in the interactive Start
	// style.
	tmpCmd *exec.Cmd

	// tmpCancel is the cancel func for the context. This will cancel
	// a clone of the original context, which gives us upper cancels
	// a chance to run out their lifecycle.
	tmpCancel func()
}

// Option is a value setters type given to the New function to set
// the optional parts of configuration. This allows us to add some of
// the other exec.Cmd fields later if we have to.
type Option func(*Capture)

// WithCmdModifier is passed into New() to allow modification of the exec.command
// after all other settings are configured but before running.
//
// This allows for adjustments to the exec.Cmd before executing it.
//
// You can modify any public property of the Cmd. Remember that this is a
// "foot-gun" and can lead to unexpected behavior.
func WithCmdModifier(fn func(*exec.Cmd) error) Option {
	return func(c *Capture) {
		c.cmdModifierFunc = fn
	}
}

// WithEnv is passed into New() to set the environment.
func WithEnv(env []string) Option {
	return func(c *Capture) {
		c.Env = env
	}
}

// WithDir is passed into New() to set the working directory.
func WithDir(dir string) Option {
	return func(c *Capture) {
		c.Dir = dir
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
// The error may contain an *exec.ExitError, which can be converted with
// errors.As to check for additional information.
//
// The parameters behave as they do in *exec.Cmd, so passing in a nil
// env will inherit the parent environment, and passing an empty slice
// will create and empty environment.
func (c *Capture) Run(ctx context.Context, name string, args ...string) (*Pipes, error) {
	c.Name = name
	c.Args = args

	cmd, err := c.makeExecCmd(ctx)
	if err != nil {
		return nil, err
	}

	return c.doRun(cmd)
}

func (c *Capture) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	c.Name = name
	c.Args = args

	cmd, err := c.makeExecCmd(ctx)
	if err != nil {
		return nil, err
	}

	return cmd.CombinedOutput()
}

// Start starts a exec.Cmd with stdin/out/err set to the values in io,
// a name, and args, and returns Capture and Error indicating whether
// the start was successful, and to provide with re-runnable operation.
//
// The results are similar to Run, where an *exec.ExitError will fire,
// in the case of failure to run.
func (c *Capture) Start(ctx context.Context, name string, args ...string) (*Interact, error) {
	c.Name = name
	c.Args = args

	// This inner cancel seems redundant but it doesn't notify upward.
	// This allows upper lifecycles to finish.
	ctx, c.tmpCancel = context.WithCancel(ctx)

	cmd, err := c.makeExecCmd(ctx)
	if err != nil {
		return nil, err
	}

	i, err := c.doStart(cmd)
	if err != nil {
		return i, err
	}

	c.tmpCmd = cmd

	return i, nil
}

func (c *Capture) doStart(cmd *exec.Cmd) (*Interact, error) {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	// Create stdout, stderr streams of type io.Reader
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		c.ExitCode = errToExitCode(err)
	}

	return &Interact{
		Plumber: &Pipes{
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: stderr,
		},
		Controller: c,
	}, nil
}

func (c *Capture) Restart(ctx context.Context) (*Interact, error) {
	// Restart runs a previously run command, binding pipes before operation.
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

// Wait waits on the underlying process to wait; typically you'd do this
// if you already have a goroutine listening to outputs.
func (c *Capture) Wait() error {
	if c.tmpCancel == nil || c.tmpCmd == nil {
		return errors.New("command was not started")
	}

	return c.tmpCmd.Wait()
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

// With allows a post-creation addition/modification to the internal
// options of the Capture. This is especially useful for situations
// where you want to re-set the command modifier function after
// unmarshalling, since functions in go are not serializable.
func (c *Capture) With(options ...Option) {
	for _, option := range options {
		option(c)
	}
}

// Rerun re-runs the Capture's parameters in a new shell, recording
// A result to it in the same way as Run.
func (c *Capture) Rerun(ctx context.Context) (*Pipes, error) {
	return c.Run(ctx, c.Name, c.Args...)
}

func (c *Capture) makeExecCmd(ctx context.Context) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Env = c.Env
	cmd.Dir = c.Dir

	if c.cmdModifierFunc != nil {
		err := c.cmdModifierFunc(cmd)
		if err != nil {
			return nil, err
		}
	}

	return cmd, nil
}

func (c *Capture) doRun(cmd *exec.Cmd) (*Pipes, error) {
	bufOut, bufErr := &bytes.Buffer{}, &bytes.Buffer{}
	p := &Pipes{
		Stdin:  nil,
		Stdout: bufOut,
		Stderr: bufErr,
	}

	cmd.Stdout = bufOut
	cmd.Stderr = bufErr

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
