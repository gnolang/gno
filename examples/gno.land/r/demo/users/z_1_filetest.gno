package main

// SEND: 2000000000ugnot

import (
	"gno.land/r/users"
)

func main() {
	users.Register("", "gnouser", "my profile")
	println("done")
}

// Output:
// done
