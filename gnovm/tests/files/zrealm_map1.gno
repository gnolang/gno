// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

func main(cur realm) {
	// Create a map in this realm with multiple entries
	m := map[string]int{"admin": 1000, "user1": 500, "user2": 200}

	crossrealm_b.SetObject(cross, m)

	obj := crossrealm_b.GetObject()
	rm := obj.(map[string]int)

	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("caught panic:", r)
			}
		}()
		delete(rm, "admin")
	}()

	val := m["admin"]

	println("after: m[admin] =", val)
}

// Output:
// caught panic: cannot modify external-realm or non-realm object
// after: m[admin] = 1000
