package lendrun

import "gno.land/p/demo/tests/lendrunlib"

var count int

func Count() int { return count }

// Lend writes through a literal lent to lendrunlib.Run (a marked parameter).
func Lend(cur realm) int {
	lendrunlib.Run(func() { count++ })
	return count
}

// NotLent passes the literal through a variable: no loan, the write is refused.
func NotLent(cur realm) (refused bool, n int) {
	func() {
		defer func() { refused = recover() != nil }()
		w := func() { count++ }
		lendrunlib.Run(w)
	}()
	return refused, count
}
