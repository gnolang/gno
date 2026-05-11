// Test updating a name via GovDAO when the name is already taken
package main

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gov/dao"
	users "gno.land/r/sys/namereg/v1"
	susers "gno.land/r/sys/users"
)

var (
	proposalID dao.ProposalID

	username         = "vitalik"
	alice            = testutils.TestAddress("alice")
	caller   address = runtime.OriginCaller()
)

func init() {
	bob := testutils.TestAddress("bob")

	// Whitelist this realm as a controller so its Register() can reach r/sys/users.
	// Also register a username for bob bypassing the "nym-" prefix check.
	testing.SetHeight(0)
	susers.AddControllerAtGenesis(cross, chain.PackageAddress("gno.land/r/sys/namereg/v1"))
	susers.RegisterUser(cross, username, bob)
	testing.SetHeight(123)

	// Register alice
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))
	users.Register(cross, "nym-alice123")
}

func main() {
	testing.SetOriginCaller(caller)
	testing.SetRealm(testing.NewUserRealm(caller))
	dao.MustCreateProposal(cross, users.ProposeNewName(alice, username))
}

// Error:
// r/sys/users: name/Alias already taken
