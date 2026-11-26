// PKGPATH: gno.land/r/zrealm_t51_vslot
package zrealm_t51_vslot

import (
	"gno.land/p/demo/tests/t51lib"
	"gno.land/p/demo/tests/t51lib/keep"
)

var count int

var h *keep.Holder

func attack() {
	h = t51lib.LendV(func() { count++ })
}

func later() {
	h = t51lib.KeepV(func() { count++ })
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
// refused: gno.land/r/zrealm_t51_vslot's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t51_vslot) package{})).count
// count: 0
// later refused: gno.land/r/zrealm_t51_vslot's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t51_vslot) package{})).count
// count: 0
