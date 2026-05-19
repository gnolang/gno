package impl

import (
	"fmt"
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
)


// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
func init(cur realm) {
	loadMembers(cur)

	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(govDAO, []string{"gno.land/r/gov/dao/v3/impl"}))
}

var (
	m1    = testutils.TestAddress("m1")
	m11   = testutils.TestAddress("m1.1")
	m111  = testutils.TestAddress("m1.1.1")
	m1111 = testutils.TestAddress("m1.1.1.1")
	m2    = testutils.TestAddress("m2")
	m3    = testutils.TestAddress("m3")
	m4    = testutils.TestAddress("m4")
	m5    = testutils.TestAddress("m5")
	m6    = testutils.TestAddress("m6")

	noMember = testutils.TestAddress("nm1")
)

func loadMembers(cur realm) {
	// This is needed because state is saved between unit tests,
	// and we want to avoid having real members used on tests
	mstore := memberstore.Get(0, cur)
	mstore.DeleteAll()

	mstore.SetTier(memberstore.T1)
	mstore.SetTier(memberstore.T2)
	mstore.SetTier(memberstore.T3)

	mstore.SetMember(memberstore.T1, m1, memberByTier(memberstore.T1))
	mstore.SetMember(memberstore.T1, m11, memberByTier(memberstore.T1))
	mstore.SetMember(memberstore.T1, m111, memberByTier(memberstore.T1))
	mstore.SetMember(memberstore.T1, m1111, memberByTier(memberstore.T1))

	mstore.SetMember(memberstore.T2, m2, memberByTier(memberstore.T2))
	mstore.SetMember(memberstore.T2, m3, memberByTier(memberstore.T2))
	mstore.SetMember(memberstore.T3, m4, memberByTier(memberstore.T3))
	mstore.SetMember(memberstore.T3, m5, memberByTier(memberstore.T3))
	mstore.SetMember(memberstore.T3, m6, memberByTier(memberstore.T3))
}

func TestCreateProposalAndVote(cur realm, t *testing.T) {
	loadMembers(cur)

	portfolio := "# This is my portfolio:\n\n- THINGS"

	testing.SetOriginCaller(noMember)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	nm1 := testutils.TestAddress("nm1")

	urequire.AbortsWithMessage(t, cur, "Only T1 and T2 members can be added by proposal. To add a T3 member use AddMember function directly.", func(cur realm) {
		dao.MustCreateProposal(cross(cur), NewAddMemberRequest(cur, nm1, memberstore.T3, portfolio))
	})

	urequire.AbortsWithMessage(t, cur, "proposer is not a member", func(cur realm) {
		dao.MustCreateProposal(cross(cur), NewAddMemberRequest(cur, nm1, memberstore.T2, portfolio))
	})

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	proposalRequest := NewAddMemberRequest(cur, nm1, memberstore.T2, portfolio)

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))
	pid := dao.MustCreateProposal(cross(cur), proposalRequest)
	urequire.Equal(t, int(pid), 0)

	// m1 votes yes because that member is interested on it
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(0)))

	testing.SetOriginCaller(m11)

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.NoVote, dao.ProposalID(0)))

	testing.SetOriginCaller(m2)

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.NoVote, dao.ProposalID(0)))

	testing.SetOriginCaller(m3)

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.NoVote, dao.ProposalID(0)))

	testing.SetOriginCaller(m4)

	urequire.AbortsWithMessage(t, cur, "member on specified tier is not allowed to vote on this proposal", func() {
		dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.NoVote, dao.ProposalID(0)))
	})

	testing.SetOriginCaller(m111)

	// Same effect as:
	// dao.MustVoteOnProposal(dao.VoteRequest{
	// 	Option:     dao.NoVote,
	// 	ProposalID: dao.ProposalID(0),
	// })
	dao.MustVoteOnProposalSimple(cross(cur), 0, "NO")

	urequire.Equal(t, true, strings.Contains(dao.Render(cross(cur), ""), "Prop #0 - New T2 Member Proposal"))
	// urequire.Equal(t, true, strings.Contains(dao.Render(cross(cur), ""), "Author: "+m1.String()))

	urequire.AbortsWithMessage(t, cur, "proposal didn't reach supermajority yet: 66.66", func() {
		dao.ExecuteProposal(cross(cur), dao.ProposalID(0))
	})

	testing.SetOriginCaller(m1111)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.NoVote, dao.ProposalID(0)))

	accepted := dao.ExecuteProposal(cross(cur), dao.ProposalID(0))
	urequire.Equal(t, false, accepted)

	urequire.Equal(t, true, contains(dao.Render(cross(cur), "0"), "**PROPOSAL HAS BEEN DENIED**"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), "0"), "NO PERCENT: 81.25%"))
}

func TestExecutorCreationRealm(cur realm, t *testing.T) {
	loadMembers(cur)

	// Test that executor creation realm is captured correctly
	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/template/contract"))

	// Create executor in the template contract realm
	executor := dao.NewSimpleExecutor(0, cur, func(realm) error { return nil }, "Test executor from template")

	proposalRequest := dao.NewProposalRequest(
		"Test Proposal",
		"This proposal tests executor creation realm tracking",
		executor,
	)

	// Create proposal from user realm (user can call DAO directly)
	testing.SetRealm(testing.NewUserRealm(m1))
	pid := dao.MustCreateProposal(cross(cur), proposalRequest)

	// Get the proposal
	prop := dao.MustGetProposal(pid)

	// Verify the author is m1
	urequire.Equal(t, m1, prop.Author())

	// Verify the executor creation realm is captured correctly
	urequire.Equal(t, "gno.land/r/template/contract", prop.ExecutorCreationRealm())

	// Check that it's displayed in the individual proposal render output
	individualRendered := dao.Render(cross(cur), pid.String())
	urequire.Equal(t, true, contains(individualRendered, "Executor created in: gno.land/r/template/contract"))
	urequire.Equal(t, true, contains(individualRendered, "Test executor from template"))

	// Also verify the main content is there
	urequire.Equal(t, true, contains(individualRendered, "Test Proposal"))
	urequire.Equal(t, true, contains(individualRendered, "This proposal tests executor creation realm tracking"))
}

func TestProposalPagination(cur realm, t *testing.T) {
	loadMembers(cur)
	portfolio := "### This is my portfolio:\n\n- THINGS"

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	nm1 := testutils.TestAddress("nm1")

	var pid dao.ProposalID

	proposalRequest := NewAddMemberRequest(cur, nm1, memberstore.T2, portfolio)

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross(cur), proposalRequest)

	// TODO: tests keep the same vm state: https://github.com/gnolang/gno/issues/1982
	urequire.Equal(t, 2, int(pid))

	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross(cur), proposalRequest)
	urequire.Equal(t, 3, int(pid))

	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross(cur), proposalRequest)
	urequire.Equal(t, 4, int(pid))

	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross(cur), proposalRequest)
	urequire.Equal(t, 5, int(pid))

	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross(cur), proposalRequest)
	urequire.Equal(t, 6, int(pid))

	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross(cur), proposalRequest)
	urequire.Equal(t, 7, int(pid))

	fmt.Println(dao.Render(cross(cur), ""))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), ""), "### [Prop #7 - New T2 Member Proposal](/r/gov/dao:7)"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), ""), "### [Prop #6 - New T2 Member Proposal](/r/gov/dao:6)"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), ""), "### [Prop #5 - New T2 Member Proposal](/r/gov/dao:5)"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), ""), "### [Prop #4 - New T2 Member Proposal](/r/gov/dao:4)"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), ""), "### [Prop #3 - New T2 Member Proposal](/r/gov/dao:3)"))

	urequire.Equal(t, true, contains(dao.Render(cross(cur), "?page=2"), "### [Prop #2 - New T2 Member Proposal](/r/gov/dao:2)"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), "?page=2"), "### [Prop #1 - Test Proposal](/r/gov/dao:1)"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), "?page=2"), "### [Prop #0 - New T2 Member Proposal](/r/gov/dao:0)"))
}

func TestUpgradeDaoImplementation(cur realm, t *testing.T) {
	loadMembers(cur)

	testing.SetOriginCaller(noMember)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	urequire.PanicsWithMessage(t, cur, "proposer is not a member", func() {
		NewUpgradeDaoImplRequest(cur, govDAO, "gno.land/r/gov/dao/v4/impl", "Something happened and we have to fix it.")
	})

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	preq := NewUpgradeDaoImplRequest(cur, govDAO, "gno.land/r/gov/dao/v4/impl", "Something happened and we have to fix it.")

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))
	pid := dao.MustCreateProposal(cross(cur), preq)
	urequire.Equal(t, int(pid), 8)

	// m1 votes yes because that member is interested on it
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(pid)))

	testing.SetOriginCaller(m11)

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(pid)))

	testing.SetOriginCaller(m2)

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(pid)))

	testing.SetOriginCaller(m3)

	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, dao.ProposalID(pid)))

	testing.SetOriginCaller(m111)

	// Same effect as:
	// dao.MustVoteOnProposal(dao.VoteRequest{
	// 	Option:     dao.YesVote,
	// 	ProposalID: dao.ProposalID(pid),
	// })
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")

	urequire.Equal(t, true, contains(dao.Render(cross(cur), "8"), "**Proposal is open for votes**"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), "8"), "68.42105263157895%"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), "8"), "0%"))

	accepted := dao.ExecuteProposal(cross(cur), dao.ProposalID(pid))
	urequire.Equal(t, true, accepted)
	urequire.Equal(t, true, contains(dao.Render(cross(cur), "8"), "**PROPOSAL HAS BEEN ACCEPTED**"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), "8"), "YES PERCENT: 68.42105263157895%"))
}

func TestAbstainVote(cur realm, t *testing.T) {
	loadMembers(cur)

	portfolio := "# This is my portfolio:\n\n- THINGS"

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	nm1 := testutils.TestAddress("nm1")
	proposalRequest := NewAddMemberRequest(cur, nm1, memberstore.T2, portfolio)

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))
	pid := dao.MustCreateProposal(cross(cur), proposalRequest)

	// m1 votes abstain
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "ABSTAIN")

	// Other members vote YES to reach supermajority
	testing.SetOriginCaller(m11)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))

	testing.SetOriginCaller(m111)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))

	testing.SetOriginCaller(m1111)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))

	testing.SetOriginCaller(m2)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))

	testing.SetOriginCaller(m3)
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))

	// Verify render shows correct percentages
	urequire.Equal(t, true, contains(dao.Render(cross(cur), pid.String()), "YES PERCENT: 81.25%"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), pid.String()), "NO PERCENT: 0%"))
	urequire.Equal(t, true, contains(dao.Render(cross(cur), pid.String()), "ABSTAIN PERCENT: 18.75%"))

	// Supermajority reached despite one abstain voter
	accepted := dao.ExecuteProposal(cross(cur), pid)
	urequire.Equal(t, true, accepted)
	urequire.Equal(t, true, contains(dao.Render(cross(cur), pid.String()), "**PROPOSAL HAS BEEN ACCEPTED**"))

	// Abstain voter appears in vote list
	urequire.Equal(t, true, contains(dao.Render(cross(cur), fmt.Sprintf("%v/votes", int64(pid))), "ABSTAIN"))
}

func contains(s, substr string) bool {
	return strings.Index(s, substr) >= 0
}
