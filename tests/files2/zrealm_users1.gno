package main

// SEND: 2000gnot

import (
	"gno.land/r/users"
)

func main() {
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Output:
// done
