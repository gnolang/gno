// PKGPATH: gno.land/r/borrow
package borrow

import crossrealm "gno.land/r/demo/tests/crossrealm_d"

func main() {
	crossrealm.GetRoot().Modify()
	println(crossrealm.GetRoot())
}

// Output:
// (struct{(1 int)} gno.land/r/demo/tests/crossrealm_d.MyMutator)
