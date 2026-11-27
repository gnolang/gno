// PKGPATH: gno.land/r/tests/vm/recoverfvzfunc
package recoverfvzfunc

import "gno.land/r/tests/vm/recoverfvz"

// A realm's plain function reached as a value is recovered, as on master:
// the panic of its own object does not leave the function.
func main(cur realm) {
	println("rec:", recoverfvz.Get()())
	println("after")
	println("OUTCOME caught")
}

// Output:
// rec: int is not of type string
// after
// OUTCOME caught
