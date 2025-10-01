package main

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const user = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func main() {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Vote(cross, 404, 1, pcommondao.ChoiceYes, "")
}

// Error:
// DAO not found
