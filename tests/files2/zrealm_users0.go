package main

// SEND: 100gnot

import (
	"gno.land/r/users"
)

func main() {
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Error:
// payment must be exactly 2000 gnots
