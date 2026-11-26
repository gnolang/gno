// PKGPATH: gno.land/r/zrealm_t51_rebind
package zrealm_t51_rebind

import (
	"gno.land/p/demo/tests/t51lib"
	"gno.land/p/demo/tests/t51lib/keep"
)

var count int

var _ = keep.Holder{}

func attack() {
	h := func() {}
	t51lib.Rebind(func(g func()) { h = g }, func() { count++ })
	h()
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
// refused: gno.land/r/zrealm_t51_rebind's code is read-only here: a function value writes only as a literal passed in place or called where it is written (otherwise give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t51_rebind) package{})).count
// count: 0
