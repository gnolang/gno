package hello

import "std"

func Hello() string {

	chainID := std.GetChainID()
	call := std.EncodeMsgCall(chainID, "gno.land/r/demo/x/calls/await/greet", "Greet", []string{"hey"})

	r, err := std.Await(call)
	println("done ")
	println("r is :", r)
	println("errMsg is: ", err)

	return "hello" + r
}