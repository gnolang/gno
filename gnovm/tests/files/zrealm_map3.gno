// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

// Verify that assign and delete on a readonly-tainted map both panic,
// even when the map's ObjectID belongs to the current realm.
var m map[string]int

func init() {
	m = map[string]int{"admin": 1000, "user1": 500}
}

func main(cur realm) {
	crossrealm_b.SetObject(cross, m)
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
// assign panic: cannot directly modify readonly tainted object (w/o method): rm<~VPBlock(1,0)>[(const ("attacker" string))]
// delete panic: cannot delete from readonly tainted map
// admin exists: true
// user1: 500
// len: 2
