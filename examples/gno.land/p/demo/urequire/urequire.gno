// urequire is a sister package for uassert.
// XXX: codegen the package.
package urequire

import "gno.land/p/demo/uassert"

// type TestingT = uassert.TestingT // XXX: bug, should work

func NoError(t uassert.TestingT, err error, msgs ...string) {
	t.Helper()
	if uassert.NoError(t, err, msgs...) {
		return
	}
	t.FailNow()
}

func Error(t uassert.TestingT, err error, msgs ...string) {
	t.Helper()
	if uassert.Error(t, err, msgs...) {
		return
	}
	t.FailNow()
}

func True(t uassert.TestingT, value bool, msgs ...string) {
	t.Helper()
	if uassert.True(t, value, msgs...) {
		return
	}
	t.FailNow()
}

func False(t uassert.TestingT, value bool, msgs ...string) {
	t.Helper()
	if uassert.False(t, value, msgs...) {
		return
	}
	t.FailNow()
}

func ErrorIs(t uassert.TestingT, err, target error, msgs ...string) {
	t.Helper()
	if uassert.ErrorIs(t, err, target, msgs...) {
		return
	}
	t.FailNow()
}

func PanicsWithError(t uassert.TestingT, errString string, f func(), msgs ...string) {
	t.Helper()
	if uassert.PanicsWithError(t, errString, f, msgs...) {
		return
	}
	t.FailNow()
}

func NotPanics(t uassert.TestingT, f func(), msgs ...string) {
	t.Helper()
	if uassert.NotPanics(t, f, msgs...) {
		return
	}
	t.FailNow()
}

func Equal(t uassert.TestingT, expected, actual interface{}, msgs ...string) {
	t.Helper()
	if uassert.Equal(t, expected, actual, msgs...) {
		return
	}
	t.FailNow()
}
