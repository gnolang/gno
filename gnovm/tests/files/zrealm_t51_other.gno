// PKGPATH: gno.land/r/zrealm_t51_other
package zrealm_t51_other

import (
	"gno.land/r/tests/vm/t51v"
)

func main(cur realm) {
	println("before:", t51v.Count())
	t51v.UseOther(cross(cur))
	println("after:", t51v.Count())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/t51v's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/t51v) package{})).count
