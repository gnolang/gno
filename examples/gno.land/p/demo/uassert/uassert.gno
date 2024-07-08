// uassert is an adapted lighter version of https://github.com/stretchr/testify/assert.
package uassert

import (
	"regexp"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
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
		return fail(t, msgs, "func should panic with message:\t%s\n\tPanic value:\t%s", msg, panicValue)
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
			es, as = ev, av
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

// Empty asserts that the specified object is empty.
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

// NotEmpty asserts that the specified object is not empty.
func NotEmpty(t TestingT, obj interface{}, msgs ...string) bool {
	t.Helper()
	switch val := obj.(type) {
	case string:
		if val == "" {
			return fail(t, msgs, "uassert.NotEmpty: string is empty")
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if val == 0 {
			return fail(t, msgs, "uassert.NotEmpty: number is empty: %d", val)
		}
	case std.Address:
		var zeroAddr std.Address
		if val == zeroAddr {
			return fail(t, msgs, "uassert.NotEmpty: std.Address is empty: %s", string(val))
		}
	default:
		return fail(t, msgs, "uassert.NotEmpty: unsupported type")
	}
	return true
}

// NotEqual asserts that two objects are not equal.
func NotEqual(t TestingT, expected, actual interface{}, msgs ...string) bool {
	t.Helper()
	if Equal(t, expected, actual, msgs...) {
		return fail(t, msgs, "uassert.NotEqual: expected %v and actual %v should not be equal", expected, actual)
	}

	return true
}

// Nil asserts that the specified object is nil.
func Nil(t TestingT, obj interface{}, msgs ...string) bool {
	t.Helper()
	if obj != nil {
		return fail(t, msgs, "uassert.Nil: object should be nil")
	}

	return true
}

// NotNil asserts that the specified object is not nil.
func NotNil(t TestingT, obj interface{}, msgs ...string) bool {
	t.Helper()
	if obj == nil {
		return fail(t, msgs, "uassert.NotNil: object should not be nil")
	}

	return true
}

// Contains asserts that the specified string or list contains the specified substring or element.
func Contains(t TestingT, container, element interface{}, msgs ...string) bool {
	t.Helper()
	switch c := container.(type) {
	case string:
		if e, ok := element.(string); ok {
			if !strings.Contains(c, e) {
				return fail(t, msgs, "uassert.Contains: %q does not contain %q", c, e)
			}
			return true
		}
	case []string:
		for _, item := range c {
			if item == element {
				return true
			}
		}
		return fail(t, msgs, "uassert.Contains: %v does not contain %v", container, element)
	default:
		return fail(t, msgs, "uassert.Contains: unsupported type %T", container)
	}

	return false
}

// NotContains asserts that the specified string or list does not contain the specified substring or element.
func NotContains(t TestingT, container, element interface{}, msgs ...string) bool {
	t.Helper()
	switch c := container.(type) {
	case string:
		if e, ok := element.(string); ok {
			if strings.Contains(c, e) {
				return fail(t, msgs, "uassert.NotContains: %q contains %q", c, e)
			}
			return true
		}
	case []string:
		for _, item := range c {
			if item == element {
				return fail(t, msgs, "uassert.NotContains: %v contains %v", container, element)
			}
		}
		return true
	default:
		return fail(t, msgs, "uassert.NotContains: unsupported type %T", container)
	}
	return false
}

// Len asserts that the specified object has specific length.
func Len(t TestingT, obj interface{}, length int, msgs ...string) bool {
	t.Helper()
	actualLen := -1
	switch v := obj.(type) {
	case string:
		actualLen = len(v)
	case []interface{}:
		actualLen = len(v)
	default:
		return fail(t, msgs, "uassert.Len: unsupported type %T", obj)
	}

	if actualLen != length {
		return fail(t, msgs, "uassert.Len: expected length %d but got %d", length, actualLen)
	}
	return true
}

// Greater asserts that the first value is greater than the second.
func Greater(t TestingT, e1, e2 interface{}, msgs ...string) bool {
	t.Helper()
	if !compare(t, e1, e2, ">") {
		return fail(t, msgs, "uassert.Greater: %v is not greater than %v", e1, e2)
	}
	return true
}

// GreaterOrEqual asserts that the first value is greater than or equal to the second.
func GreaterOrEqual(t TestingT, e1, e2 interface{}, msgs ...string) bool {
	t.Helper()
	if !compare(t, e1, e2, ">=") {
		return fail(t, msgs, "uassert.GreaterOrEqual: %v is not greater than or equal to %v", e1, e2)
	}
	return true
}

// Less asserts that the first value is less than the second.
func Less(t TestingT, e1, e2 interface{}, msgs ...string) bool {
	t.Helper()
	if !compare(t, e1, e2, "<") {
		return fail(t, msgs, "uassert.Less: %v is not less than %v", e1, e2)
	}
	return true
}

// LessOrEqual asserts that the first value is less than or equal to the second.
func LessOrEqual(t TestingT, e1, e2 interface{}, msgs ...string) bool {
	t.Helper()
	if !compare(t, e1, e2, "<=") {
		return fail(t, msgs, "uassert.LessOrEqual: %v is not less than or equal to %v", e1, e2)
	}
	return true
}

// IsType asserts that two objects are of the same type.
func IsType(t TestingT, expectedType, actual interface{}, msgs ...string) bool {
	t.Helper()
	switch expectedType.(type) {
	case bool:
		if _, ok := actual.(bool); !ok {
			return fail(t, msgs, "uassert.IsType: expected type bool but got %T", actual)
		}
	case int:
		if _, ok := actual.(int); !ok {
			return fail(t, msgs, "uassert.IsType: expected type int but got %T", actual)
		}
	case int8:
		if _, ok := actual.(int8); !ok {
			return fail(t, msgs, "uassert.IsType: expected type int8 but got %T", actual)
		}
	case int16:
		if _, ok := actual.(int16); !ok {
			return fail(t, msgs, "uassert.IsType: expected type int16 but got %T", actual)
		}
	case int32:
		if _, ok := actual.(int32); !ok {
			return fail(t, msgs, "uassert.IsType: expected type int32 but got %T", actual)
		}
	case int64:
		if _, ok := actual.(int64); !ok {
			return fail(t, msgs, "uassert.IsType: expected type int64 but got %T", actual)
		}
	case uint:
		if _, ok := actual.(uint); !ok {
			return fail(t, msgs, "uassert.IsType: expected type uint but got %T", actual)
		}
	case uint8:
		if _, ok := actual.(uint8); !ok {
			return fail(t, msgs, "uassert.IsType: expected type uint8 but got %T", actual)
		}
	case uint16:
		if _, ok := actual.(uint16); !ok {
			return fail(t, msgs, "uassert.IsType: expected type uint16 but got %T", actual)
		}
	case uint32:
		if _, ok := actual.(uint32); !ok {
			return fail(t, msgs, "uassert.IsType: expected type uint32 but got %T", actual)
		}
	case uint64:
		if _, ok := actual.(uint64); !ok {
			return fail(t, msgs, "uassert.IsType: expected type uint64 but got %T", actual)
		}
	case string:
		if _, ok := actual.(string); !ok {
			return fail(t, msgs, "uassert.IsType: expected type string but got %T", actual)
		}
	case std.Address:
		if _, ok := actual.(std.Address); !ok {
			return fail(t, msgs, "uassert.IsType: expected type std.Address but got %T", actual)
		}
	default:
		return fail(t, msgs, "uassert.IsType: unsupported type %T", expectedType)
	}
	return true
}

// Regexp asserts that a specified regexp matches matches a string.
func Regexp(t TestingT, rx interface{}, str interface{}, msgs ...string) bool {
	t.Helper()
	var r *regexp.Regexp
	switch v := rx.(type) {
	case *regexp.Regexp:
		r = v
	case string:
		r = regexp.MustCompile(v)
	default:
		return fail(t, msgs, "uassert.Regexp: invalid regexp type")
	}

	s, ok := str.(string)
	if !ok {
		return fail(t, msgs, "uassert.Regexp: invalid string type")
	}

	if !r.MatchString(s) {
		return fail(t, msgs, "uassert.Regexp: \"%s\" does not match \"%s\"", s, r.String())
	}
	return true
}

// ElementsMatch asserts that two lists contain the same elements, regardless of order.
func ElementsMatch(t TestingT, listA, listB []interface{}, msgs ...string) bool {
	t.Helper()
	if len(listA) != len(listB) {
		return fail(t, msgs, "uassert.ElementsMatch: lists have different lengths")
	}
	elements := make(map[interface{}]int)
	for _, item := range listA {
		elements[item]++
	}
	for _, item := range listB {
		if count, ok := elements[item]; !ok || count == 0 {
			return fail(t, msgs, "uassert.ElementsMatch: lists do not match")
		} else {
			elements[item]--
		}
	}
	return true
}

// Subset asserts that all elements in the subset are present in the set.
func Subset(t TestingT, set, subset []interface{}, msgs ...string) bool {
	t.Helper()
	elements := make(map[interface{}]bool)
	for _, item := range set {
		elements[item] = true
	}
	for _, item := range subset {
		if !elements[item] {
			return fail(t, msgs, "uassert.Subset: item %v not found in set", item)
		}
	}
	return true
}

// NotSubset asserts that the subset is not fully present in the set.
func NotSubset(t TestingT, set, subset []interface{}, msgs ...string) bool {
	t.Helper()
	elements := make(map[interface{}]bool)
	for _, item := range set {
		elements[item] = true
	}
	for _, item := range subset {
		if !elements[item] {
			return true
		}
	}
	return fail(t, msgs, "uassert.NotSubset: subset is fully present in set")
}

// IsInstanceOf asserts that the specified object is an instance of the given type.
func IsInstanceOf(t TestingT, expectedType, actual interface{}, msgs ...string) bool {
	t.Helper()
	if !IsType(t, expectedType, actual, msgs...) {
		return fail(t, msgs, "uassert.IsInstanceOf: %v is not an instance of %T", actual, expectedType)
	}
	return true
}

// Same asserts that two pointers reference the same object.
func Same(t TestingT, expected, actual interface{}, msgs ...string) bool {
	t.Helper()
	if expected != actual {
		return fail(t, msgs, "uassert.Same: expected and actual do not reference the same object")
	}
	return true
}

// NotSame asserts that two pointers do not reference the same object.
func NotSame(t TestingT, expected, actual interface{}, msgs ...string) bool {
	t.Helper()
	if expected == actual {
		return fail(t, msgs, "uassert.NotSame: expected and actual reference the same object")
	}
	return true
}

// Panics asserts that the code inside the specified func panics.
func Panics(t TestingT, f func(), msgs ...string) bool {
	t.Helper()

	didPanic, _ := checkDidPanic(f)
	if !didPanic {
		return fail(t, msgs, "uassert.Panics: function did not panic")
	}
	return true
}

// PanicsWithValue asserts that the code inside the specified func panics,
// and that the recovered panic value equals the specified value.
func PanicsWithValue(t TestingT, expected interface{}, f func(), msgs ...string) bool {
	t.Helper()

	didPanic, panicValue := checkDidPanic(f)
	if !didPanic {
		return fail(t, msgs, "uassert.PanicsWithValue: function did not panic")
	}

	if panicValue != expected {
		return fail(t, msgs, "uassert.PanicsWithValue: panic value %v does not equal expected %v", panicValue, expected)
	}
	return true
}
