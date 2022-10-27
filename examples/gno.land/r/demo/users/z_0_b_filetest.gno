package main

// SEND: 199000000ugnot

import (
	"std"

	"gno.land/r/users"
)

func main() {
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Error:
// payment must not be less than 200000000
