package command_test

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/metrumresearchgroup/wrapt"

	"github.com/metrumresearchgroup/command"
)

func TestNewOptions(tt *testing.T) {
	tests := []struct {
		name    string
		options []command.Option
		checkFn func(t *wrapt.T, c *command.Capture)
	}{
		{
			name:    "with no options",
			options: nil,
			checkFn: func(t *wrapt.T, c *command.Capture) {
				t.R.Equal([]string(nil), c.Env)
				t.R.Equal("", c.Dir)
			},
		},
		{
			name:    "with dir",
			options: []command.Option{command.WithDir("/tmp")},
			checkFn: func(t *wrapt.T, c *command.Capture) {
				t.R.Equal("/tmp", c.Dir)
			},
		},
		{
			name:    "with env",
			options: []command.Option{command.WithEnv([]string{"TMP=/tmp"})},
			checkFn: func(t *wrapt.T, c *command.Capture) {
				t.R.Equal([]string{"TMP=/tmp"}, c.Env)
			},
		},
		{
			name: "with modifier error",
			options: []command.Option{command.WithCmdModifier(func(cmd *exec.Cmd) error {
				return errors.New("error")
			})},
			checkFn: func(t *wrapt.T, c *command.Capture) {
				cmd := exec.Command("ls")
				err := c.ModifierFunc(cmd)
				t.R.Error(err)
				t.R.Equal(errors.New("error"), err)
			},
		},
		{
			name: "with modifier success",
			options: []command.Option{command.WithCmdModifier(func(cmd *exec.Cmd) error {
				cmd.Path = "/tmp"
				return nil
			})},
			checkFn: func(t *wrapt.T, c *command.Capture) {
				cmd := exec.Command("ls")
				err := c.ModifierFunc(cmd)
				t.R.NoError(err)
				t.R.Equal("/tmp", cmd.Path)
			},
		},
	}
	for _, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			t := wrapt.WrapT(tt)

			// Create all at once.
			c := command.New(test.options...)
			test.checkFn(t, c)

			// Create with With function.
			c = command.New()
			c.With(test.options...)
			test.checkFn(t, c)
		})
	}
}
