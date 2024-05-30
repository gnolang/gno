package main

import (
	"std"
)

func main() {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/users"))
	banker := std.GetBanker(std.BankerTypeOrigSend)
	println(banker)
	std.TestSetRealm(std.NewCodeRealm("gno.land/p/demo/users"))
	defer func() {
		println(recover())
	}()
	banker = std.GetBanker(std.BankerTypeOrigSend)
}



// Output:
// (struct{(1 std.BankerType)} std.banker)
// pure packages cannot instantiate a banker