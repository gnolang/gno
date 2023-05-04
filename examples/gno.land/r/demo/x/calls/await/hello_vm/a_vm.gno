package hello

import "std"

func SayHello() string {

	chainID := std.GetChainID()
	call := std.EncodeMsgCall(chainID, "gno.land/r/demo/x/calls/await/greet", "Greet", []string{"hey"})

	r, err := std.Await(call)
	println("done ")
	println("r is :", r)
	println("errMsg is: ", err)

	return "say hello from A, going to call B" + r
}