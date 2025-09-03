// PKGPATH: gno.land/r/govdao
package govdao

import (
	"std"
	"testing"

	"gno.land/r/gov/dao"
	daoimpl "gno.land/r/gov/dao/v3/impl"
	daoinit "gno.land/r/gov/dao/v3/init"
	susers "gno.land/r/sys/users"
)

var (
	user1 std.Address = std.DerivePkgAddr("user1.gno")
	user2 std.Address = std.DerivePkgAddr("user2.gno")
	user3 std.Address = std.DerivePkgAddr("user3.gno")
)

func init() {
	daoinit.InitWithUsers(user1, user2, user3)
	registerTestUsers(user1, user2, user3)
	govDAO := daoimpl.NewGovDAO() // A DAO object without an attached realm.
	testing.SetRealm(std.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))
	dao.UpdateImpl(cross, dao.UpdateRequest{
		DAO: govDAO,
	})
}

const gUsersV1Path = "gno.land/r/gnoland/users/v1"

// Register a namespace for every addresses
// Necessary to test GovDAO Vote
func registerTestUsers(addrs ...std.Address) {
	// Set realm to users admin to register test user
	testing.SetRealm(std.NewCodeRealm(gUsersV1Path))
	for _, addr := range addrs {
		err := susers.RegisterUser(cross, addr.String()[1:], addr)
		if err != nil {
			panic(err.Error() + " : " + addr.String())
		}
	}
}

func main() {
	// run the test as end user instead of the testing realm
	testing.SetRealm(std.NewUserRealm(user1))
	pid := CreateProposal()
	VoteProposal(pid)
	// run the test as another user
	func() {
		testing.SetRealm(std.NewUserRealm(user2))
		VoteProposal(pid)
	}()
	accepted := ExecuteProposal(pid)
	println(accepted)
	Render()
}

func CreateProposal() dao.ProposalID {
	e := dao.NewSimpleExecutor(
		func(realm) error {
			return nil
		},
		"",
	)
	pid := dao.MustCreateProposal(cross,
		dao.NewProposalRequest("Proposal Title: Dummy proposal", "Description: Do nothing", e),
	)
	return pid
}

func VoteProposal(pid dao.ProposalID) {
	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: pid,
	})
}

func ExecuteProposal(pid dao.ProposalID) bool {
	accepted := dao.ExecuteProposal(cross, pid)
	return accepted
}

func Render() {
	println(dao.Render(""))
}

// Output:
// true
// # GovDAO
// ## Members
// [> Go to Memberstore <](/r/gov/dao/v3/memberstore)
// ## Proposals
// ### [Prop #0 - Proposal Title: Dummy proposal](/r/gov/dao:0)
// Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// Status: ACCEPTED
//
// Tiers eligible to vote: T1, T2, T3
//
// ---
