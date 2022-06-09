package main

// SEND: 199gnot

import (
	"gno.land/r/users"
	"std"
)

func main() {
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Error:
// payment must not be less than 200
