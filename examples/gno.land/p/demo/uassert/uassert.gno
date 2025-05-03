// uassert is an adapted lighter version of https://github.com/stretchr/testify/assert.
package uassert

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/diff"
)

// NoError asserts that a function returned no error (i.e. `nil`).
func NoError(t TestingT, err error, msgs ...string) bool {
	t.Helper()
	if err != nil {
		return fail(t, msgs, "unexpected error: %s", err.Error())
	}
	return true
}

// Error asserts that a function returned an error (i.e. not `nil`).
func Error(t TestingT, err error, msgs ...string) bool {
	t.Helper()
	if err == nil {
		return fail(t, msgs, "an error is expected but got nil")
	}
	return true
}

// ErrorContains asserts that a function returned an error (i.e. not `nil`)
// and that the error contains the specified substring.
func ErrorContains(t TestingT, err error, contains string, msgs ...string) bool {
	t.Helper()

	if !Error(t, err, msgs...) {
		return false
	}

	actual := err.Error()
	if !strings.Contains(actual, contains) {
		return fail(t, msgs, "error %q does not contain %q", actual, contains)
	}

	return true
}

// True asserts that the specified value is true.
func True(t TestingT, value bool, msgs ...string) bool {
	t.Helper()
	if !value {
		return fail(t, msgs, "should be true")
	}
	return true
}

// False asserts that the specified value is false.
func False(t TestingT, value bool, msgs ...string) bool {
	t.Helper()
	if value {
		return fail(t, msgs, "should be false")
	}
	return true
}

// ErrorIs asserts the given error matches the target error
func ErrorIs(t TestingT, err, target error, msgs ...string) bool {
	t.Helper()

	if err == nil || target == nil {
		return err == target
	}

	// XXX: if errors.Is(err, target) return true

	if err.Error() != target.Error() {
		return fail(t, msgs, "error mismatch, expected %s, got %s", target.Error(), err.Error())
	}

	return true
}

// AbortsWithMessage asserts that the code inside the specified func aborts,
// and that the revived abort value satisfies the given message.
// Note: This relies on gno's `abort` and `revive` mechanism.
func AbortsWithMessage(t TestingT, msg string, f func(), msgs ...string) bool {
	t.Helper()

	var didAbort bool
	var abortValue any

	//r := revive(f) // revive() captures the value passed to panic()
	r := revive(func() { println("FOO") })
	if r != nil {
		didAbort = true
		abortValue = r
	}

	if !didAbort {
		// If the function didn't abort as expected
		return fail(t, msgs, "func should abort")
	}

	// Check if the abort value matches the expected message string
	abortStr, ok := abortValue.(string)
	if !ok {
		// Fail if the abort value wasn't a string
		return fail(t, msgs, "panic value was not a string: got type %T, value %v", abortValue, abortValue)
	}

	if abortStr != msg {
		// Fail if the abort message doesn't match
		return fail(t, msgs, "func should panic with message:\t%q\n\tActual panic value:\t%q", msg, abortStr)
	}

	// Success: function aborted with the expected message
	return true
}

// NotAborts asserts that the code inside the specified func does NOT abort.
// Note: This relies on Gno's `abort` and `revive` mechanism.
func NotAborts(t TestingT, f func(), msgs ...string) bool {
	t.Helper()

	var didAbort bool
	var abortValue any

	//r := revive(f) // revive() captures the value passed to panic()
	r := revive(func() { println("FOO") })
	if r != nil {
		didAbort = true
		abortValue = r
	}

	if didAbort {
		// Fail if the function aborted when it shouldn't have
		// Attempt to format the abort value in the error message
		return fail(t, msgs, "func should not abort\n\tAbort value:\t%v", abortValue)
	}

	// Success: function did not abort
	return true
}

// Equal asserts that two objects are equal.
func Equal(t TestingT, expected, actual any, msgs ...string) bool {
	t.Helper()

	if expected == nil || actual == nil {
		return expected == actual
	}

	// XXX: errors
	// XXX: slices
	// XXX: pointers

	equal := false
	ok_ := false
	es, as := "unsupported type", "unsupported type"

	switch ev := expected.(type) {
	case string:
		if av, ok := actual.(string); ok {
			equal = ev == av
			ok_ = true
			es, as = ev, av
			if !equal {
				dif := diff.MyersDiff(ev, av)
				return fail(t, msgs, "uassert.Equal: strings are different\n\tDiff: %s", diff.Format(dif))
			}
		}
	case std.Address:
		if av, ok := actual.(std.Address); ok {
			equal = ev == av
			ok_ = true
			es, as = string(ev), string(av)
		}
	case int:
		if av, ok := actual.(int); ok {
			equal = ev == av
			ok_ = true
			es, as = strconv.Itoa(ev), strconv.Itoa(av)
		}
	case int8:
		if av, ok := actual.(int8); ok {
			equal = ev == av
			ok_ = true
			es, as = strconv.Itoa(int(ev)), strconv.Itoa(int(av))
		}
	case int16:
		if av, ok := actual.(int16); ok {
			equal = ev == av
			ok_ = true
			es, as = strconv.Itoa(int(ev)), strconv.Itoa(int(av))
		}
	case int32:
		if av, ok := actual.(int32); ok {
			equal = ev == av
			ok_ = true
			es, as = strconv.Itoa(int(ev)), strconv.Itoa(int(av))
		}
	case int64:
		if av, ok := actual.(int64); ok {
			equal = ev == av
			ok_ = true
			es, as = strconv.Itoa(int(ev)), strconv.Itoa(int(av))
		}
	case uint:
		if av, ok := actual.(uint); ok {
			equal = ev == av
			ok_ = true
			es, as = strconv.FormatUint(uint64(ev), 10), strconv.FormatUint(uint64(av), 10)
		}
	case uint8:
		if av, ok := actual.(uint8); ok {
			equal = ev == av
			ok_ = true
			es, as = strconv.FormatUint(uint64(ev), 10), strconv.FormatUint(uint64(av), 10)
		}
	case uint16:
		if av, ok := actual.(uint16); ok {
			equal = ev == av
			ok_ = true
			es, as = strconv.FormatUint(uint64(ev), 10), strconv.FormatUint(uint64(av), 10)
		}
	case uint32:
		if av, ok := actual.(uint32); ok {
			equal = ev == av
			ok_ = true
			es, as = strconv.FormatUint(uint64(ev), 10), strconv.FormatUint(uint64(av), 10)
		}
	case uint64:
		if av, ok := actual.(uint64); ok {
			equal = ev == av
			ok_ = true
			es, as = strconv.FormatUint(ev, 10), strconv.FormatUint(av, 10)
		}
	case bool:
		if av, ok := actual.(bool); ok {
			equal = ev == av
			ok_ = true
			if ev {
				es, as = "true", "false"
			} else {
				es, as = "false", "true"
			}
		}
	case float32:
		if av, ok := actual.(float32); ok {
			equal = ev == av
			ok_ = true
		}
	case float64:
		if av, ok := actual.(float64); ok {
			equal = ev == av
			ok_ = true
		}
	default:
		return fail(t, msgs, "uassert.Equal: unsupported type")
	}

	/*
		// XXX: implement stringer and other well known similar interfaces
		type stringer interface{ String() string }
		if ev, ok := expected.(stringer); ok {
			if av, ok := actual.(stringer); ok {
				equal = ev.String() == av.String()
				ok_ = true
			}
		}
	*/

	if !ok_ {
		return fail(t, msgs, "uassert.Equal: different types") // XXX: display the types
	}
	if !equal {
		return fail(t, msgs, "uassert.Equal: same type but different value\n\texpected: %s\n\tactual:   %s", es, as)
	}

	return true
}

// NotEqual asserts that two objects are not equal.
func NotEqual(t TestingT, expected, actual any, msgs ...string) bool {
	t.Helper()

	if expected == nil || actual == nil {
		return expected != actual
	}

	// XXX: errors
	// XXX: slices
	// XXX: pointers

	notEqual := false
	ok_ := false
	es, as := "unsupported type", "unsupported type"

	switch ev := expected.(type) {
	case string:
		if av, ok := actual.(string); ok {
			notEqual = ev != av
			ok_ = true
			es, as = ev, av
		}
	case std.Address:
		if av, ok := actual.(std.Address); ok {
			notEqual = ev != av
			ok_ = true
			es, as = string(ev), string(av)
		}
	case int:
		if av, ok := actual.(int); ok {
			notEqual = ev != av
			ok_ = true
			es, as = strconv.Itoa(ev), strconv.Itoa(av)
		}
	case int8:
		if av, ok := actual.(int8); ok {
			notEqual = ev != av
			ok_ = true
			es, as = strconv.Itoa(int(ev)), strconv.Itoa(int(av))
		}
	case int16:
		if av, ok := actual.(int16); ok {
			notEqual = ev != av
			ok_ = true
			es, as = strconv.Itoa(int(ev)), strconv.Itoa(int(av))
		}
	case int32:
		if av, ok := actual.(int32); ok {
			notEqual = ev != av
			ok_ = true
			es, as = strconv.Itoa(int(ev)), strconv.Itoa(int(av))
		}
	case int64:
		if av, ok := actual.(int64); ok {
			notEqual = ev != av
			ok_ = true
			es, as = strconv.Itoa(int(ev)), strconv.Itoa(int(av))
		}
	case uint:
		if av, ok := actual.(uint); ok {
			notEqual = ev != av
			ok_ = true
			es, as = strconv.FormatUint(uint64(ev), 10), strconv.FormatUint(uint64(av), 10)
		}
	case uint8:
		if av, ok := actual.(uint8); ok {
			notEqual = ev != av
			ok_ = true
			es, as = strconv.FormatUint(uint64(ev), 10), strconv.FormatUint(uint64(av), 10)
		}
	case uint16:
		if av, ok := actual.(uint16); ok {
			notEqual = ev != av
			ok_ = true
			es, as = strconv.FormatUint(uint64(ev), 10), strconv.FormatUint(uint64(av), 10)
		}
	case uint32:
		if av, ok := actual.(uint32); ok {
			notEqual = ev != av
			ok_ = true
			es, as = strconv.FormatUint(uint64(ev), 10), strconv.FormatUint(uint64(av), 10)
		}
	case uint64:
		if av, ok := actual.(uint64); ok {
			notEqual = ev != av
			ok_ = true
			es, as = strconv.FormatUint(ev, 10), strconv.FormatUint(av, 10)
		}
	case bool:
		if av, ok := actual.(bool); ok {
			notEqual = ev != av
			ok_ = true
			if ev {
				es, as = "true", "false"
			} else {
				es, as = "false", "true"
			}
		}
	case float32:
		if av, ok := actual.(float32); ok {
			notEqual = ev != av
			ok_ = true
		}
	case float64:
		if av, ok := actual.(float64); ok {
			notEqual = ev != av
			ok_ = true
		}
	default:
		return fail(t, msgs, "uassert.NotEqual: unsupported type")
	}

	/*
		// XXX: implement stringer and other well known similar interfaces
		type stringer interface{ String() string }
		if ev, ok := expected.(stringer); ok {
			if av, ok := actual.(stringer); ok {
				notEqual = ev.String() != av.String()
				ok_ = true
			}
		}
	*/

	if !ok_ {
		return fail(t, msgs, "uassert.NotEqual: different types") // XXX: display the types
	}
	if !notEqual {
		return fail(t, msgs, "uassert.NotEqual: same type and same value\n\texpected: %s\n\tactual:   %s", es, as)
	}

	return true
}

func isNumberEmpty(n any) (isNumber, isEmpty bool) {
	switch n := n.(type) {
	// NOTE: the cases are split individually, so that n becomes of the
	// asserted type; the type of '0' was correctly inferred and converted
	// to the corresponding type, int, int8, etc.
	case int:
		return true, n == 0
	case int8:
		return true, n == 0
	case int16:
		return true, n == 0
	case int32:
		return true, n == 0
	case int64:
		return true, n == 0
	case uint:
		return true, n == 0
	case uint8:
		return true, n == 0
	case uint16:
		return true, n == 0
	case uint32:
		return true, n == 0
	case uint64:
		return true, n == 0
	case float32:
		return true, n == 0
	case float64:
		return true, n == 0
	}
	return false, false
}

func Empty(t TestingT, obj any, msgs ...string) bool {
	t.Helper()

	isNumber, isEmpty := isNumberEmpty(obj)
	if isNumber {
		if !isEmpty {
			return fail(t, msgs, "uassert.Empty: not empty number: %d", obj)
		}
	} else {
		switch val := obj.(type) {
		case string:
			if val != "" {
				return fail(t, msgs, "uassert.Empty: not empty string: %s", val)
			}
		case std.Address:
			var zeroAddr std.Address
			if val != zeroAddr {
				return fail(t, msgs, "uassert.Empty: not empty std.Address: %s", string(val))
			}
		default:
			return fail(t, msgs, "uassert.Empty: unsupported type")
		}
	}
	return true
}

func NotEmpty(t TestingT, obj any, msgs ...string) bool {
	t.Helper()
	isNumber, isEmpty := isNumberEmpty(obj)
	if isNumber {
		if isEmpty {
			return fail(t, msgs, "uassert.NotEmpty: empty number: %d", obj)
		}
	} else {
		switch val := obj.(type) {
		case string:
			if val == "" {
				return fail(t, msgs, "uassert.NotEmpty: empty string: %s", val)
			}
		case std.Address:
			var zeroAddr std.Address
			if val == zeroAddr {
				return fail(t, msgs, "uassert.NotEmpty: empty std.Address: %s", string(val))
			}
		default:
			return fail(t, msgs, "uassert.NotEmpty: unsupported type")
		}
	}
	return true
}

// Nil asserts that the value is nil.
func Nil(t TestingT, value any, msgs ...string) bool {
	t.Helper()
	if value != nil {
		return fail(t, msgs, "should be nil")
	}
	return true
}

// NotNil asserts that the value is not nil.
func NotNil(t TestingT, value any, msgs ...string) bool {
	t.Helper()
	if value == nil {
		return fail(t, msgs, "should not be nil")
	}
	return true
}

// TypedNil asserts that the value is a typed-nil (nil pointer) value.
func TypedNil(t TestingT, value any, msgs ...string) bool {
	t.Helper()
	if value == nil {
		return fail(t, msgs, "should be typed-nil but got nil instead")
	}
	if !istypednil(value) {
		return fail(t, msgs, "should be typed-nil")
	}
	return true
}

// NotTypedNil asserts that the value is not a typed-nil (nil pointer) value.
func NotTypedNil(t TestingT, value any, msgs ...string) bool {
	t.Helper()
	if istypednil(value) {
		return fail(t, msgs, "should not be typed-nil")
	}
	return true
}
