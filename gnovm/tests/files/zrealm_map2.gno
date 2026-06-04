// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

func main(cur realm) {
	m := map[string]int{"admin": 1000}
	crossrealm_b.SetObject(cross(cur), m)

	obj := crossrealm_b.GetObject()
	rm := obj.(map[string]int)

	// Storage is authority and is stamped at construction, so m's
	// PkgID stays /r/crossrealm even after being handed to
	// crossrealm_b for storage. The constructor's realm retains
	// write authority, so this assignment succeeds — no readonly-
	// taint panic, the recover is a no-op.
	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("caught panic:", r)
			}
		}()
		rm["attacker"] = 999
	}()

	println("len(m) =", len(m))
	_, ok := m["attacker"]
	println("attacker key exists:", ok)
}

// Output:
// len(m) = 2
// attacker key exists: true
