package command

import (
	"bufio"
	"io"
)

// Pipes represents the trinity of standard io pipes
// under POSIX.
type Pipes struct {
	Stdin  io.WriteCloser
	Stdout io.Reader
	Stderr io.Reader
}

// Pipes Returns the raw types object, because
// this may be something we want to make by scratch
// in a future type, such as in situations where you
// filter/wrap output.
func (p *Pipes) Pipes() *Pipes {
	return p
}

// StdoutScanner returns a bufio.Scanner over stdio.
func (p *Pipes) StdoutScanner() *bufio.Scanner {
	return bufio.NewScanner(p.Stdout)
}

// StderrScanner returns a bufio.Scanner over stderr.
func (p *Pipes) StderrScanner() *bufio.Scanner {
	return bufio.NewScanner(p.Stderr)
}

// CloseInput closes stdin. Sometimes required to let
// a process natually terminate when using pipes.
func (p *Pipes) CloseInput() error {
	return p.Stdin.Close()
}
