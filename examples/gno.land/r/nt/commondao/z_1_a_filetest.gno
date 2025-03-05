package main

import (
	"std"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

func init() {
	std.TestSetOriginCaller(owner)
}

func main() {
	commondao.Invite(user)

	println(commondao.IsInvited(user))
}

// Output:
// true
