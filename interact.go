package command

import "bufio"

type Interact struct {
	Plumber
	Controller
}

// Plumber defines functions on Pipes that handle modes of interaction.
// It is meant as a template, not as a literal interface, an serves an illustration
// of the required interactivity.
type Plumber interface {
	// RawPipes returns the pipes for direct manipulation by the end-user
	Pipes() *Pipes

	// CloseInput is required by some applications to close the input pipe
	// and stop interaction. This signals the application with pipe control
	// that it is OK to exit.
	CloseInput() error
}

// StdoutScanner returns a bufio.Scanner over stdio.
func (i *Interact) StdoutScanner() *bufio.Scanner {
	return bufio.NewScanner(i.Plumber.Pipes().Stdout)
}

// StderrScanner returns a bufio.Scanner over stderr.
func (i *Interact) StderrScanner() *bufio.Scanner {
	return bufio.NewScanner(i.Plumber.Pipes().Stderr)
}

// Controller defines the process control portion of the command and what
// users can do with it. It is, again, an illustration of possibilities.
type Controller interface {
	// Stop immediately cancels the context and tries
	// very hard to kill it.
	Stop() error

	// Wait calls the underlying Wait() function and
	// will wait indefinitely until a process is finished.
	Wait() error
}
