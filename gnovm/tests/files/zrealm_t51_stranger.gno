// PKGPATH: gno.land/r/zrealm_t51_stranger
package zrealm_t51_stranger

import (
	"gno.land/p/demo/tests/t51lib/evil"
	"gno.land/r/tests/vm/t51v"
)

func main(cur realm) {
	println("before:", t51v.Count())
	t51v.UseRunner(cross(cur), evil.Runner{})
	println("after:", t51v.Count())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/t51v's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/t51v) package{})).count
