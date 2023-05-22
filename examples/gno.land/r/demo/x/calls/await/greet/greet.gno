package greet

import "std"

func Greet(msg string) string {

	chainID := std.GetChainID()
	call := std.EncodeMsgCall(chainID, "gno.land/r/demo/x/calls/await/hola", "Hola", []string{"hola"})

	r, err := std.Await(call)
	println("done ")
	println("r is :", r)
	println("errMsg is: ", err)
	return "greet" + r
}

