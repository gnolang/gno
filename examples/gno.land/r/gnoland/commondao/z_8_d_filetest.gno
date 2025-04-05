// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"
	"testing"

	"gno.land/r/gnoland/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

func init() {
	// Invite a user to be able to start creating DAOs
	testing.SetOriginCaller(owner)
	commondao.Invite(user)

	// Create a DAO
	testing.SetOriginCaller(user)
}

func main() {
	// Use a caller that is not the owner of the DAO
	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))

	commondao.GetSubDAO(1, "foo/bar")
}

// Error:
// unauthorized
