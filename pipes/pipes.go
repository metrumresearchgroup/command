package pipes

import (
	"io"
	"os/exec"
)

type Pipes struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}

func New() *Pipes {
	return &Pipes{}
}

func (p *Pipes) Attach(cmd *exec.Cmd) error {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	p.Stdin = stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	p.Stdout = stdout

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	p.Stderr = stderr

	return nil
}
