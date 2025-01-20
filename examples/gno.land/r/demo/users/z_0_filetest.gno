package main

import (
	"std"

	"gno.land/r/demo/users"
)

func main() {
	std.TestSetOrigSend(std.Coins{std.NewCoin("dontcare", 1)}, nil)
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Error:
// incompatible coin denominations: dontcare, ugnot
