// PKGPATH: gno.land/r/zrealm_lend_other
package zrealm_lend_other

import (
	"gno.land/r/tests/vm/lendvictim"
)

func main(cur realm) {
	println("before:", lendvictim.Count())
	lendvictim.UseOther(cross(cur))
	println("after:", lendvictim.Count())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/lendvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendvictim) package{})).count
