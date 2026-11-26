// PKGPATH: gno.land/r/zrealm_t51_direct
package zrealm_t51_direct

import (
	"gno.land/r/tests/vm/t51v"
)

func main(cur realm) {
	println("before:", t51v.Count())
	t51v.UseDirect(cross(cur))
	println("after:", t51v.Count())
}

// Output:
// before: 0
// after: 1
