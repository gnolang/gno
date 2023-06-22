package utestify

import (
	"std"
	"testing"
)

func SameStrings(t *testing.T, a, b string) {
	t.Helper()
	if a != b {
		t.Errorf("%q != %q")
	}
}

func SimilarStrings(t *testing.T, a, b string) {
	t.Helper()

	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	if a != b {
		t.Errorf("%q != %q")
	}
}
