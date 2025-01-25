package main

import (
	"std"
	"testing"

	"gno.land/r/demo/users"
)

func main() {
	testing.SetOriginSend(std.Coins{std.NewCoin("dontcare", 1)}, nil)
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Error:
// incompatible coin denominations: dontcare, ugnot
