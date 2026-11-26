// PKGPATH: gno.land/r/zrealm_t57_any_ok
package zrealm_t57_any_ok

import lib "gno.land/p/demo/tests/t57lib"

var count int

type fn func()

// A literal passed as `any` to a callee that only type-switches or asserts it
// and calls the result is lent like any other in-place literal.
func runConv(f any) {
	switch g := f.(type) {
	case fn:
		call(g)
	case func():
		call(fn(g))
	}
}

func call(g fn) { g() }

func main(cur realm) {
	lib.RunAny(func() { count++ })
	lib.CallAny(func() { count++ })
	runConv(func() { count++ })
	println("count:", count)
}

// Output:
// count: 3
