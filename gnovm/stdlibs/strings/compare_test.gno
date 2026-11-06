// Ported from upstream Go's compare_test.go at go1.25.9:
// https://github.com/golang/go/blob/go1.25.9/src/strings/compare_test.go

package strings_test

import (
	"strings"
	"testing"
)

var compareTests = []struct {
	a, b string
	i    int
}{
	{"", "", 0},
	{"a", "", 1},
	{"", "a", -1},
	{"abc", "abc", 0},
	{"ab", "abc", -1},
	{"abc", "ab", 1},
	{"x", "ab", 1},
	{"ab", "x", -1},
	{"x", "a", 1},
	{"b", "x", -1},
	{"abcdefgh", "abcdefgh", 0},
	{"abcdefghi", "abcdefghi", 0},
	{"abcdefghi", "abcdefghj", -1},
}

func TestCompare(t *testing.T) {
	for _, tt := range compareTests {
		numShifts := 16
		for offset := 0; offset <= numShifts; offset++ {
			shiftedB := (strings.Repeat("*", offset) + tt.b)[offset:]
			cmp := strings.Compare(tt.a, shiftedB)
			if cmp != tt.i {
				t.Errorf("Compare(%q, %q), offset %d = %v; want %v", tt.a, tt.b, offset, cmp, tt.i)
			}
		}
	}
}

func TestCompareIdenticalString(t *testing.T) {
	var s = "Hello Gophers!"
	if strings.Compare(s, s) != 0 {
		t.Error("s != s")
	}
	if strings.Compare(s, s[:1]) != 1 {
		t.Error("s > s[:1] failed")
	}
}
