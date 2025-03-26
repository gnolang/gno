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
	slug  = "foo"
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

	// Add a subDAO with the same slug
	commondao.NewSubDAO("Foo", slug, rootDAO)
}

func main() {
	commondao.NewSubDAO("Foo", slug, rootDAO)
}

// Error:
// subDAO slug is already taken
