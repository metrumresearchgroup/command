package command

import (
	"context"
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
