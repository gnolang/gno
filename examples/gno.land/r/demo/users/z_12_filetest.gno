package main

// SEND: 200000000ugnot

import (
	"strconv"

	"gno.land/r/demo/users"
)

func main() {
	users.Register("", "alicia", "my profile")

	{
		// Normal usage
		names := users.ListUsersByPrefix("a", 1)
		println("# names: " + strconv.Itoa(len(names)))
		println("name: " + names[0])
	}

	{
		// Empty prefix: match all
		names := users.ListUsersByPrefix("", 1)
		println("# names: " + strconv.Itoa(len(names)))
		println("name: " + names[0])
	}

	{
		// The prefix is before "alicia"
		names := users.ListUsersByPrefix("alich", 1)
		println("# names: " + strconv.Itoa(len(names)))
	}

	{
		// The prefix is after the last name
		names := users.ListUsersByPrefix("y", 10)
		println("# names: " + strconv.Itoa(len(names)))
	}

	// More tests are in p/demo/avlhelpers
}

// Output:
// # names: 1
// name: alicia
// # names: 1
// name: alicia
// # names: 0
// # names: 0
