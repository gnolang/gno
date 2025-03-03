// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"

	pcommondao "gno.land/p/nt/commondao"

	"gno.land/r/nt/commondao"
)

const (
	owner = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	user  = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // @test2
)

var rootDAO *pcommondao.CommonDAO

func init() {
	// Invite a user to be able to start creating DAOs
	std.TestSetOriginCaller(owner)
	commondao.Invite(user)

	// The origin must be the invited user
	std.TestSetOriginCaller(user)

	// Create the root DAO
	rootDAO = commondao.New("Parent DAO")
}

func main() {
	commondao.NewSubDAO("Foo", "", rootDAO)
}

// Error:
// invalid DAO slug
