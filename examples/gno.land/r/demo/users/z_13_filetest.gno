package main

// SEND: 200000000ugnot

import (
	"strconv"

	"gno.land/r/demo/users"
)

func main() {
	{
		// Verify pre-registered test1 user
		names := users.ListUsersByPrefix("test1", 1)
		println("# names: " + strconv.Itoa(len(names)))
		println("name: " + names[0])
	}
}

// Output:
// # names: 1
// name: test1

