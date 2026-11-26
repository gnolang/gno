// PKGPATH: gno.land/r/zrealm_lend_each
package zrealm_lend_each

import (
	"gno.land/p/demo/tests/lendlib"
	"gno.land/p/demo/tests/lendlib/keep"
)

var count int

var _ = keep.Holder{}

func attack() {
	lendlib.Each([]int{1, 2, 3}, func(x int) { count += x })
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
// count: 6
