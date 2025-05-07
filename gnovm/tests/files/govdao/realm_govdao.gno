// PKGPATH: gno.land/r/govdao_test
package govdao_test

import (
	"gno.land/r/gov/dao"
	daoimpl "gno.land/r/gov/dao/v3/impl"
	daoinit "gno.land/r/gov/dao/v3/init"
	"std"
	"testing"
)

var (
	user1 std.Address = std.DerivePkgAddr("user1.gno")
	user2 std.Address = std.DerivePkgAddr("user2.gno")
	user3 std.Address = std.DerivePkgAddr("user3.gno")
)

func init() {
	daoinit.InitWithUsers(user1, user2, user3)
	govDAO := daoimpl.NewGovDAO() // A DAO object without an attached realm.
	testing.SetRealm(std.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))
	cross(dao.UpdateImpl)(dao.UpdateRequest{
		DAO: govDAO,
	})
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
		func() error {
			return nil
		},
		"",
	)
	pid := cross(dao.MustCreateProposal)(
		dao.NewProposalRequest("Proposal Title: Dummy proposal", "Description: Do nothing", e),
	)
	return pid

}

func VoteProposal(pid dao.ProposalID) {

	cross(dao.MustVoteOnProposal)(dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: pid,
	})

}
func ExecuteProposal(pid dao.ProposalID) bool {
	accepted := cross(dao.ExecuteProposal)(pid)
	return accepted

}

func Render() {
	println(dao.Render(""))
}

// Output:
// true
// # Active Proposals:
// ## Proposal with id: 0
// ### Title: Proposal Title: Dummy proposal
//
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// Description: Do nothing
//
//
// ### Proposal Status:
//
// - **PROPOSAL HAS BEEN ACCEPTED**
// - Allowed tiers to vote: T1 T2 T3
// - YES PERCENT: 66.66666666666666%
// - NO PERCENT: 0%
// - [Go to votes list](/r/gov/dao:0/votes).
