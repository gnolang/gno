// uassert is an adapted lighter version of https://github.com/stretchr/testify/assert.
package uassert

import (
	"std"
	"strconv"
	"strings"
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

// PanicsWithMessage asserts that the code inside the specified func panics,
// and that the recovered panic value satisfies the given message
func PanicsWithMessage(t TestingT, msg string, f func(), msgs ...string) bool {
	t.Helper()

	didPanic, panicValue := checkDidPanic(f)
	if !didPanic {
		return fail(t, msgs, "func should panic\n\tPanic value:\t%v", panicValue)
	}

	if panicValue != msg {
		return fail(t, msgs, "func should panic with message:\t%s\n\tPanic value:\t%s", panicValue, msg)
	}
	return true
}

// NotPanics asserts that the code inside the specified func does NOT panic.
func NotPanics(t TestingT, f func(), msgs ...string) bool {
	t.Helper()

	didPanic, panicValue := checkDidPanic(f)

	if didPanic {
		return fail(t, msgs, "func should not panic\n\tPanic value:\t%s", panicValue)
	}
	return true
}

// Equal asserts that two objects are equal.
func Equal(t TestingT, expected, actual interface{}, msgs ...string) bool {
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
			es, as = ev, as
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

func Empty(t TestingT, obj interface{}, msgs ...string) bool {
	t.Helper()
	switch val := obj.(type) {
	case string:
		if val != "" {
			return fail(t, msgs, "uassert.Empty: not empty string: %s", val)
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if val != 0 {
			return fail(t, msgs, "uassert.Empty: not empty number: %d", val)
		}
	case std.Address:
		var zeroAddr std.Address
		if val != zeroAddr {
			return fail(t, msgs, "uassert.Empty: not empty std.Address: %s", string(val))
		}
	default:
		return fail(t, msgs, "uassert.Empty: unsupported type")
	}
	return true
}
