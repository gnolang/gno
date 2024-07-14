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

func PanicsWithMessage(t uassert.TestingT, msg string, f func(), msgs ...string) {
	t.Helper()
	if uassert.PanicsWithMessage(t, msg, f, msgs...) {
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

func NotEqual(t uassert.TestingT, expected, actual interface{}, msgs ...string) {
	t.Helper()
	if uassert.NotEqual(t, expected, actual, msgs...) {
		return
	}
	t.FailNow()
}

func Empty(t uassert.TestingT, obj interface{}, msgs ...string) {
	t.Helper()
	if uassert.Empty(t, obj, msgs...) {
		return
	}
	t.FailNow()
}

func NotEmpty(t uassert.TestingT, obj interface{}, msgs ...string) {
	t.Helper()
	if uassert.NotEmpty(t, obj, msgs...) {
		return
	}
	t.FailNow()
}

func Nil(t uassert.TestingT, obj interface{}, msgs ...string) {
	t.Helper()
	if uassert.Nil(t, obj, msgs...) {
		return
	}
	t.FailNow()
}

func NotNil(t uassert.TestingT, obj interface{}, msgs ...string) {
	t.Helper()
	if uassert.NotNil(t, obj, msgs...) {
		return
	}
	t.FailNow()
}

func Contains(t uassert.TestingT, container, element interface{}, msgs ...string) {
	t.Helper()
	if uassert.Contains(t, container, element, msgs...) {
		return
	}
	t.FailNow()
}

func NotContains(t uassert.TestingT, container, element interface{}, msgs ...string) {
	t.Helper()
	if uassert.NotContains(t, container, element, msgs...) {
		return
	}
	t.FailNow()
}

func Len(t uassert.TestingT, obj interface{}, length int, msgs ...string) {
	t.Helper()
	if uassert.Len(t, obj, length, msgs...) {
		return
	}
	t.FailNow()
}

func Greater(t uassert.TestingT, e1, e2 interface{}, msgs ...string) {
	t.Helper()
	if uassert.Greater(t, e1, e2, msgs...) {
		return
	}
	t.FailNow()
}

func GreaterOrEqual(t uassert.TestingT, e1, e2 interface{}, msgs ...string) {
	t.Helper()
	if uassert.GreaterOrEqual(t, e1, e2, msgs...) {
		return
	}
	t.FailNow()
}

func Less(t uassert.TestingT, e1, e2 interface{}, msgs ...string) {
	t.Helper()
	if uassert.Less(t, e1, e2, msgs...) {
		return
	}
	t.FailNow()
}

func LessOrEqual(t uassert.TestingT, e1, e2 interface{}, msgs ...string) {
	t.Helper()
	if uassert.LessOrEqual(t, e1, e2, msgs...) {
		return
	}
	t.FailNow()
}

func IsType(t uassert.TestingT, expectedType, actual interface{}, msgs ...string) {
	t.Helper()
	if uassert.IsType(t, expectedType, actual, msgs...) {
		return
	}
	t.FailNow()
}

func Regexp(t uassert.TestingT, rx, str interface{}, msgs ...string) {
	t.Helper()
	if uassert.Regexp(t, rx, str, msgs...) {
		return
	}
	t.FailNow()
}

func ElementsMatch(t uassert.TestingT, listA, listB []interface{}, msgs ...string) {
	t.Helper()
	if uassert.ElementsMatch(t, listA, listB, msgs...) {
		return
	}
	t.FailNow()
}

func Subset(t uassert.TestingT, set, subset []interface{}, msgs ...string) {
	t.Helper()
	if uassert.Subset(t, set, subset, msgs...) {
		return
	}
	t.FailNow()
}

func NotSubset(t uassert.TestingT, set, subset []interface{}, msgs ...string) {
	t.Helper()
	if uassert.NotSubset(t, set, subset, msgs...) {
		return
	}
	t.FailNow()
}

func IsInstanceOf(t uassert.TestingT, expectedType, actual interface{}, msgs ...string) {
	t.Helper()
	if uassert.IsInstanceOf(t, expectedType, actual, msgs...) {
		return
	}
	t.FailNow()
}

func Same(t uassert.TestingT, expected, actual interface{}, msgs ...string) {
	t.Helper()
	if uassert.Same(t, expected, actual, msgs...) {
		return
	}
	t.FailNow()
}

func NotSame(t uassert.TestingT, expected, actual interface{}, msgs ...string) {
	t.Helper()
	if uassert.NotSame(t, expected, actual, msgs...) {
		return
	}
	t.FailNow()
}

func Panics(t uassert.TestingT, f func(), msgs ...string) {
	t.Helper()
	if uassert.Panics(t, f, msgs...) {
		return
	}
	t.FailNow()
}

func PanicsWithValue(t uassert.TestingT, expected interface{}, f func(), msgs ...string) {
	t.Helper()
	if uassert.PanicsWithValue(t, expected, f, msgs...) {
		return
	}
	t.FailNow()
}
