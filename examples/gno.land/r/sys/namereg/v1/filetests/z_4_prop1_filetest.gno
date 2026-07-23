// PKGPATH: gno.land/r/sys/namereg/v1/filetests/z_4_prop1_filetest

// Test updating a name via GovDAO when there's a canonical collision bypassing it
package z_4_prop1_filetest

import (
	"chain"
	"chain/runtime/unsafe"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gov/dao"
	daov3init "gno.land/r/gov/dao/v3/init"
	users "gno.land/r/sys/namereg/v1"
	susers "gno.land/r/sys/users"
)

var (
	proposalID dao.ProposalID
	username           = "vitalik"
	caller     address = unsafe.OriginCaller()
)

func init(cur realm) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	// Whitelist this realm as a controller so its Register() can reach r/sys/users.
	// Also register a username for bob bypassing the "nym-" prefix check.
	testing.SetHeight(0)
	susers.AddControllerAtGenesis(cross(cur), chain.PackageAddress("gno.land/r/sys/namereg/v1"))
	susers.RegisterUser(cross(cur), "vital1k", bob)
	testing.SetHeight(123)

	daov3init.InitWithUsers(cross(cur), caller)

	// Register alice
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))
	users.Register(cross(cur), "nym-alice123")

	// Proposal to change alice's username to one that canonically colides with bob's
	testing.SetOriginCaller(caller)
	testing.SetRealm(testing.NewUserRealm(caller))
	proposalID = dao.MustCreateProposal(cross(cur), users.ProposeNewName(cross(cur), alice, username))
}

func main(cur realm) {
	testing.SetOriginCaller(caller)

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, proposalID))

	dao.ExecuteProposal(cross(cur), dao.ProposalID(0))

	// Make sure the new username is assigned to Alice
	data, isCurrent := susers.ResolveName(username)
	println(isCurrent)
	println(data.Addr())

	// Make sure canonical collision was bypassed and now
	// the canonical name is linked to Alice's username.
	current, _ := susers.IsCanonicalTaken("v1tal1k")
	println(current)
}

// Output:
// true
// g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh
// vitalik
