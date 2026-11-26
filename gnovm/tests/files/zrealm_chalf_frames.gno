// PKGPATH: gno.land/r/zrealm_chalf_frames
package zrealm_chalf_frames

import lib "gno.land/p/demo/tests/chalflib"

var count int

// Legit: writes while the lending frame is live, across panics and recursion.
func main(cur realm) {
	println("defer-panic:", try(func() { lib.DeferAfterPanic(func() { count++ }) }), count)
	println("recover-then:", try(func() { lib.RecoverThenCall(func() { count++ }) }), count)
	println("rec:", try(func() { lib.Rec(50, func() { count++ }) }), count)
	for i := 0; i < 1000; i++ {
		lib.Rec(2, func() { count++ })
	}
	println("count:", count)
}

func try(fn func()) (r any) {
	defer func() { r = recover() }()
	fn()
	return nil
}

// Output:
// defer-panic: x 1
// recover-then: undefined 3
// rec: undefined 54
// count: 3054
