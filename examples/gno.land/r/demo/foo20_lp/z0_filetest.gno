// PKGPATH: gno.land/r/demo/foo20_lp_test
package foo20_lp_test

import (
	"std"

	"gno.land/r/demo/foo20_lp"
)

func main() {
	println("A -", foo20_lp.Render(""))
	foo20_lp.DepositFoo20(10_000)
	println("B -", foo20_lp.Render(""))
	std.TestSetOrigSend(std.Coins{{"ugnot", 100_000_000}}, nil)
	foo20_lp.DepositUgnot()
	println("C -", foo20_lp.Render(""))
	foo20_lp.WithdrawFoo20(100)
	println("D -", foo20_lp.Render(""))
	foo20_lp.WithdrawFoo20(100)
	println("E -", foo20_lp.Render(""))
	foo20_lp.WithdrawUgnot(100)
	println("F -", foo20_lp.Render(""))
	println("G -", foo20_lp.Render(""))
}

// Output:
// A - LP is empty.
// B - LP has a total of 10000 foo20 in 1 orders for a capacity of 100 ugnot.
// not implemented
// C - LP has a total of 10000 foo20 in 1 orders for a capacity of 100 ugnot.
// not implemented
// D - LP has a total of 10000 foo20 in 1 orders for a capacity of 100 ugnot.
// not implemented
// E - LP has a total of 10000 foo20 in 1 orders for a capacity of 100 ugnot.
// not implemented
// F - LP has a total of 10000 foo20 in 1 orders for a capacity of 100 ugnot.
// G - LP has a total of 10000 foo20 in 1 orders for a capacity of 100 ugnot.
