// PKGPATH: gno.land/r/zrealm_lend_panic
package zrealm_lend_panic

import (
	"gno.land/p/demo/tests/lendlib"
	"gno.land/p/demo/tests/lendlib/keep"
)

var count int

var _ = keep.Holder{}

func attack() {
	defer func() {
		g := recover().(func())
		g()
	}()
	lendlib.Throw(func() { count++ })
}

func attempt() (r any) {
	defer func() { r = recover() }()
	attack()
	return nil
}

func main(cur realm) {
	println("refused:", attempt())
	println("count:", count)
}

// Output:
// refused: gno.land/r/zrealm_lend_panic's code is read-only here: a function value writes only as a literal called where it is written, or passed in place to a plain function of this realm or a /p/ function called by name, while that call runs (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_panic) package{})).count
// count: 0
