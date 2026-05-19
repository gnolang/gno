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

func main(cur realm) {
	// run the test as end user instead of the testing realm.
	// SetRealm mutates main's cur HIV in place; subsequent helpers
	// receive the mutated cur as rlm.
	testing.SetRealm(testing.NewUserRealm(user1))
	pid := CreateProposal(0, cur)
	VoteProposal(0, cur, pid)
	// switch to another user and vote again. SetRealm must run in
	// main's crossing frame (not a non-crossing lambda) so the in-place
	// HIV mutation lands on main's cur — see migration_guide.md §12.
	testing.SetRealm(testing.NewUserRealm(user2))
	VoteProposal(0, cur, pid)
	accepted := ExecuteProposal(0, cur, pid)
	println(accepted)
	Render(0, cur)
}

func CreateProposal(_ int, rlm realm) dao.ProposalID {
	e := dao.NewSimpleExecutor(0, rlm,
		func(realm) error { return nil },
		"",
	)
	return dao.MustCreateProposal(cross2(rlm),
		dao.NewProposalRequest("Proposal Title: Dummy proposal", "Description: Do nothing", e),
	)
}

func VoteProposal(_ int, rlm realm, pid dao.ProposalID) {
	// VoteRequest is /r/gov/dao-declared; use the dao-side constructor
	// to satisfy the construction-time check.
	dao.MustVoteOnProposal(cross2(rlm), dao.NewVoteRequest(dao.YesVote, pid))
}

func ExecuteProposal(_ int, rlm realm, pid dao.ProposalID) bool {
	return dao.ExecuteProposal(cross2(rlm), pid)
}

func Render(_ int, rlm realm) {
	println(dao.Render(cross2(rlm), ""))
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
