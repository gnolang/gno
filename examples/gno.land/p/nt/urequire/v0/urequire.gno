// urequire is a sister package for uassert.
// XXX: codegen the package.
package urequire

import "gno.land/p/nt/uassert/v0"

// type TestingT = uassert.TestingT // XXX: bug, should work

// NoError requires that a function returned no error (i.e. `nil`).
func NoError(t uassert.TestingT, err error, msgs ...string) {
	t.Helper()
	if uassert.NoError(t, err, msgs...) {
		return
	}
	t.FailNow()
}

// Error requires that a function returned an error (i.e. not `nil`).
func Error(t uassert.TestingT, err error, msgs ...string) {
	t.Helper()
	if uassert.Error(t, err, msgs...) {
		return
	}
	t.FailNow()
}

// ErrorContains requires that a function returned an error (i.e. not `nil`)
// and that the error contains the specified substring.
func ErrorContains(t uassert.TestingT, err error, contains string, msgs ...string) {
	t.Helper()
	if uassert.ErrorContains(t, err, contains, msgs...) {
		return
	}
	t.FailNow()
}

// True requires that the specified value is true.
func True(t uassert.TestingT, value bool, msgs ...string) {
	t.Helper()
	if uassert.True(t, value, msgs...) {
		return
	}
	t.FailNow()
}

// False requires that the specified value is false.
func False(t uassert.TestingT, value bool, msgs ...string) {
	t.Helper()
	if uassert.False(t, value, msgs...) {
		return
	}
	t.FailNow()
}

// ErrorIs requires that the given error matches the target error.
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
// See uassert.AbortsWithMessage for `rlm` semantics.
func AbortsWithMessage(t uassert.TestingT, rlm realm, msg string, f any, msgs ...string) {
	t.Helper()
	if uassert.AbortsWithMessage(t, rlm, msg, f, msgs...) {
		return
	}
	t.FailNow()
}

// AbortsContains requires that the code inside the specified func aborts
// (panics when crossing another realm) and the abort message contains the specified substring.
func AbortsContains(t uassert.TestingT, substr string, f any, msgs ...string) {
	t.Helper()
	if uassert.AbortsContains(t, substr, f, msgs...) {
		return
	}
	t.FailNow()
}

// NotAborts requires that the code inside the specified func does NOT abort
// when crossing an execution boundary (e.g., VM call).
// Use NotPanics for requiring the absence of local panics within the same realm.
// Note: This relies on Gno's `revive` mechanism.
// See uassert.AbortsWithMessage for `rlm` semantics.
func NotAborts(t uassert.TestingT, rlm realm, f any, msgs ...string) {
	t.Helper()
	if uassert.NotPanics(t, rlm, f, msgs...) {
		return
	}
	t.FailNow()
}

// PanicsWithMessage requires that the code inside the specified func panics
// locally within the same execution realm.
// Use AbortsWithMessage for requiring panics that cross execution boundaries (aborts).
// See uassert.AbortsWithMessage for `rlm` semantics.
func PanicsWithMessage(t uassert.TestingT, rlm realm, msg string, f any, msgs ...string) {
	t.Helper()
	if uassert.PanicsWithMessage(t, rlm, msg, f, msgs...) {
		return
	}
	t.FailNow()
}

// PanicsContains requires that the code inside the specified func panics
// locally within the same execution realm and the panic message contains the specified substring.
func PanicsContains(t uassert.TestingT, substr string, f any, msgs ...string) {
	t.Helper()
	if uassert.PanicsContains(t, substr, f, msgs...) {
		return
	}
	t.FailNow()
}

// NotPanics requires that the code inside the specified func does NOT panic
// locally within the same execution realm.
// Use NotAborts for requiring the absence of panics that cross execution boundaries (aborts).
// See uassert.AbortsWithMessage for `rlm` semantics.
func NotPanics(t uassert.TestingT, rlm realm, f any, msgs ...string) {
	t.Helper()
	if uassert.NotPanics(t, rlm, f, msgs...) {
		return
	}
	t.FailNow()
}

// Equal requires that two objects are equal.
func Equal(t uassert.TestingT, expected, actual any, msgs ...string) {
	t.Helper()
	if uassert.Equal(t, expected, actual, msgs...) {
		return
	}
	t.FailNow()
}

// NotEqual requires that two objects are not equal.
func NotEqual(t uassert.TestingT, expected, actual any, msgs ...string) {
	t.Helper()
	if uassert.NotEqual(t, expected, actual, msgs...) {
		return
	}
	t.FailNow()
}

// Empty requires that the specified object is empty
// (zero value, empty string/slice/map, or nil).
func Empty(t uassert.TestingT, obj any, msgs ...string) {
	t.Helper()
	if uassert.Empty(t, obj, msgs...) {
		return
	}
	t.FailNow()
}

// NotEmpty requires that the specified object is not empty.
func NotEmpty(t uassert.TestingT, obj any, msgs ...string) {
	t.Helper()
	if uassert.NotEmpty(t, obj, msgs...) {
		return
	}
	t.FailNow()
}

// Nil requires that the value is nil.
func Nil(t uassert.TestingT, value any, msgs ...string) {
	t.Helper()
	if uassert.Nil(t, value, msgs...) {
		return
	}
	t.FailNow()
}

// NotNil requires that the value is not nil.
func NotNil(t uassert.TestingT, value any, msgs ...string) {
	t.Helper()
	if uassert.NotNil(t, value, msgs...) {
		return
	}
	t.FailNow()
}

// TypedNil requires that the value is a typed-nil (nil pointer) value.
func TypedNil(t uassert.TestingT, value any, msgs ...string) {
	t.Helper()
	if uassert.TypedNil(t, value, msgs...) {
		return
	}
	t.FailNow()
}

// NotTypedNil requires that the value is not a typed-nil (nil pointer) value.
func NotTypedNil(t uassert.TestingT, value any, msgs ...string) {
	t.Helper()
	if uassert.NotTypedNil(t, value, msgs...) {
		return
	}
	t.FailNow()
}
