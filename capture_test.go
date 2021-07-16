// +build !windows

package command_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

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
			want: command.Capture{
				Output:   []byte(""),
				ExitCode: 0,
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
			want: command.Capture{
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
			want: command.Capture{
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
			want: command.Capture{
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
			want: command.Capture{
				Name:     "/bin/bash",
				Args:     []string{"-c", "echo $A $B"},
				Output:   []byte("A B\n"),
				ExitCode: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.want
			capture := command.New(command.WithDir(tt.args.dir), command.WithEnv(tt.args.env))
			got, err := capture.Run(tt.args.ctx, tt.args.name, tt.args.args...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			AllMatcher(t, want, got)
		})
	}
}

func TestCapture_Rerun(t *testing.T) {
	capture := command.New()

	want, err := capture.Run(context.Background(), "/bin/bash", "-c", "echo $A $B")
	if err != nil {
		t.Fatalf("setup failed with error: %v", err)
	}

	got, err := want.Rerun(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	AllMatcher(t, want, got)
}

func TestCapture_Marshaling(t *testing.T) {
	cmd := command.New(command.WithEnv([]string{"A=A"}), command.WithDir("/tmp"))

	want, err := cmd.Run(context.Background(), "ls", "-la")
	if err != nil {
		t.Fatalf("expected no error")
	}

	var marshaled []byte

	marshaled, err = json.Marshal(want)
	if err != nil {
		t.Fatalf("Marshal(): expected no error")
	}

	var got command.Capture

	err = json.Unmarshal(marshaled, &got)
	if err != nil {
		t.Fatalf("Unmarshal(): expected no error")
	}

	AllMatcher(t, want, got)
}

func AllMatcher(t *testing.T, want, got command.Capture) {
	ExitCodeMatcher(t, want, got)
	OutputMatcher(t, want, got)
	ArgsMatcher(t, want, got)
	NameMatcher(t, want, got)
	DirMatcher(t, want, got)
}

func OutputMatcher(t *testing.T, want, got command.Capture) {
	t.Run("Output", func(t *testing.T) {
		t.Run("NilNess", func(t *testing.T) {
			if (want.Output == nil) != (got.Output == nil) {
				t.Errorf("want.Output == nil: %t, got.Output == nil: %t", want.Output == nil, got.Output == nil)
			}
		})

		t.Run("DeepEqual", func(t *testing.T) {
			if !reflect.DeepEqual(want.Output, got.Output) {
				t.Errorf("want.Output: '%v', got.Output: '%v'", want.Output, got.Output)
			}
		})
	})
}

func NameMatcher(t *testing.T, want, got command.Capture) {
	t.Run("Name", func(t *testing.T) {
		t.Run("Equal", func(t *testing.T) {
			if want.Name != got.Name {
				t.Errorf("want.Name: %s, got.Name: %s", want.Name, got.Name)
			}
		})
	})
}

func ArgsMatcher(t *testing.T, want, got command.Capture) {
	t.Run("Args", func(t *testing.T) {
		t.Run("DeepEqual", func(t *testing.T) {
			if !reflect.DeepEqual(want.Args, got.Args) {
				t.Errorf("want.Args: %v, got.Args: %v", want.Args, got.Args)
			}
		})

		t.Run("NilNess", func(t *testing.T) {
			if (want.Args == nil) != (got.Args == nil) {
				t.Errorf("want.Args == nil: %t, got.Args == nil: %t", want.Args == nil, got.Args == nil)
			}
		})

		t.Run("Length", func(t *testing.T) {
			if len(want.Args) != len(got.Args) {
				t.Errorf("len(want.Args): %d, len(got.Args): %d", len(want.Args), len(got.Args))
			}
		})
	})
}

func DirMatcher(t *testing.T, want, got command.Capture) {
	t.Run("Dir", func(t *testing.T) {
		if want.Dir != "" {
			t.Run("Equal", func(t *testing.T) {
				if want.Dir != got.Dir {
					t.Errorf("want.Dir: %s, got.Dir: %s", want.Dir, got.Dir)
				}
			})
		}
	})
}

func ExitCodeMatcher(t *testing.T, want, got command.Capture) {
	t.Run("ExitCode", func(t *testing.T) {
		t.Run("Equal", func(t *testing.T) {
			if want.ExitCode != got.ExitCode {
				t.Errorf("want.ExitCode: %d, got.ExitCode: %d", want.ExitCode, got.ExitCode)
			}
		})
	})
}
