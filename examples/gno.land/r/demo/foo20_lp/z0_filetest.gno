// PKGPATH: gno.land/r/demo/foo20_lp_test
package foo20_lp_test

import "gno.land/r/demo/foo20_lp"

func main() {
	println("A", foo20_lp.Render(""))
	foo20_lp.DepositFoo20()
	println("B", foo20_lp.Render(""))
	foo20_lp.DepositGnot()
	println("C", foo20_lp.Render(""))
}

// Output:
// A Queues: ...
// B Queues: ...
// C Queues: ...
