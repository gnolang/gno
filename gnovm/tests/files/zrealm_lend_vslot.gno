// PKGPATH: gno.land/r/zrealm_lend_vslot
package zrealm_lend_vslot

import (
	"gno.land/p/demo/tests/lendlib"
	"gno.land/p/demo/tests/lendlib/keep"
)

var count int

var h *keep.Holder

func attack() {
	h = lendlib.LendV(func() { count++ })
}

func later() {
	h = lendlib.KeepV(func() { count++ })
	h.Fire()
}

func attempt() (r any) {
	defer func() { r = recover() }()
	attack()
	return nil
}

func attemptLater() (r any) {
	defer func() { r = recover() }()
	later()
	return nil
}

func main(cur realm) {
	println("refused:", attempt())
	println("count:", count)
	println("later refused:", attemptLater())
	println("count:", count)
}

// Output:
// refused: gno.land/r/zrealm_lend_vslot's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_vslot) package{})).count
// count: 0
// later refused: gno.land/r/zrealm_lend_vslot's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_lend_vslot) package{})).count
// count: 0
