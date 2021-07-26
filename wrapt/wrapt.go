// Package wrapt wraps a *testing.T and adds functionality such as RunFatal,
// and Run with its own extended features.
package wrapt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// T is simply a wrap of *testing.T.
type T struct {
	*testing.T
	A             *assert.Assertions
	ResultHandler func(t *T, success bool, format string, args ...interface{}) bool
}

// WrapT takes a *testing.T and returns the equivalent *T from it. This is the
// entry point into all functionality, and should be done at the top of any
// *testing.T test to impart the new functionality.
func WrapT(t *testing.T) *T {
	return &T{
		T: t,
		A: assert.New(t),
		ResultHandler: func(t *T, success bool, format string, args ...interface{}) bool {
			if !success {
				t.Fatalf(format, args)
			}

			return success
		},
	}
}

func (t *T) wrapT(tt *testing.T) *T {
	return &T{
		T:             tt,
		A:             assert.New(tt),
		ResultHandler: t.ResultHandler,
	}
}

// RunFatal is like t.Run() but stops the outer test once a fatal is raised.
// This is especially useful if you need an
// inner test to fail the outer test,
// but stop before going too high up in a cluster of tests.
func (t *T) RunFatal(name string, fn func(t *T)) (success bool) {
	t.Helper()

	return t.ResultHandler(t, t.Run(name, fn), "inner test failed")
}

// Run implements the standard testing.T.Run() by wrapping *testing.T for us so
// the inner test has full access to our *T.
func (t *T) Run(name string, fn func(t *T)) (success bool) {
	t.Helper()

	return t.T.Run(name, t.wrapFn(fn))
}

// ValidateError will check an error vs an expected state and fail the outer
// test if the validation fails.
func (t *T) ValidateError(desc string, wantErr bool, err error) (failed bool) {
	t.Helper()

	t.RunFatal(desc, func(t *T) {
		failed = t.AssertError(wantErr, err)
	})

	return failed
}

func (t *T) AssertError(wantErr bool, err error) (success bool) {
	t.Helper()

	if wantErr {
		success = t.A.Error(err)
	} else {
		success = t.A.NoError(err)
	}

	return success
}

// wrapFn wraps a function taking *T with a function that looks like it takes
// *testing.T so the testing.T.Run() can operate.
func (t *T) wrapFn(fn func(*T)) func(*testing.T) {
	return func(tt *testing.T) {
		fn(t.wrapT(tt))
	}
}
