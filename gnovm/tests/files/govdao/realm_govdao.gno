// PKGPATH: gno.land/r/govdao
package govdao

import (
	"chain"
	"testing"

	"gno.land/r/gov/dao"
	daoimpl "gno.land/r/gov/dao/v3/impl"
	daoinit "gno.land/r/gov/dao/v3/init"
)

var (
	user1 address = chain.PackageAddress("user1.gno")
	user2 address = chain.PackageAddress("user2.gno")
	user3 address = chain.PackageAddress("user3.gno")
)

func init(cur realm) {
	daoinit.InitWithUsers(cross2(cur), user1, user2, user3)
	govDAO := daoimpl.NewGovDAO() // A DAO object without an attached realm.
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))
	// UpdateRequest is declared in /r/gov/dao; the construction-time
	// check forbids allocating it from /r/govdao directly. Use the
	// dao-side constructor, which runs under Rule-1 borrow to /r/gov/dao.
	dao.UpdateImpl(cross2(cur), dao.NewUpdateRequest(govDAO, nil))
}

func main() {
	// run the test as end user instead of the testing realm
	testing.SetRealm(testing.NewUserRealm(user1))
	pid := CreateProposal()
	VoteProposal(pid)
	// run the test as another user
	func() {
		testing.SetRealm(testing.NewUserRealm(user2))
		VoteProposal(pid)
	}()
	accepted := ExecuteProposal(pid)
	println(accepted)
	Render()
}

func CreateProposal() dao.ProposalID {
	// NewSimpleExecutor takes (_ int, rlm realm, ...); mint a live cur
	// via a same-realm cross to pass through.
	var e *dao.SimpleExecutor
	func(cur realm) {
		e = dao.NewSimpleExecutor(0, cur,
			func(realm) error { return nil },
			"",
		)
	}(cross)
	pid := dao.MustCreateProposal(cross,
		dao.NewProposalRequest("Proposal Title: Dummy proposal", "Description: Do nothing", e),
	)
	return pid
}

func VoteProposal(pid dao.ProposalID) {
	// VoteRequest is /r/gov/dao-declared; use the dao-side constructor
	// to satisfy the construction-time check.
	dao.MustVoteOnProposal(cross, dao.NewVoteRequest(dao.YesVote, pid))
}

func ExecuteProposal(pid dao.ProposalID) bool {
	accepted := dao.ExecuteProposal(cross, pid)
	return accepted
}

func Render() {
	println(dao.Render(cross, ""))
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
