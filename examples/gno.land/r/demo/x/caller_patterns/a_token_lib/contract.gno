package token_lib

import "std"

func TestGetCaller() {
	//println("token_lib.TestGetCaller=", std.GetCaller())
}

func TestGetOrigCaller() {
	println("token_lib.TestGetOrigCaller=", std.GetOrigCaller())
}
