package main

import (
	"std"
	"testing"

	"gno.land/r/gnoland/commondao"
)

const user = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	testing.SetOriginCaller(user)
}

func main() {
	commondao.CreateMembersUpdateProposal(404, "", "", "")
}

// Error:
// DAO not found
