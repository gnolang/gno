// PKGPATH: gno.land/r/zrealm_lend_perm
package zrealm_lend_perm

import (
	"gno.land/p/demo/tests/lendlib"
	"gno.land/p/demo/tests/lendlib/keep"
)

var count int

var _ = keep.Holder{}

func attack() {
	p := &lendlib.Perms{Allowed: "me"}
	p.WithPermission("me", func() { count++ })
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
