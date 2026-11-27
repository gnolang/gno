// Package lendanyother is another realm whose non-crossing callees take a literal.
package lendanyother

import lib "gno.land/p/demo/tests/lendanylib"

// RunAny calls f through a type switch.
func RunAny(f any) {
	switch g := f.(type) {
	case func():
		g()
	}
}

// EachVia hands f on to a library-owned receiver's method.
func EachVia(f func(int)) { lib.Default.Each(f) }
