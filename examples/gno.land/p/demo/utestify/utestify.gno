package utestify

import (
	"std"
	"strings"
	"testing"
)

func MustEqualStrings(t *testing.T, a, b string) {
	t.Helper()
	if a != b {
		t.Errorf("%q != %q", a, b)
	}
}

func ShouldEqualStrings(t *testing.T, a, b string) bool {
	t.Helper()
	if a != b {
		t.Logf("%q != %q", a, b)
		return false
	}
	return true
}

func MustResembleStrings(t *testing.T, a, b string) {
	t.Helper()

	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a != b {
		t.Errorf("%q != %q", a, b)
	}
}

func ShouldResembleStrings(t *testing.T, a, b string) bool {
	t.Helper()

	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a != b {
		t.Logf("%q != %q", a, b)
		return false
	}
	return true
}
