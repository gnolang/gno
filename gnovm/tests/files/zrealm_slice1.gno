// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

// With the readonly taint removed, copy into the slice succeeds: the
// backing array's ObjectID belongs to the current realm, so the ownership
// check permits the write. (Pre-removal the sticky taint blocked it.)
var s []int

func init() {
	s = []int{1, 2, 3}
}

func main(cur realm) {
	crossrealm_b.SetObject(cross(cur), s)
	rs := crossrealm_b.Object.([]int)

	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("copy panic:", r)
			}
		}()
		copy(rs, []int{99, 99, 99})
	}()

	println("s[0]:", s[0])
	println("s[1]:", s[1])
	println("s[2]:", s[2])
}

// Output:
// s[0]: 99
// s[1]: 99
// s[2]: 99
