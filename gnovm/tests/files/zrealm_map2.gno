// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

func main(cur realm) {
	// Create a map in this realm.
	m := map[string]int{"admin": 1000}

	crossrealm_b.SetObject(cross, m)

	obj := crossrealm_b.GetObject()
	rm := obj.(map[string]int)

	// Without the fix, the defer/recover would swallow the readonly panic
	// and leave a phantom entry in the map (key "attacker" with value 0).
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
// caught panic: cannot directly modify readonly tainted object (w/o method): rm<~VPBlock(1,0)>[(const ("attacker" string))]
// len(m) = 1
// attacker key exists: false
