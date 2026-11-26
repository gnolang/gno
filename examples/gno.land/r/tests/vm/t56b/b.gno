package t56b

import "gno.land/p/demo/tests/t56lib"

var count int

func Count() int { return count }

// Lend writes through a literal lent to t56lib.Run (a marked parameter).
func Lend(cur realm) int {
	t56lib.Run(func() { count++ })
	return count
}

// NotLent passes the literal through a variable: no loan, the write is refused.
func NotLent(cur realm) (refused bool, n int) {
	func() {
		defer func() { refused = recover() != nil }()
		w := func() { count++ }
		t56lib.Run(w)
	}()
	return refused, count
}
