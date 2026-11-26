// PKGPATH: gno.land/r/zrealm_t51_spread2
package zrealm_t51_spread2

import (
	"gno.land/p/demo/tests/t51lib"
	"gno.land/p/demo/tests/t51lib/keep"
)

var count int

var _ = keep.Holder{}

func attack() {
	var none []func()
	t51lib.RunS(func() { count++ }, none...)
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
// refused: undefined
// count: 1
