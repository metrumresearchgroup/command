package command

import (
	"context"
	"errors"
	"os/exec"
	user "os/user"
	"strconv"
	"syscall"
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

// Impersonate does a sets the SysProcAttrs to impersonate a permitted
// user.
func (c *Cmd) Impersonate(username string, setPgid bool) error {
	usr, cred, err := userCredential(username)
	if err != nil {
		return err
	}

	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    setPgid,
		Credential: cred,
	}

	if len(usr.Username) != 0 {
		c.Env = append(c.Env, "USER="+usr.Username)
	}

	if len(usr.HomeDir) != 0 {
		c.Env = append(c.Env, "HOME="+usr.HomeDir)
	}

	return nil
}

func userCredential(username string) (*user.User, *syscall.Credential, error) {
	if len(username) == 0 {
		return nil, nil, errors.New("username empty")
	}

	var (
		u *user.User
		c syscall.Credential
	)
	{
		var err error

		u, err = user.Lookup(username)
		if err != nil {
			return nil, nil, err
		}

		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			return nil, nil, err
		}

		gid, err := strconv.Atoi(u.Gid)
		if err != nil {
			return nil, nil, err
		}

		c = syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		}
	}

	return u, &c, nil
}
