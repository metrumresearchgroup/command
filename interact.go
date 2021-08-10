package command

import "bufio"

type Interact struct {
	Plumber
	Controller

	outScanner, errScanner *bufio.Scanner
}

// Plumber defines functions on Pipes that handle modes of interaction.
// It is meant as a template, not as a literal interface, an serves an illustration
// of the required interactivity.
type Plumber interface {
	// Pipes returns the pipes for direct manipulation by the end-user
	Pipes() *Pipes

	// CloseInput is required by some applications to close the input pipe
	// and stop interaction. This signals the application with pipe control
	// that it is OK to exit.
	CloseInput() error
}

// StdoutScanner returns a bufio.Scanner over stdout.
func (i *Interact) StdoutScanner() *bufio.Scanner {
	if i.outScanner != nil {
		return i.outScanner
	}

	i.outScanner = bufio.NewScanner(i.Plumber.Pipes().Stdout)

	return i.outScanner
}

// StderrScanner returns a bufio.Scanner over stderr.
func (i *Interact) StderrScanner() *bufio.Scanner {
	if i.errScanner != nil {
		return i.errScanner
	}

	i.errScanner = bufio.NewScanner(i.Plumber.Pipes().Stderr)

	return i.errScanner
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
