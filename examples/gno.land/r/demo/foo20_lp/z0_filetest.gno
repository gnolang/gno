// PKGPATH: gno.land/r/demo/foo20_lp_test
package foo20_lp_test

import (
	"std"

	"gno.land/r/demo/foo20_lp"
)

func main() {
	println("A -", foo20_lp.Render(""))
	println(foo20_lp.Render("queues"))
	foo20_lp.DepositFoo20(10_000)
	println("B -", foo20_lp.Render(""))
	foo20_lp.DepositFoo20(10_000)
	println("C -", foo20_lp.Render(""))
	foo20_lp.DepositFoo20(10_000)
	foo20_lp.DepositFoo20(10_000)
	foo20_lp.DepositFoo20(10_000)
	foo20_lp.DepositFoo20(10_000)
	println(foo20_lp.Render("queues"))
	println("D -", foo20_lp.Render(""))
	std.TestSetOrigSend(std.Coins{{"ugnot", 100_000_000}}, nil)
	foo20_lp.DepositUgnot()
	println("E -", foo20_lp.Render(""))
	foo20_lp.WithdrawFoo20(100)
	println("F -", foo20_lp.Render(""))
	foo20_lp.WithdrawFoo20(100)
	println("G -", foo20_lp.Render(""))
	foo20_lp.WithdrawUgnot(100)
	println("H -", foo20_lp.Render(""))
	println(foo20_lp.Render("queues"))
}

// Output:
// A - LP is empty.
// ## Queues are empty
// B - LP has a total of 10000 foo20 in 1 orders for a capacity of 1000000 ugnot.
// C - LP has a total of 20000 foo20 in 2 orders for a capacity of 2000000 ugnot.
// ## foo20Queue (total=60000, orders=6)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=10000foo20)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=20000foo20)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=30000foo20)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=40000foo20)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=50000foo20)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=60000foo20)
//
// D - LP has a total of 60000 foo20 in 6 orders for a capacity of 6000000 ugnot.
// not implemented: C
// E - LP has a total of 60000 foo20 in 6 orders for a capacity of 6000000 ugnot.
// not implemented: D
// F - LP has a total of 60000 foo20 in 6 orders for a capacity of 6000000 ugnot.
// not implemented: D
// G - LP has a total of 60000 foo20 in 6 orders for a capacity of 6000000 ugnot.
// not implemented: E
// H - LP has a total of 60000 foo20 in 6 orders for a capacity of 6000000 ugnot.
// ## foo20Queue (total=60000, orders=6)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=10000foo20)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=20000foo20)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=30000foo20)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=40000foo20)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=50000foo20)
// - addr=g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm amount=10000foo20 (cum=60000foo20)
