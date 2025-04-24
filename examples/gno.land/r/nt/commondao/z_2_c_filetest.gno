package main

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const user = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func main() {
	// Caller user is not an invited one
	testing.SetRealm(std.NewUserRealm(user))

	commondao.New("Foo")
}

// Error:
// unauthorized
