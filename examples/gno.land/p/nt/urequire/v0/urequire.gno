// urequire is a sister package for uassert.
// XXX: codegen the package.
package urequire

import "gno.land/p/nt/uassert/v0"

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

func ErrorContains(t uassert.TestingT, err error, contains string, msgs ...string) {
	t.Helper()
	if uassert.ErrorContains(t, err, contains, msgs...) {
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

// AbortsWithMessage requires that the code inside the specified func aborts
// (panics when crossing another realm).
// Use PanicsWithMessage for requiring local panics within the same realm.
// Note: This relies on gno's `revive` mechanism to catch aborts.
func AbortsWithMessage(t uassert.TestingT, msg string, f any, msgs ...string) {
	t.Helper()
	if uassert.AbortsWithMessage(t, msg, f, msgs...) {
		return
	}
	t.FailNow()
}

// NotAborts requires that the code inside the specified func does NOT abort
// when crossing an execution boundary (e.g., VM call).
// Use NotPanics for requiring the absence of local panics within the same realm.
// Note: This relies on Gno's `revive` mechanism.
func NotAborts(t uassert.TestingT, f any, msgs ...string) {
	t.Helper()
	if uassert.NotPanics(t, f, msgs...) {
		return
	}
	t.FailNow()
}

// PanicsWithMessage requires that the code inside the specified func panics
// locally within the same execution realm.
// Use AbortsWithMessage for requiring panics that cross execution boundaries (aborts).
func PanicsWithMessage(t uassert.TestingT, msg string, f any, msgs ...string) {
	t.Helper()
	if uassert.PanicsWithMessage(t, msg, f, msgs...) {
		return
	}
	t.FailNow()
}

// NotPanics requires that the code inside the specified func does NOT panic
// locally within the same execution realm.
// Use NotAborts for requiring the absence of panics that cross execution boundaries (aborts).
func NotPanics(t uassert.TestingT, f any, msgs ...string) {
	t.Helper()
	if uassert.NotPanics(t, f, msgs...) {
		return
	}
	t.FailNow()
}

func Equal(t uassert.TestingT, expected, actual any, msgs ...string) {
	t.Helper()
	if uassert.Equal(t, expected, actual, msgs...) {
		return
	}
	t.FailNow()
}

func NotEqual(t uassert.TestingT, expected, actual any, msgs ...string) {
	t.Helper()
	if uassert.NotEqual(t, expected, actual, msgs...) {
		return
	}
	t.FailNow()
}

func Empty(t uassert.TestingT, obj any, msgs ...string) {
	t.Helper()
	if uassert.Empty(t, obj, msgs...) {
		return
	}
	t.FailNow()
}

func NotEmpty(t uassert.TestingT, obj any, msgs ...string) {
	t.Helper()
	if uassert.NotEmpty(t, obj, msgs...) {
		return
	}
	t.FailNow()
}
