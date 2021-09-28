package command

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/metrumresearchgroup/wrapt"
)

func TestCmd_Kill_not_running(tt *testing.T) {
	t := wrapt.WrapT(tt)
	c := New("more", "-")
	t.R.Error(c.Kill())
}

func TestCmd_Kill(tt *testing.T) {
	tests := []struct {
		name    string
		cmd     *Cmd
		wantErr bool
	}{
		{
			name:    "New kill",
			cmd:     New("more", "-"),
			wantErr: true,
		},
		{
			name:    "NewWithContext kill",
			cmd:     NewWithContext(context.Background(), "more", "-"),
			wantErr: true,
		},
	}
	for _, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			t := wrapt.WrapT(tt)

			t.R.NoError(test.cmd.Start())
			defer t.R.NoError(test.cmd.Wait())

			t.R.WantError(test.wantErr, test.cmd.Kill())
		})
	}
}

func TestCmd_KillAfter(tt *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "base",
			args: args{
				d: 100 * time.Millisecond,
			},
		},
	}
	for _, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			t := wrapt.WrapT(tt)

			c := New("more", "-")
			errCh := make(chan error)
			c.KillAfter(time.Now().Add(test.args.d), errCh)

			t.R.NoError(c.Start())

			t.R.NoError(<-errCh)
		})
	}
}

func TestCmd_Impersonate(tt *testing.T) {
	type fields struct {
		Cmd        *exec.Cmd
		cancelFunc func()
	}
	type args struct {
		username string
		setPgid  bool
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantErr         bool
		wantEnv         []string
		wantSysProcAttr *syscall.SysProcAttr
	}{
		{
			name: "garbage username",
			fields: fields{
				Cmd: exec.Command("dummy"),
			},
			args: args{
				username: "asdfasdfsadf",
				setPgid:  false,
			},
			wantErr: true,
		},
		{
			name: "empty username",
			fields: fields{
				Cmd: exec.Command("dummy"),
			},
			args: args{
				username: "",
				setPgid:  false,
			},
			wantErr: true,
		},
		{
			name: "assign self to test",
			fields: fields{
				Cmd: func() *exec.Cmd {
					cmd := exec.Command("dummy")
					cmd.Env = []string{"BASE=ALL"}

					return cmd
				}(),
			},
			args: args{
				username: func() string {
					u, err := user.Current()
					if err != nil {
						panic(err)
					}
					if len(u.Username) == 0 {
						panic(errors.New("no user returned by user.Current()"))
					}
					return u.Username
				}(),
				setPgid: true,
			},
			wantErr: false,
			wantEnv: func() []string {
				home, err := os.UserHomeDir()
				if err != nil {
					panic(err)
				}
				u, err := user.Current()
				un := u.Username
				return []string{
					"BASE=ALL",
					"USER=" + un,
					"HOME=" + home,
				}
			}(),
			wantSysProcAttr: &syscall.SysProcAttr{
				Setpgid: true,
				Credential: func() *syscall.Credential {
					u, err := user.Current()
					if err != nil {
						panic(err)
					}
					uid, err := strconv.Atoi(u.Uid)
					if err != nil {
						panic(err)
					}
					gid, err := strconv.Atoi(u.Gid)
					if err != nil {
						panic(err)
					}
					return &syscall.Credential{
						Gid: uint32(gid),
						Uid: uint32(uid),
					}
				}(),
			},
		},
	}
	for _, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			t := wrapt.WrapT(tt)

			c := &Cmd{
				Cmd:        test.fields.Cmd,
				cancelFunc: test.fields.cancelFunc,
			}

			err := c.Impersonate(test.args.username, test.args.setPgid)
			t.R.WantError(test.wantErr, err)

			if test.wantErr {
				return
			}

			t.R.Equal(test.wantEnv, c.Env)
			t.R.Equal(test.wantSysProcAttr, c.SysProcAttr)
		})
	}
}
