// PKGPATH: gno.land/r/zrealm_lend_iterate
package zrealm_lend_iterate

import (
	"gno.land/p/demo/tests/lendlib"
	"gno.land/p/demo/tests/lendlib/keep"
)

var count int

var _ = keep.Holder{}

func attack() {
	t := &lendlib.Tree{Keys: []string{"a", "b", "c"}}
	t.Iterate(func(k string) bool {
		count++
		return false
	})
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
// count: 3
