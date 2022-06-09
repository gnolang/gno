package main

// SEND: 10gnot

import (
	"gno.land/r/users"
)

func main() {
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Error:
// payment must not be less than 200
