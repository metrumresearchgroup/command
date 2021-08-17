package command

import (
	"io"
)

// Pipes represents the trinity of standard io pipes
// under POSIX.
type Pipes struct {
	Stdin  io.WriteCloser
	Stdout io.Reader
	Stderr io.Reader
}

// Pipes Returns the raw pipes struct, because
// this may be something we want to make by scratch
// in a future type, such as in situations where you
// filter/wrap output.
func (p *Pipes) Pipes() *Pipes {
	return p
}

// CloseInput closes stdin. Sometimes required to let
// a process naturally terminate when using pipes.
func (p *Pipes) CloseInput() error {
	return p.Stdin.Close()
}
