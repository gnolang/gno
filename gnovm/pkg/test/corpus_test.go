package test

import "testing"

func TestCleanErrorMessage(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ in, want string }{
		{"main/main.gno:9:2: invalid recursive type", "invalid recursive type"},
		{"main/main.gno:9:2-11: cannot use x", "cannot use x"},
		// Multi-line span: the end position must not leak into the message.
		{"gno.land/p/filetest/p/main.go:9:2-11:2: invalid recursive type: I1 -> I2 -> I1", "invalid recursive type: I1 -> I2 -> I1"},
		{"main/main.gno:10:3-12:3'': cannot convert b", "cannot convert b"},
		{"main/main.gno:4:5': primed point", "primed point"},
		{"./main.go:5:3: undefined: y", "undefined: y"},
		{"main.go:5: line only", "line only"},
		{"no position here: keep it", "no position here: keep it"},
		{"  main.gno:1:1:   padded  ", "padded"},
	} {
		if got := CleanErrorMessage(tc.in); got != tc.want {
			t.Errorf("CleanErrorMessage(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}
