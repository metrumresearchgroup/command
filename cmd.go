package command

import (
	"context"
	"errors"
	"os/exec"
	"time"
)

type Cmd struct {
	*exec.Cmd
	cancelFunc func()
}

func New(name string, args ...string) *Cmd {
	return &Cmd{
		Cmd: exec.Command(name, args...),
	}
}

func NewWithContext(ctx context.Context, name string, args ...string) *Cmd {
	ctx, cancel := context.WithCancel(ctx)

	return &Cmd{
		Cmd:        exec.CommandContext(ctx, name, args...),
		cancelFunc: cancel,
	}
}

// Kill ends a process. Its operation depends on whether you created the Cmd
// with a context or not.
func (c *Cmd) Kill() error {
	if c.cancelFunc != nil {
		c.cancelFunc()

		return c.Wait()
	}
	if c.Process != nil {
		if err := c.Process.Kill(); err != nil {
			return err
		}

		return c.Wait()
	}

	return errors.New("not running")
}

// KillTimer waits for the duration stated and then sends back the results
// of calling Kill via the errCh channel.
func (c *Cmd) KillTimer(d time.Duration, errCh chan<- error) {
	go func() {
		time.Sleep(d)
		errCh <- c.Kill()
	}()
}

// KillAfter waits until the time stated and then sends back the results
// of calling Kill via the errCh channel.
func (c *Cmd) KillAfter(t time.Time, errCh chan<- error) {
	d := time.Until(t)
	c.KillTimer(d, errCh)
}
