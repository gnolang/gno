// errorcheck -lang=go1.22

//go:build go1.21

// We need a line directive before the package clause,
// but don't change file name or position so that the
// error message appears at the right place.

//line issue67141.go:10
package p

func _() {
	for range 10 { // ERROR "cannot range over 10"
	}
}

// GnoStaticIncomplete: covered 0 of 1 markers (Gno preprocess: 0, go/types guard: 0); Gno's own preprocess flags none (lenient); the rest are caught by neither — a runnable variant may exercise more

// KnownIssue:
// line 12: 0: range iteration requires map, string, array, slice, or pointer to array
