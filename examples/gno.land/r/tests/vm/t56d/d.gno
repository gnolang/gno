package t56d

import "gno.land/p/demo/tests/t56lib"

var count int

func Count() int { return count }

// Both writes through a local helper and a lent literal.
func Both(cur realm) int {
	f := func() { count++ }
	f()
	t56lib.Run(func() { count++ })
	return count
}

// Neither repeats Both's shapes without the marks: both writes are refused.
func Neither(cur realm) (helperRefused, lendRefused bool, n int) {
	func() {
		defer func() { helperRefused = recover() != nil }()
		var f func()
		f = func() { count++ }
		f = func() { count += 100 }
		f()
	}()
	func() {
		defer func() { lendRefused = recover() != nil }()
		w := func() { count++ }
		t56lib.Run(w)
	}()
	return helperRefused, lendRefused, count
}
