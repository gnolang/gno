package main

import (
	"std"

	"gno.land/r/nt/commondao"
)

const user = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	std.TestSetOriginCaller(user)
}

func main() {
	commondao.CreateMembersUpdateProposal(404, "", "", "")
}

// Error:
// DAO not found
