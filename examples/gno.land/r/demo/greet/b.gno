package greet

import "std"

func Greet(msg string) string {

	chainID := std.GetChainID()
	call := std.EncodeMsgCall(chainID, "gno.land/r/demo/hola", "Sing", []string{"song"})

	r, err := std.Send(call)
	println("done ")
	println("r is :", r)
	println("errMsg is: ", err)
	return "greet from b" + r
}

