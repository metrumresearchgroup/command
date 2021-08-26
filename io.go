package command

import (
	"fmt"
	"io"
	"os"
)

type IO struct {
	Stdin  io.ReadCloser
	Stdout io.Writer
	Stderr io.Writer
}

func (i *IO) Apply(c *Cmd) {
	if i.Stdin != nil {
		c.Stdin = i.Stdin
	}
	if i.Stdout != nil {
		c.Stdout = i.Stdout
	}
	if i.Stderr != nil {
		c.Stderr = i.Stderr
	}
}

func (i *IO) CloseInput() error {
	if i.Stdin != nil {
		return i.Stdin.Close()
	}

	// not having stdin wired is not an error.
	return nil
}

func InteractiveIO() *IO {
	return &IO{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// WireIO sets up pipes on readers/writers the user supplies to take over
// all IO taps of the program.
func WireIO(stdin io.Reader, stdout, stderr io.Writer) *IO {
	i := &IO{}
	if stdin != nil {
		stdinReader, stdinWriter, err := os.Pipe()
		if err != nil {
			panic(err)
		}
		go func() {
			if _, err := io.Copy(stdinWriter, stdin); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}()
		i.Stdin = stdinReader
	}

	if stdout != nil {
		stdoutReader, stdoutWriter, err := os.Pipe()
		if err != nil {
			panic(err)
		}
		go func() {
			if _, err := io.Copy(stdout, stdoutReader); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}()
		i.Stdout = stdoutWriter
	}

	if stderr != nil {
		stderrReader, stderrWriter, err := os.Pipe()
		if err != nil {
			panic(err)
		}
		go func() {
			if _, err := io.Copy(stderr, stderrReader); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
			}
		}()
		i.Stderr = stderrWriter
	}

	return i
}
