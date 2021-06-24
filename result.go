package command

import (
	"context"
	"errors"
	"os/exec"
	"path"
)

// Result represents a single executable Diff. Output is combined output. If ExitError has a value, it will have
// ExitError.Stderr embedded for stderr output.
type Result struct {
	Name     string   `json:"name"`
	Args     []string `json:"args,omitempty"`
	Env      []string `json:"env,omitempty"`
	Output   string   `json:"output,omitempty"`
	ExitCode int      `json:"exitCode"`
}

// CaptureContext executes an exec.CommandContext and captures the output as a Result and an error.
// The Result will always be returned, even if it's incomplete.
func CaptureContext(ctx context.Context, env []string, name string, args ...string) (cr Result, err error) {
	if ctx == nil {
		ctx = context.TODO()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = env

	return capture(cmd)
}

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
// A result in the same way as command.CaptureContext.
func (cr Result) CaptureContext(ctx context.Context) (Result, error) {
	if ctx == nil {
		ctx = context.TODO()
	}

	// suppress:G204
	cmd := exec.CommandContext(ctx, cr.Name, cr.Args...)
	cmd.Env = cr.Env

	return capture(cmd)
}

// Capture re-runs the Result's parameters in a new shell, recording
// A result in the same way as command.Capture.
func (cr Result) Capture() (Result, error) {
	return cr.CaptureContext(context.Background())
}
