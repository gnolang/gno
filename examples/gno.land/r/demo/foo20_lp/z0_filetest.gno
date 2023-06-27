// PKGPATH: gno.land/r/demo/foo20_lp_test
package foo20_lp_test

import (
	"std"

	"gno.land/r/demo/foo20_lp"
)

func main() {
	println("A", foo20_lp.Render(""))
	foo20_lp.DepositFoo20(10000)
	println("B", foo20_lp.Render(""))
	std.TestSetOrigSend(std.Coins{{"ugnot", 100000000}}, nil)
	foo20_lp.DepositUgnot()
	println("C", foo20_lp.Render(""))
}

// Output:
// A Queues: ...
// B Queues: ...
// not implemented
// C Queues: ...
