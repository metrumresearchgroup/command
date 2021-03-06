# command

This package wraps the functionality of `*exec.Cmd` with some simple IO tools to make creating a `Cmd` easier.

## Goal

The project's goal is ease of use when configuring, starting/stopping, and directly capturing output of a `exec.Cmd`
call.

## Use Cases

Not knowing where to start with the sometimes daunting `*exec.Cmd`, you can use this library to simplify the process.
The entry points are `New` and `NewWithContext`. You can Kill a process without knowing/caring how it gets done.

## Usage

Simple interactive use:

```go
c := command.New("vi")
command.InteractiveIO().Apply(c)
_ = c.Start() // ignoring err for demonstration purposes
_ = c.Wait() 
```

Programmatic input, standard output/err:

```go
reader, writer, _ := os.Pipe()
c := cmd.New("vi")
command.WireIO(reader, os.Stdout, os.Stderr).Apply(c)
_ = c.Start()
_, _ = fmt.Fprintln(writer, ":q!")
_ = c.Wait()
```

If you want to use a context:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel

c := command.NewWithContext(ctx, "name", "arg1", …)
```

## Availability of base functionality

Everything in `*exec.Cmd` is available. See the [official docs](https://pkg.go.dev/os/exec#Cmd) for expanded help:

```go
type Cmd struct {
Path string
Args []string
Env []string
Dir string
Stdin io.Reader
Stdout io.Writer
Stderr io.Writer
ExtraFiles []*os.File
SysProcAttr *syscall.SysProcAttr
Process *os.Process
ProcessState *os.ProcessState
}

func (c *Cmd) CombinedOutput() ([]byte, error)
func (c *Cmd) Output() ([]byte, error)
func (c *Cmd) Run() error
func (c *Cmd) Start() error
func (c *Cmd) StderrPipe() (io.ReadCloser, error)
func (c *Cmd) StdinPipe() (io.WriteCloser, error)
func (c *Cmd) StdoutPipe() (io.ReadCloser, error)
func (c *Cmd) String() string
func (c *Cmd) Wait() error
```

## Killing a process

We added additional "Kill" functionality to the library for your convenience. As always, you can also cancel the context
you're handing off to the Cmd if you want a shortcut.

```go
// Kill ends a process. Its operation depends on whether you created the Cmd
// with a context or not.
Kill() error

// KillTimer waits for the duration stated and then sends back the results
// of calling Kill via the errCh channel.
KillTimer(d time.Duration, errCh chan<- error)

// KillAfter waits until the time stated and then sends back the results
// of calling Kill via the errCh channel.
KillAfter(t time.Time, errCh chan<- error)
```

The standard library
function [signal.NotifyContext](https://cs.opensource.google/go/go/+/go1.17.1:src/os/signal/signal.go;l=277) is one
provides a convenient contextual hook over interrupts. Remember that putting a hook on an interrupt will intercept it,
and can prevent exit if you don't take manual steps to gracefully shut down.

```go
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
cmd := NewWithContext(ctx, "ls")
// etc…
```

## Impersonating a user

You can impersonate a valid user by using the `Impersonate(username string) error` function:

```go
cmd := New("ls", "/")
_ = cmd.Impersonate("username", false)

out, _ := cmd.CombinedOutput()
fmt.Println(out)
```

Your user needs to be capable of assuming the target user otherwise you'll receive an `operation not permitted` error.

## Testing

This package only depends upon our own [wrapt](https://github.com/metrumresearchgroup/wrapt/) testing library.
Running `make test` is sufficient to verify its contents.

We include .golangci.yml configuration and a .drone.yaml for quality purposes.

## Demo

There is a demo available that steps through different interaction modes in the `demo` directory of this project.