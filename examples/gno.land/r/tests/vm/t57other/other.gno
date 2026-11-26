// Package t57other is another realm whose non-crossing callees take a literal.
package t57other

import lib "gno.land/p/demo/tests/t57lib"

// RunAny calls f through a type switch.
func RunAny(f any) {
	switch g := f.(type) {
	case func():
		g()
	}
}

// EachVia hands f on to a library-owned receiver's method.
func EachVia(f func(int)) { lib.Default.Each(f) }
