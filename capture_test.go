// +build !windows

package command_test

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/metrumresearchgroup/command"
	"github.com/metrumresearchgroup/command/pipes"
	. "github.com/metrumresearchgroup/command/wrapt"
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
		name    string
		args    args
		want    *command.Capture
		wantErr bool
	}{
		{
			name: "invalid name",
			args: args{
				ctx:  context.Background(),
				name: "asdfasdf",
			},
			wantErr: true,
			want: &command.Capture{
				Name:     "asdfasdf",
				Output:   nil,
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
			want: &command.Capture{
				Output:   []byte(""),
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
			want: &command.Capture{
				Name:     "/bin/bash",
				Args:     []string{"-c", "exit 1"},
				Env:      nil,
				Output:   []byte(""),
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
			want: &command.Capture{
				Name:     "/bin/bash",
				Args:     []string{"-c", "exit 0"},
				Dir:      "",
				Env:      nil,
				Output:   nil,
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
			want: &command.Capture{
				Name:     "/bin/bash",
				Args:     []string{"-c", `echo "message" 1>&2`},
				Output:   []byte("message\n"),
				ExitCode: 0,
			},
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
			want: &command.Capture{
				Name:     "/bin/bash",
				Args:     []string{"-c", "echo $A $B"},
				Output:   []byte("A B\n"),
				ExitCode: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(testingT *testing.T) {
			t := WrapT(testingT)
			capture := command.New(command.WithDir(tt.args.dir), command.WithEnv(tt.args.env))
			err := capture.Run(tt.args.ctx, tt.args.name, tt.args.args...)
			t.AssertError(tt.wantErr, err)

			AllMatcher(t, tt.want, capture)
		})
	}
}

func TestCapture_Rerun(tt *testing.T) {
	t := WrapT(tt)

	capture := command.New()

	err := capture.Run(context.Background(), "/bin/bash", "-c", "echo $A $B")
	if t.A.NoError(err, "capture.Run()") {
		return
	}

	want := *capture

	err = capture.Rerun(context.Background())
	if t.A.NoError(err) {
		return
	}

	AllMatcher(t, &want, capture)
}

func TestCapture_Marshaling(tt *testing.T) {
	t := WrapT(tt)

	cmd := command.New(command.WithEnv([]string{"A=A"}), command.WithDir("/tmp"))

	err := cmd.Run(context.Background(), "ls", "-la")
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
		OutputMatcher(t, want, got)
		ArgsMatcher(t, want, got)
		NameMatcher(t, want, got)
		DirMatcher(t, want, got)
	})
}

func OutputMatcher(t *T, want, got *command.Capture) {
	t.Run("Output", func(t *T) {
		t.A.Equal(want.Output, got.Output)
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
	t := WrapT(tt)

	type args struct {
		ctx  context.Context
		dir  string
		env  []string
		name string
		args []string
	}

	tests := []struct {
		name string
		args args
		act  func(t *T, capture *command.Capture) (p *pipes.Pipes, err error)
		test func(t *T, capture *command.Capture, p *pipes.Pipes, err error)
	}{
		{
			name: "stop without start",
			args: args{
				ctx:  context.Background(),
				name: "cat",
			},
			act: func(t *T, capture *command.Capture) (p *pipes.Pipes, err error) {
				err = capture.Stop()

				return
			},
			test: func(t *T, capture *command.Capture, p *pipes.Pipes, err error) {
				t.ValidateError("want error", true, err)
			},
		},
		{
			name: "start",
			act: func(t *T, capture *command.Capture) (p *pipes.Pipes, err error) {
				return capture.Start(context.Background(), "cat")
			},
			test: func(t *T, capture *command.Capture, p *pipes.Pipes, err error) {
				t.A.NotNil(p)
				t.A.NoError(err)

				err = capture.Stop()
				t.A.NoError(err)
			},
		},
		{
			name: "start and use input",
			act: func(t *T, capture *command.Capture) (p *pipes.Pipes, err error) {
				p, err = capture.Start(context.Background(), "cat")
				t.A.NoError(err)

				count, err := p.Stdin.Write([]byte("hello\n"))
				t.A.Equal(6, count)
				t.A.NoError(err)

				return p, nil
			},
			test: func(t *T, capture *command.Capture, p *pipes.Pipes, err error) {
				err = p.Stdin.Close()
				t.A.NoError(err)

				pos, err := io.ReadAll(p.Stdout)
				t.A.NoError(err)

				err = capture.Stop()
				t.A.NoError(err)

				t.A.Equal([]byte("hello\n"), pos)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *T) {
			capture := command.New(command.WithDir(tt.args.dir), command.WithEnv(tt.args.env))
			t.RunFatal("check tt.act", func(t *T) {
				t.A.NotNil(tt.act)
			})
			t.RunFatal("check tt.test", func(t *T) {
				t.A.NotNil(tt.test)
			})

			var ps *pipes.Pipes
			var err error
			t.RunFatal("tt.act", func(t *T) {
				ps, err = tt.act(t, capture)
			})

			t.RunFatal("tt.test", func(t *T) {
				tt.test(t, capture, ps, err)
			})
		})
		/*
			// false stop:
			err := capture.Stop()
			t.ValidateError("Stop() without Start()", true, err)

			p, err := capture.Start(tt.args.ctx, tt.args.name, tt.args.args...)
			t.ValidateError("Start()", false, err)

			if _, err = p.Stdin.Write(tt.args.input); err != nil && err != io.EOF {
				t.Fatalf("stdin.Write(): err: %v", err)
			}

			err = p.Stdin.Close()
			t.StopOnError("stdin.Close()", err)

			output, err := io.ReadAll(p.Stdout)
			t.StopOnError("stdout.Readall()", err)

			t.Run("stdout.Read()", func(t *T) {
				t.A.Equal(tt.wantOutput, output)
			})

			t.Run(fmt.Sprintf("verify exit code"), func(t *T) {
				t.A.Equal(tt.want.ExitCode, capture.ExitCode, "exit code")
			})

			err = capture.Stop()
			t.ValidateError("Stop()", tt.wantErrStop, err)

			if capture.Output != nil {
				t.Fatal("Output: expected nil Output because pipes were set!")
			}*/
	}
}
