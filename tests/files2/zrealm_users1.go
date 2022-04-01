package main

// SEND: 2000gnot

import (
	"gno.land/r/users"
)

func main() {
	err := users.Register("", "jaekwon", "my profile")
	if err != nil {
		panic(err)
	}
	println("done")
}

// Output:
// done
