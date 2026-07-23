// PKGPATH: gno.land/r/sys/namereg/v1/filetests/z_5_prop1_filetest

// Test updating a name via GovDAO when the name is already taken
package z_5_prop1_filetest

import (
	"chain"
	"chain/runtime/unsafe"
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
	caller   address = unsafe.OriginCaller()
)

func init(cur realm) {
	bob := testutils.TestAddress("bob")

	// Whitelist this realm as a controller so its Register() can reach r/sys/users.
	// Also register a username for bob bypassing the "nym-" prefix check.
	testing.SetHeight(0)
	susers.AddControllerAtGenesis(cross(cur), chain.PackageAddress("gno.land/r/sys/namereg/v1"))
	susers.RegisterUser(cross(cur), username, bob)
	testing.SetHeight(123)

	// Register alice
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))
	users.Register(cross(cur), "nym-alice123")
}

func main(cur realm) {
	testing.SetOriginCaller(caller)
	testing.SetRealm(testing.NewUserRealm(caller))
	dao.MustCreateProposal(cross(cur), users.ProposeNewName(cross(cur), alice, username))
}

// Error:
// r/sys/users: name/Alias already taken
