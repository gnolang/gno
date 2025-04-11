package main

import (
	"std"
	"testing"

	pcommondao "gno.land/p/gnoland/commondao"

	"gno.land/r/gnoland/commondao"
)

const user = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1

func init() {
	testing.SetOriginCaller(user)
}

func main() {
	commondao.Vote(404, 1, pcommondao.ChoiceYes, "")
}

// Error:
// DAO not found
