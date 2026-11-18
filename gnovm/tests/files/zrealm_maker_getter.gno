// PKGPATH: gno.land/r/zrealm_maker_getter
package zrealm_maker_getter

// A closure the victim's helper builds while read-only has no maker, so the
// victim's gate runs it read-only.

import "gno.land/r/tests/vm/cbvictim"

func main(cur realm) {
	cbvictim.RunS(cross(cur), cbvictim.Getter())
	println("owner:", cbvictim.Owner)
}

// Error:
// gno.land/r/tests/vm/cbvictim's code was entered without crossing and is read-only (call its crossing function): (const (ref(gno.land/r/tests/vm/cbvictim) package{})).Owner
