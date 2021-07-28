package pipes_test

import (
	"errors"
	"io"
	"os/exec"
	"testing"

	. "github.com/metrumresearchgroup/wrapt"

	"github.com/metrumresearchgroup/command/pipes"
)

func Test_Attach(tt *testing.T) {
	t := WrapT(tt)

	c := exec.Command("test")

	p, err := pipes.Attach(c)

	t.A.NoError(err)

	t.A.NotNil(p.Stdin)
	t.A.NotNil(p.Stdout)
	t.A.NotNil(p.Stderr)
}

func Test_Failures(tt *testing.T) {
	t := WrapT(tt)

	ep := erroringPlumber(0)
	p, err := pipes.Attach(ep)
	t.A.Error(err)
	t.A.Nil(p)

	ep = erroringPlumber(1)
	p, err = pipes.Attach(ep)
	t.A.Error(err)
	t.A.Nil(p)

	ep = erroringPlumber(2)
	p, err = pipes.Attach(ep)
	t.A.Error(err)
	t.A.Nil(p)
}

type erroringPlumber int

func (e erroringPlumber) StdinPipe() (io.WriteCloser, error) {
	return e.errOnMatch(0)
}

func (e erroringPlumber) StdoutPipe() (io.ReadCloser, error) {
	return e.errOnMatch(1)
}

func (e erroringPlumber) StderrPipe() (io.ReadCloser, error) {
	return e.errOnMatch(2)
}

//nolint:unparam
func (e erroringPlumber) errOnMatch(n int) (io.ReadWriteCloser, error) {
	if int(e) == n {
		return nil, errors.New("error")
	}

	return nil, nil
}
