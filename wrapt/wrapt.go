// Package wrapt wraps a *testing.T and adds functionality such as RunFatal,
// and Run with its own extended features.
package wrapt

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// T is simply a wrap of *testing.T
type T struct {
	*testing.T
	A *assert.Assertions
}

// WrapT takes a *testing.T and returns the equivalent *T from it. This is the
// entry point into all functionality, and should be done at the top of any
// *testing.T test to impart the new functionality.
func WrapT(t *testing.T) *T {
	return &T{
		T: t,
		A: assert.New(t),
	}
}

// RunFatal is like t.Run() but stops the outer test once a fatal is raised.
// This is especially useful if you need an
// inner test to fail the outer test,
// but stop before going too high up in a cluster of tests.
func (t *T) RunFatal(name string, fn func(t *T)) {
	t.Helper()

	t.failTrap(name, fn)
}

// Run implements the standard testing.T.Run() by wrapping *testing.T for us so
// the inner test has full access to our *T.
func (t *T) Run(name string, fn func(t *T)) {
	t.Helper()

	t.T.Run(name, wrapFn(fn))
}

// failTrap wraps a regular testing function in a deferrer that fails the outer
// test as well as the inner tests in cases where tests are used to "fence"
// features off.
//
// There are situations where you want to fail an outer test, especially when
// the tests cannot continue. Previous to this, with *testing.T.Run(), the
// tests would bumble on past a point of no return, and this wasn't providing
// value.
func (t *T) failTrap(name string, fn func(t *T)) {
	t.Helper()

	var failed bool

	t.Run(name, func(t *T) {
		defer func() {
			v := recover()
			if t.Failed() {
				failed = true
			}
			if v != nil {
				panic(v)
			}
		}()

		fn(t)
	})

	if failed {
		t.Fatalf("RunFatal(): inner test failed")
	}
}

func (t *T) StopOnError(desc string, err error) (failed bool) {
	t.Helper()

	return t.ValidateError(desc, false, err)
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

func (t *T) AssertError(wantErr bool, err error) (failed bool) {
	t.Helper()

	if wantErr {
		t.A.Error(err)
	} else {
		t.A.NoError(err)
	}

	return (err != nil) != wantErr
}

// DeepEqual performs a reflect.DeepEqual check with a consistent message
// and dump.
func (t *T) DeepEqual(desc string, want, got interface{}) {
	t.Helper()

	desc = colonize(desc)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%swant: %v, got %v", desc, pp(want), pp(got))
	}
}

func pp(v interface{}) interface{} {
	switch v2 := v.(type) {
	case []byte:
		return string(v2)
	default:
		return v
	}
}

func colonize(s string) string {
	if len(s) == 0 {
		return s
	}

	if strings.HasSuffix(s, ":") {
		s += " "
	} else {
		s += ": "
	}

	return s
}

// wrapFn wraps a function taking *T with a function that looks like it takes
// *testing.T so the testing.T.Run() can operate.
func wrapFn(fn func(*T)) func(*testing.T) {
	return func(tt *testing.T) {
		fn(WrapT(tt))
	}
}
