package wrapt_test

import (
	"errors"
	"testing"

	"github.com/metrumresearchgroup/command/wrapt"
)

func Test_WrapT(tt *testing.T) {
	t := wrapt.WrapT(tt)
	t.RunFatal("positive assertion without failure", func(t *wrapt.T) {
		t.A.True(true, "not true")
	})
}

func Test_ResultHandler(tt *testing.T) {
	t := wrapt.WrapT(tt)

	t.ResultHandler = func(t *wrapt.T, success bool, format string, args ...interface{}) bool {
		return success
	}

	retSuccess := t.RunFatal("antifail", func(t *wrapt.T) {
		// do nothing
	})

	t.A.True(retSuccess)
}

func Test_AssertError(tt *testing.T) {
	t := wrapt.WrapT(tt)

	t.AssertError(false, nil)
	t.AssertError(true, errors.New("error"))
}

func Test_ValidateError(tt *testing.T) {
	t := wrapt.WrapT(tt)

	t.ValidateError("err", false, nil)
	t.ValidateError("err", true, errors.New("error"))
}
