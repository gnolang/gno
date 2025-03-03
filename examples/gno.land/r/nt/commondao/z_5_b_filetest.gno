package main

import (
	"std"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const user = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2

func init() {
	std.TestSetOriginCaller(user)
}

func main() {
	commondao.Vote(404, 1, pcommondao.ChoiceYes)
}

// Error:
// DAO not found
