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

	// StdoutScanner returns a bufio.Scanner that allows the user
	// to follow stdout output on a line-by-line basis. Since it returns
	// a Scanner, you can change the scan method at will
	StdoutScanner() *bufio.Scanner

	// StderrScanner is similar to StdoutScanner, but with Stderr.
	StderrScanner() *bufio.Scanner

	// CloseInput is required by some applications to close the input pipe
	// and stop interaction. This signals the application with pipe control
	// that it is OK to exit.
	CloseInput() error
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
