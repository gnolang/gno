// PKGPATH: gno.land/r/demo/crosssend0
// SEND: 100ugnot
package crosssend0

import (
	"chain"
	"chain/banker"
)

func helper(cur realm) int64 { return banker.CallSend().AmountOf("ugnot") }

// Deposit reads what THIS call delivered; a plain helper(cur) call sees the
// same frame, a self-cross does not.
func Deposit(cur realm) {
	println("deposit got", banker.CallSend().AmountOf("ugnot"))
	println("helper got ", helper(cur))
	println("self-cross ", Peek(cross(cur)))
}

func Peek(cur realm) int64 { return banker.CallSend().AmountOf("ugnot") }

func main(cur realm) {
	println("entry got  ", banker.CallSend().AmountOf("ugnot"))
	Deposit(cross(cur, chain.Coins{{"ugnot", 40}}))
	Deposit(cross(cur))
	println("entry still", banker.CallSend().AmountOf("ugnot"))
}

// Output:
// entry got   100
// deposit got 40
// helper got  40
// self-cross  0
// deposit got 0
// helper got  0
// self-cross  0
// entry still 100
