// +build !windows

package command_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"testing"

	. "github.com/metrumresearchgroup/wrapt"

	"github.com/metrumresearchgroup/command"
)

//goland:noinspection GoNilness
func TestCapture_Run(t *testing.T) {
	type args struct {
		ctx  context.Context
		dir  string
		env  []string
		name string
		args []string
	}
	tests := []struct {
		name        string
		args        args
		wantCapture *command.Capture
		wantStdout  []byte
		wantStderr  []byte
		wantErr     bool
	}{
		{
			name: "invalid name",
			args: args{
				ctx:  context.Background(),
				name: "asdfasdf",
			},
			wantErr: true,
			wantCapture: &command.Capture{
				Name:     "asdfasdf",
				ExitCode: 0,
			},
		},
		{
			name: "success return",
			args: args{
				ctx:  context.Background(),
				dir:  ".", // this is explicitly stated to test the WithDir below.
				name: "/bin/bash",
				args: []string{"-c", "exit 0"},
			},
			wantCapture: &command.Capture{
				ExitCode: 0,
				Dir:      ".",
				Name:     "/bin/bash",
				Args:     []string{"-c", "exit 0"},
			},
		},
		{
			name: "nonzero return",
			args: args{
				ctx:  context.Background(),
				name: "/bin/bash",
				args: []string{"-c", "exit 1"},
			},
			wantErr: true,
			wantCapture: &command.Capture{
				Name:     "/bin/bash",
				Args:     []string{"-c", "exit 1"},
				Env:      nil,
				ExitCode: 1,
			},
		},
		{
			name: "ctx canceled",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()

					return ctx
				}(),
				name: "/bin/bash",
				args: []string{"-c", "exit 0"},
			},
			wantErr: true,
			wantCapture: &command.Capture{
				Name:     "/bin/bash",
				Args:     []string{"-c", "exit 0"},
				Dir:      "",
				Env:      nil,
				ExitCode: 0,
			},
		},
		{
			name: "captures stderr",
			args: args{
				ctx:  context.Background(),
				name: "/bin/bash",
				args: []string{"-c", `echo "message" 1>&2`},
			},
			wantErr: false,
			wantCapture: &command.Capture{
				Name:     "/bin/bash",
				Args:     []string{"-c", `echo "message" 1>&2`},
				ExitCode: 0,
			},
			wantStderr: []byte("message\n"),
		},
		{
			name: "captures env",
			args: args{
				ctx: context.Background(),
				env: []string{
					"A=A",
					"B=B",
				},
				name: "/bin/bash",
				args: []string{"-c", "echo $A $B"},
			},
			wantErr: false,
			wantCapture: &command.Capture{
				Name:     "/bin/bash",
				Args:     []string{"-c", "echo $A $B"},
				ExitCode: 0,
			},
			wantStdout: []byte("A B\n"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(testingT *testing.T) {
			t := WrapT(testingT)
			capture := command.New(command.WithDir(test.args.dir), command.WithEnv(test.args.env))
			p, err := capture.Run(test.args.ctx, test.args.name, test.args.args...)
			t.R.WantError(test.wantErr, err)
			AllMatcher(t, test.wantCapture, capture)
			if test.wantStdout == nil {
				test.wantStdout = []byte{}
			}
			if test.wantStderr == nil {
				test.wantStderr = []byte{}
			}
			stdout, err := io.ReadAll(p.Stdout)
			t.A.NoError(err)
			t.A.Equal(test.wantStdout, stdout)
			stderr, err := io.ReadAll(p.Stderr)
			t.A.NoError(err)
			t.A.Equal(test.wantStderr, stderr)
		})
	}
}

func TestCapture_Rerun(tt *testing.T) {
	t := WrapT(tt)

	capture := command.New()

	p, err := capture.Run(context.Background(), "/bin/bash", "-c", "echo $A $B")
	if t.A.NoError(err) {
		return
	}
	wantOutput, err := io.ReadAll(p.Stdout)
	t.A.NoError(err)

	want := *capture

	p, err = capture.Rerun(context.Background())
	if t.A.NoError(err) {
		return
	}
	gotOutput, err := io.ReadAll(p.Stdout)
	t.A.NoError(err)

	AllMatcher(t, &want, capture)
	t.A.Equal(wantOutput, gotOutput)
}

func TestCapture_Marshaling(tt *testing.T) {
	t := WrapT(tt)

	cmd := command.New(command.WithEnv([]string{"A=A"}), command.WithDir("/tmp"))

	_, err := cmd.Run(context.Background(), "ls", "-la")
	if t.A.NoError(err) {
		return
	}

	marshaled, err := json.Marshal(cmd)
	if t.A.NoError(err) {
		return
	}

	got := new(command.Capture)
	err = json.Unmarshal(marshaled, got)
	t.A.NoError(err)

	AllMatcher(t, cmd, got)
}

func AllMatcher(t *T, want, got *command.Capture) {
	t.RunFatal("matchers", func(t *T) {
		ExitCodeMatcher(t, want, got)
		ArgsMatcher(t, want, got)
		NameMatcher(t, want, got)
		DirMatcher(t, want, got)
	})
}

func NameMatcher(t *T, want, got *command.Capture) {
	t.Run("Name", func(t *T) {
		t.A.Equal(want.Name, got.Name)
	})
}

func ArgsMatcher(t *T, want, got *command.Capture) {
	t.Run("Args", func(t *T) {
		t.A.Equal(want.Args, got.Args)
	})
}

func DirMatcher(t *T, want, got *command.Capture) {
	t.Run("Dir", func(t *T) {
		t.A.Equal(want.Dir, got.Dir)
	})
}

func ExitCodeMatcher(t *T, want, got *command.Capture) {
	t.Run("ExitCode", func(t *T) {
		t.A.Equal(want.ExitCode, got.ExitCode)
	})
}

func TestStartStop(tt *testing.T) {
	type args struct {
		ctx  context.Context
		dir  string
		env  []string
		name string
	}

	tests := []struct {
		name string
		args args
		act  func(t *T, capture *command.Capture) (i *command.Interact, err error)
		test func(t *T, capture *command.Capture, i *command.Interact, err error)
	}{
		{
			name: "stop without start",
			args: args{
				ctx:  context.Background(),
				name: "cat",
			},
			act: func(t *T, capture *command.Capture) (i *command.Interact, err error) {
				err = capture.Stop()

				return
			},
			test: func(t *T, capture *command.Capture, i *command.Interact, err error) {
				t.R.WantError(true, err)
			},
		},
		{
			name: "start",
			act: func(t *T, capture *command.Capture) (i *command.Interact, err error) {
				return capture.Start(context.Background(), "cat")
			},
			test: func(t *T, capture *command.Capture, i *command.Interact, err error) {
				t.A.NotNil(i)
				t.A.NoError(err)

				err = capture.Stop()
				t.A.NoError(err)
			},
		},
		{
			name: "start and use input",
			act: func(t *T, capture *command.Capture) (i *command.Interact, err error) {
				i, err = capture.Start(context.Background(), "cat")
				t.A.NoError(err)

				count, err := i.Pipes().Stdin.Write([]byte("hello\n"))
				t.A.Equal(6, count)
				t.A.NoError(err)

				return i, nil
			},
			test: func(t *T, capture *command.Capture, i *command.Interact, _ error) {
				err := i.Pipes().Stdin.Close()
				t.A.NoError(err)

				pos, err := ioutil.ReadAll(i.Pipes().Stdout)
				t.A.NoError(err)

				err = capture.Stop()
				t.A.NoError(err)

				t.A.Equal([]byte("hello\n"), pos)
			},
		},
		{
			name: "echo test",
			act: func(t *T, capture *command.Capture) (i *command.Interact, err error) {
				i, err = capture.Start(context.Background(), "echo", "foo")
				t.A.NoError(err)

				return i, nil
			},
			test: func(t *T, capture *command.Capture, i *command.Interact, _ error) {
				// t.A.NoError(p.Stdin.Close(), "close stdin")

				pos, err := io.ReadAll(i.Pipes().Stdout)
				t.A.NoError(err)

				err = capture.Stop()
				t.A.NoError(err)

				t.A.Equal("foo\n", string(pos))
			},
		},
		{
			name: "echo test, close before ReadAll",
			act: func(t *T, capture *command.Capture) (i *command.Interact, err error) {
				i, err = capture.Start(context.Background(), "echo", "foo")
				t.A.NoError(err)

				return i, nil
			},
			test: func(t *T, capture *command.Capture, i *command.Interact, _ error) {
				err := i.CloseInput()
				t.A.NoError(err)

				out, err := io.ReadAll(i.Pipes().Stdout)
				if !errors.Is(err, io.EOF) {
					t.A.NoError(err)
				}

				err = capture.Stop()
				t.A.NoError(err)

				t.A.Equal("foo\n", string(out))
			},
		},
	}

	for _, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			t := WrapT(tt)

			capture := command.New(command.WithDir(test.args.dir), command.WithEnv(test.args.env))
			t.RunFatal("check act", func(t *T) {
				t.A.NotNil(test.act)
			})
			t.RunFatal("check test", func(t *T) {
				t.A.NotNil(test.test)
			})

			var i *command.Interact
			var err error
			t.RunFatal("act", func(t *T) {
				i, err = test.act(t, capture)
			})

			t.RunFatal("test", func(t *T) {
				test.test(t, capture, i, err)
			})
		})
	}
}
