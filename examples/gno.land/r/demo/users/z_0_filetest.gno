package main

import (
	"std"

	"gno.land/r/demo/users"
)

func main() {
	std.TestSetOrigSend(std.Coins{{"dontcare", 1}}, nil)
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Error:
// invalid coin denominations: dontcare
