// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

// With the readonly taint removed, assign and delete succeed: the map's
// ObjectID belongs to the current realm, so the ownership check permits
// the writes. (Pre-removal the sticky taint blocked them.)
var m map[string]int

func init() {
	m = map[string]int{"admin": 1000, "user1": 500}
}

func main(cur realm) {
	crossrealm_b.SetObject(cross(cur), m)
	rm := crossrealm_b.Object.(map[string]int)

	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("assign panic:", r)
			}
		}()
		rm["attacker"] = 999
	}()

	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("delete panic:", r)
			}
		}()
		delete(rm, "admin")
	}()

	_, hasAdmin := m["admin"]
	println("admin exists:", hasAdmin)
	println("user1:", m["user1"])
	println("len:", len(m))
}

// Output:
// admin exists: false
// user1: 500
// len: 2
