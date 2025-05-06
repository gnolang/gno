package main

import (
	"std"
	"testing"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

func main() {
	testing.SetRealm(std.NewUserRealm(owner))

	cross(commondao.Invite)(user)

	println(commondao.IsInvited(user))
}

// Output:
// true
