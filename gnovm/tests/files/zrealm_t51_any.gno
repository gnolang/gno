// PKGPATH: gno.land/r/zrealm_t51_any
package zrealm_t51_any

import (
	"gno.land/p/demo/tests/t51lib"
	"gno.land/p/demo/tests/t51lib/keep"
)

var count int

var _ = keep.Holder{}

func attack() {
	g := t51lib.Any(func() { count++ }).(func())
	g()
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
// refused: gno.land/r/zrealm_t51_any's code is read-only here: a function value writes only in the crossing call that made it (to write elsewhere, give it a cur realm parameter and call it with cross): (const (ref(gno.land/r/zrealm_t51_any) package{})).count
// count: 0
