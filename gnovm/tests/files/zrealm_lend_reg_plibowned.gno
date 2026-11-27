// PKGPATH: gno.land/r/tests/vm/lendregplibowned
package lendregplibowned

import lib "gno.land/p/demo/tests/lendreglib"

var items []*lib.Item
var count int

// (b) A library-owned receiver's method calls the library's stored closure,
// which allocates; the realm then stores the result. Same output without the loan rule.
func main(cur realm) {
	println("e02", try(func() { items = append(items, lib.MakeVia(8)) }), len(items))
	println("e04 lent literal through plain lib func", try(func() { lib.Call(func() { count++ }) }), count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// e02 undefined 1
// e04 lent literal through plain lib func undefined 1
