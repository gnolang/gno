package main

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const user = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func main() {
	// Calling with a user that was not invited
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.New(cross, "Foo")
}

// Error:
// unauthorized
