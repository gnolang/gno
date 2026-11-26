// PKGPATH: gno.land/r/zrealm_t51_panic
package zrealm_t51_panic

import (
	"gno.land/p/demo/tests/t51lib"
	"gno.land/p/demo/tests/t51lib/keep"
)

var count int

var _ = keep.Holder{}

func attack() {
	defer func() {
		g := recover().(func())
		g()
	}()
	t51lib.Throw(func() { count++ })
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
// refused: gno.land/r/zrealm_t51_panic's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t51_panic) package{})).count
// count: 0
