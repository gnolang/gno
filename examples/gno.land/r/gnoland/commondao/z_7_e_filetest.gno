// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"
	"testing"

	pcommondao "gno.land/p/gnoland/commondao"

	"gno.land/r/gnoland/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var rootDAO *pcommondao.CommonDAO

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetOriginCaller(owner)
	commondao.Invite(user)

	// The origin must be the invited user
	testing.SetOriginCaller(user)

	// Create the root DAO
	rootDAO = commondao.New("Parent DAO")
}

func main() {
	commondao.NewSubDAO("Foo", "", rootDAO)
}

// Error:
// invalid DAO slug
