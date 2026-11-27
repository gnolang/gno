// PKGPATH: gno.land/r/tests/vm/recoverfvstored
package recoverfvstored

import "gno.land/p/demo/tests/recoverfv"

// A closure the library stores, called through a value, acts for the library:
// its recover catches the library object's panic, as on master. The second
// calls the object cleanly first; the third panics two closures deep.
func main(cur realm) {
	ops := []func() any{recoverfv.StoredMk, recoverfv.StoredNop, recoverfv.StoredNested}
	for i := range ops {
		println(i, ops[i]())
	}
	println("after")
	println("OUTCOME caught")
}

// Output:
// 0 int is not of type string
// 1 int is not of type string
// 2 int is not of type string
// after
// OUTCOME caught
