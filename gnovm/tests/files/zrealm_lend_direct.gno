// PKGPATH: gno.land/r/zrealm_lend_direct
package zrealm_lend_direct

import (
	"gno.land/r/tests/vm/lendvictim"
)

func main(cur realm) {
	println("before:", lendvictim.Count())
	lendvictim.UseDirect(cross(cur))
	println("after:", lendvictim.Count())
}

// Output:
// before: 0
// after: 1
