package uassert

import (
	"std"
	"strings"
	"testing"
)

type TestingT interface {
	Errorf(fmt string, args ...interface{})
}

func MustEqualStrings(t TestingT, a, b string) {
	t.Helper()
	_ = checkEqualStrings(t, a, b, t.Errorf)
}

func EqualStrings(t TestingT, a, b string) bool {
	t.Helper()
	return checkEqualStrings(t, a, b, t.Logf)
}

func checkEqualStrings(t TestingT, a, b string, cb printerFn) bool {
	if a != b {
		cb("%q != %q", a, b)
		return false
	}
	return true
}

func MustResembleStrings(t TestingT, a, b string) {
	t.Helper()

	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a != b {
		t.Errorf("%q != %q", a, b)
	}
}

func ShouldResembleStrings(t TestingT, a, b string) bool {
	t.Helper()

	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a != b {
		t.Logf("%q != %q", a, b)
		return false
	}
	return true
}
