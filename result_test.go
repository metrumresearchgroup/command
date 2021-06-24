package command_test

import (
	"context"
	"reflect"
	"testing"

	"command"
)

func TestCaptureContext(t *testing.T) {
	type args struct {
		ctx     context.Context
		env     []string
		command string
		args    []string
	}
	tests := []struct {
		name    string
		args    args
		want    command.Result
		wantErr bool
	}{
		{
			name: "invalid command",
			args: args{
				ctx:     context.Background(),
				command: "asdfasdf",
			},
			wantErr: true,
			want: command.Result{
				Output:   "",
				ExitCode: 0,
			},
		},
		{
			name: "success return",
			args: args{
				ctx:     context.Background(),
				command: "/bin/bash",
				args:    []string{"-c", "exit 0"},
			},
			wantErr: false,
			want: command.Result{
				Output:   "",
				ExitCode: 0,
			},
		},
		{
			name: "nonzero return",
			args: args{
				ctx:     context.Background(),
				command: "/bin/bash",
				args:    []string{"-c", "exit 1"},
			},
			wantErr: true,
			want: command.Result{
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
				command: "/bin/bash",
				args:    []string{"-c", "exit 0"},
			},
			wantErr: true,
		},
		{
			name: "captures stderr",
			args: args{
				ctx:     context.Background(),
				command: "/bin/bash",
				args:    []string{"-c", `echo "message" 1>&2`},
			},
			wantErr: false,
			want: command.Result{
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
				command: "/bin/bash",
				args:    []string{"-c", "echo $A $B"},
			},
			wantErr: false,
			want: command.Result{
				Output:   "A B\n",
				ExitCode: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := command.CaptureContext(tt.args.ctx, tt.args.env, tt.args.command, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Capture() error = %v, wantErr %v", err, tt.wantErr)
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

func TestResult_Capture(t *testing.T) {
	want, err := command.Capture(nil, "/bin/bash", "-c", "echo $A $B")
	if err != nil {
		t.Fatalf("setup failed with error: %v", err)
	}

	got, err := want.Capture()
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
}
