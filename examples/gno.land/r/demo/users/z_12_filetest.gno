package main

// SEND: 200000000ugnot

import (
	"strconv"

	"gno.land/r/demo/users"
)

func main() {
	users.Register("", "gnouser", "my profile")

	{
		// Normal usage
		names := users.ListUsersByPrefix("g", 10)
		println("# names: " + strconv.Itoa(len(names)))
		println("name: " + names[0])
	}

	{
		// Empty prefix: match all
		names := users.ListUsersByPrefix("", 10)
		println("# names: " + strconv.Itoa(len(names)))
		println("name: " + names[0])
	}

	{
		// The prefix is before "gnouser"
		names := users.ListUsersByPrefix("gnouseq", 10)
		println("# names: " + strconv.Itoa(len(names)))
	}

	{
		// The prefix is after "gnouser"
		names := users.ListUsersByPrefix("go", 10)
		println("# names: " + strconv.Itoa(len(names)))
	}

	// More tests are in p/demo/avlhelpers
}

// Output:
// # names: 1
// name: gnouser
// # names: 1
// name: gnouser
// # names: 0
// # names: 0
