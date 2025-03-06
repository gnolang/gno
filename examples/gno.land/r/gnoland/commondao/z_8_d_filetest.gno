// PKGPATH: gno.land/r/demo/commondao_test
package commondao_test

import (
	"std"

	"gno.land/r/gnoland/commondao"
)

const (
	owner = std.Address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

func init() {
	// Invite a user to be able to start creating DAOs
	std.TestSetOriginCaller(owner)
	commondao.Invite(user)

	// Create a DAO
	std.TestSetOriginCaller(user)
}

func main() {
	// Use a caller that is not the owner of the DAO
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))

	commondao.GetSubDAO(1, "foo/bar")
}

// Error:
// unauthorized
