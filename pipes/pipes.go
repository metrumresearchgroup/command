package pipes

import (
	"io"
)

type Pipes struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}

type Plumber interface {
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
}

func Attach(pbr Plumber) (*Pipes, error) {
	var p Pipes

	stdin, err := pbr.StdinPipe()
	if err != nil {
		return nil, err
	}
	p.Stdin = stdin

	stdout, err := pbr.StdoutPipe()
	if err != nil {
		return nil, err
	}
	p.Stdout = stdout

	stderr, err := pbr.StderrPipe()
	if err != nil {
		return nil, err
	}
	p.Stderr = stderr

	return &p, nil
}
