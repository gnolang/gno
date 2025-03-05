package main

import (
	"std"

	"gno.land/r/nt/commondao"
)

const user = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2

func init() {
	std.TestSetOriginCaller(user)
}

func main() {
	commondao.CreateMembersUpdateProposal(404, "", "", "")
}

// Error:
// DAO not found
