package main

import (
	"std"
	"testing"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const user = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func main() {
	testing.SetRealm(std.NewUserRealm(user))

	cross(commondao.Vote)(404, 1, pcommondao.ChoiceYes, "")
}

// Error:
// DAO not found
