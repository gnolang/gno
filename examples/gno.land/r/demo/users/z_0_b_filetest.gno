package main

// SEND: 19900000ugnot

import (
	"gno.land/r/demo/users"
)

func main() {
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Error:
// payment must not be less than 20000000
