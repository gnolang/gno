// PKGPATH: gno.land/r/tests/vm/recoverfvzvia
package recoverfvzvia

import "gno.land/r/tests/vm/recoverfvz"

// A realm's plain function run inside a borrowed library method, reached
// from that realm's read-only call, recovers its object's panic, as on master.
func main(cur realm) {
	println("rec:", recoverfvz.Via())
	println("after")
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// after
// OUTCOME caught
