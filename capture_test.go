// +build !windows

package command_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/metrumresearchgroup/command"
)

//goland:noinspection GoNilness
func TestCapture(t *testing.T) {
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
		want    command.Capture
		wantErr bool
	}{
		{
			name: "invalid name",
			args: args{
				ctx:  context.Background(),
				name: "asdfasdf",
			},
			wantErr: true,
			want: command.Capture{
				Output:   "",
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
			wantErr: false,
			want: command.Capture{
				Output:   "",
				ExitCode: 0,
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
			want: command.Capture{
				Name:     "/bin/bash",
				Args:     []string{"-c", "exit 1"},
				Env:      nil,
				Output:   "",
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
		},
		{
			name: "captures stderr",
			args: args{
				ctx:  context.Background(),
				name: "/bin/bash",
				args: []string{"-c", `echo "message" 1>&2`},
			},
			wantErr: false,
			want: command.Capture{
				Output:   "message\n",
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
			want: command.Capture{
				Output:   "A B\n",
				ExitCode: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := command.New(command.WithDir(tt.args.dir), command.WithEnv(tt.args.env))
			got, err := capture.Run(tt.args.ctx, tt.args.name, tt.args.args...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Output != tt.want.Output {
				t.Errorf("mismatch in output: wanted %s, got %s", tt.want.Output, got.Output)
			}

			if got.ExitCode != tt.want.ExitCode {
				t.Errorf("mismatch in exitcode: wanted %d, got %d", tt.want.ExitCode, got.ExitCode)
			}
		})
	}
}

func TestRerun(t *testing.T) {
	capture := command.New()
	want, err := capture.Run(context.Background(), "/bin/bash", "-c", "echo $A $B")
	if err != nil {
		t.Fatalf("setup failed with error: %v", err)
	}

	got, err := want.Rerun(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if want.Output != got.Output {
		t.Errorf("mismatch in output: wanted %s, got %s", want.Output, got.Output)
	}

	if want.ExitCode != got.ExitCode {
		t.Errorf("mismatch in exitcode: wanted %d, got %d", want.ExitCode, got.ExitCode)
	}

	if !reflect.DeepEqual(want.Args, got.Args) {
		t.Errorf("mismatch in args: wanted %v, got %v", want.Args, got.Args)
	}

	if want.Name != got.Name {
		t.Errorf("mismatch in name: wanted %s, got %s", want.Name, got.Name)
	}

	if want.Dir != got.Dir {
		t.Errorf("mismatch in dir: wanted %s, got %s", want.Dir, got.Dir)
	}
}
