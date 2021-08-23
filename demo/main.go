package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/metrumresearchgroup/command"
)

const timeout = 5 * time.Second

func main() {
	fmt.Print("\n\tthis is a cat test with pipes\n")
	piper()

	fmt.Print("\n\tTHIS IS INTERACTIVE, type something^M exit\n")
	interact()

	fmt.Print("\n\tINPUT is program-driven, from machine input\n")
	input()

	fmt.Printf("\n\tWill be killed by context timeout in %s\n", timeout)
	killTimeout()

	fmt.Printf("\n\tWill be killed by KillTimer() command in %s\n", timeout)
	useKillTimer()

	fmt.Printf("\n\tWill be killed by KillAfter() command at ~%s\n", time.Now().Add(timeout))
	useKillAfter()
}

// input pipes input into the command; it does not scan output.
func input() {
	// create a input for the programmatic input
	funcInput, writer, err := os.Pipe()
	ee(err)

	c := command.New("read")

	// apply the pipe to the program.
	command.WireIO(funcInput, os.Stdout, os.Stderr).Apply(c)

	ee(c.Start())

	ee(fmt.Fprintln(writer, "hello"))

	ee(c.Wait())
}

// interact is is user interactive, it just uses the stdio pipes.
func interact() {
	c := command.New("read")
	command.InteractiveIO().Apply(c)
	ee(c.Start())
	ee(c.Wait())
}

// piper pipes input into the command and output into a buffer.
func piper() {
	c := command.New("cat")
	input, err := c.StdinPipe()
	ee(err)

	output, err := c.StdoutPipe()
	ee(err)
	go func() {
		ee(io.Copy(os.Stdout, output))
	}()

	command.WireIO(nil, nil, os.Stderr).Apply(c)

	ee(c.Start())

	ee(fmt.Fprintln(input, "hello world"))
	ee(input.Close())

	ee(c.Wait())
}

// This example has a deadline of when it will close. It closes by
// timed out context.
func killTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	c := command.NewWithContext(ctx, "read")
	command.InteractiveIO().Apply(c)

	ee(c.Start())
	fmt.Println(c.Wait())
}

// This example has a timer after which it will close.
func useKillTimer() {
	c := command.New("read")
	command.InteractiveIO().Apply(c)

	ee(c.Start())
	errCh := make(chan error)
	c.KillTimer(timeout, errCh)
	fmt.Println(<-errCh)
}

// This example has a deadline of when it will close.
func useKillAfter() {
	c := command.New("read")
	command.InteractiveIO().Apply(c)

	ee(c.Start())
	errCh := make(chan error)
	c.KillAfter(time.Now().Add(timeout), errCh)
	fmt.Println(<-errCh)
}

func ee(params ...interface{}) {
	if len(params) == 0 {
		return
	}
	p := params[len(params)-1]
	if err, ok := p.(error); ok {
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}
	}
}
