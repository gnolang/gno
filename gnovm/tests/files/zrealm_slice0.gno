// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

// Verify that append within capacity on a readonly-tainted slice panics,
// even when the backing array's ObjectID belongs to the current realm.
var s []int

func init() {
	s = make([]int, 3, 10)
	s[0] = 1
	s[1] = 2
	s[2] = 3
}

func main(cur realm) {
	crossrealm_b.SetObject(cross, s)
	rs := crossrealm_b.Object.([]int)

	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("append panic:", r)
			}
		}()
		_ = append(rs, 99)
	}()

	println("s[0]:", s[0])
	println("s[1]:", s[1])
	println("s[2]:", s[2])
	println("len:", len(s))
}

// Output:
// append panic: cannot append to readonly tainted slice
// s[0]: 1
// s[1]: 2
// s[2]: 3
// len: 3
