// PKGPATH: gno.land/r/zrealm_lend_run
package zrealm_lend_run

import (
	"gno.land/p/demo/tests/lendlib"
	"gno.land/p/demo/tests/lendlib/keep"
)

var count int

var _ = keep.Holder{}

func attack() {
	lendlib.Run(func() { count++ })
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
