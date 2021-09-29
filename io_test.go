//go:build !windows
// +build !windows

package command

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/metrumresearchgroup/wrapt"
)

func TestIO_Apply(tt *testing.T) {
	// a reminder: stdin must be a closeable pipe.
	tt.Run("stdin", func(tt *testing.T) {
		t := wrapt.WrapT(tt)

		// set up a pipe, we'll pass the inReader to the stdin, and
		// the inWriter will be used to Fprint*.
		inReader, inWriter, err := os.Pipe()
		t.R.NoError(err)

		// Simple output buffer, the WireIO handles the copying.
		output := &bytes.Buffer{}

		c := New("cat")
		t.R.NoError(err)

		i := WireIO(inReader, output, nil)
		i.Apply(c)

		t.R.NoError(c.Start())
		defer func() { t.R.Error(c.Kill()) }()

		_, err = fmt.Fprintln(inWriter, "hello")
		t.R.NoError(err)

		// Cat takes > 1ms to do its job on my desktop. Since we can't
		// scan for when it's willing to accept input, we wait before
		// closing our writer.
		time.Sleep(1 * time.Second)

		t.R.NoError(i.CloseInput())

		v, err := ioutil.ReadAll(output)
		t.R.NoError(err)

		// hello\n under unix, 6 chars
		t.R.Equal(6, len(v))
	})

	tt.Run("stdout", func(tt *testing.T) {
		t := wrapt.WrapT(tt)

		c := New("echo", "hello")

		buf := &bytes.Buffer{}
		i := WireIO(nil, buf, nil)
		i.Apply(c)

		err := c.Run()
		t.R.NoError(err)

		v, err := ioutil.ReadAll(buf)
		t.R.NoError(err)

		// hello\n under unix, 6 chars
		t.R.Equal(6, len(v))
	})

	tt.Run("stderr", func(tt *testing.T) {
		t := wrapt.WrapT(tt)

		c := New("sh", "-c", `echo "hello" 1>&2`)

		buf := &bytes.Buffer{}
		i := WireIO(nil, nil, buf)
		i.Apply(c)

		err := c.Run()
		t.R.NoError(err)

		v, err := ioutil.ReadAll(buf)
		t.R.NoError(err)

		// hello\n under unix, 6 chars
		t.R.Equal(6, len(v))
	})
}
