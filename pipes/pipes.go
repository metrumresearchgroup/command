package pipes

import (
	"io"
	"os"
)

type Pipes struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type Option func(*Pipes)

func New(options ...Option) *Pipes {
	if len(options) == 0 {
		options = []Option{
			WithStdin(os.Stdin),
			WithStdout(os.Stdout),
			WithStderr(os.Stderr),
		}
	}

	var p Pipes
	for _, option := range options {
		option(&p)
	}

	// back-fill missing values

	if p.Stdin == nil {
		p.Stdin = os.Stdin
	}
	if p.Stdout == nil {
		p.Stdout = os.Stdout
	}
	if p.Stderr == nil {
		p.Stderr = os.Stderr
	}

	return &p
}

func WithStdin(stdin io.Reader) func(*Pipes) {
	return func(p *Pipes) {
		p.Stdin = stdin
	}
}

func WithStdout(stdout io.Writer) func(*Pipes) {
	return func(p *Pipes) {
		p.Stdout = stdout
	}
}

func WithStderr(stderr io.Writer) func(*Pipes) {
	return func(p *Pipes) {
		p.Stderr = stderr
	}
}
