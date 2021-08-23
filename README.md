# command

This package wraps the functionality of `*exec.Cmd` with some simple IO tools to make creating a `Cmd` easier. It provides
the 

## Goal

The project's goal is ease of use when configuring, starting/stopping, and directly capturing output of a `exec.Cmd` call.

## Use Cases

Not knowing where to start with the sometimes daunting `*exec.Cmd`, you can use this library to simplify the process.

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

## Additional functionality

We added additional "Kill" functionality to the library for your convenience. As always, you can also cancel the context you're handing off to the Cmd if you want a shortcut.

```go
// Kill immediately stops the process via context or signal, depending on how you set your command up.
Kill() error

// KillTimer waits on a Sleep before killing the process.
KillTimer(d time.Duration)

// KillAfter waits until the specified time is reached before closing.
KillAfter(t time.Time)
```

## Testing

This package only depends upon our own [wrapt](https://github.com/metrumresearchgroup/wrapt/) testing library. Running `make test` is sufficient to verify its contents.

We include .golangci.yml configuration and a .drone.yaml for quality purposes.