// PKGPATH: gno.land/r/zrealm_lend_stranger
package zrealm_lend_stranger

import (
	"gno.land/p/demo/tests/lendlib/evil"
	"gno.land/r/tests/vm/lendvictim"
)

func main(cur realm) {
	println("before:", lendvictim.Count())
	lendvictim.UseRunner(cross(cur), evil.Runner{})
	println("after:", lendvictim.Count())
}

// Output:
// before: 0

// Error:
// gno.land/r/tests/vm/lendvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/lendvictim) package{})).count
