package main

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const user = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func main() {
	// Calling with a user that was not invited
	testing.SetRealm(std.NewUserRealm(user))

	commondao.New(cross, "Foo")
}

// Error:
// unauthorized
