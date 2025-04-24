package impl

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
)

func init() {
	loadMembers()
	dao.UpdateImpl(dao.UpdateRequest{
		DAO:         govDAO,
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})
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

func loadMembers() {
	// This is needed because state is saved between unit tests,
	// and we want to avoid having real members used on tests
	memberstore.Get().DeleteAll()

	memberstore.Get().SetTier(memberstore.T1)
	memberstore.Get().SetTier(memberstore.T2)
	memberstore.Get().SetTier(memberstore.T3)

	memberstore.Get().SetMember(memberstore.T1, m1, memberByTier(memberstore.T1))
	memberstore.Get().SetMember(memberstore.T1, m11, memberByTier(memberstore.T1))
	memberstore.Get().SetMember(memberstore.T1, m111, memberByTier(memberstore.T1))
	memberstore.Get().SetMember(memberstore.T1, m1111, memberByTier(memberstore.T1))

	memberstore.Get().SetMember(memberstore.T2, m2, memberByTier(memberstore.T2))
	memberstore.Get().SetMember(memberstore.T2, m3, memberByTier(memberstore.T2))
	memberstore.Get().SetMember(memberstore.T3, m4, memberByTier(memberstore.T3))
	memberstore.Get().SetMember(memberstore.T3, m5, memberByTier(memberstore.T3))
	memberstore.Get().SetMember(memberstore.T3, m6, memberByTier(memberstore.T3))
}

func TestCreateProposalAndVote(t *testing.T) {
	loadMembers()

	portfolio := "# This is my portfolio:\n\n- THINGS"

	testing.SetOriginCaller(noMember)

	nm1 := testutils.TestAddress("nm1")

	urequire.PanicsWithMessage(t, "Only T1 and T2 members can be added by proposal. To add a T3 member use AddMember function directly.", func() {
		dao.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T3, portfolio))
	})

	urequire.PanicsWithMessage(t, "proposer is not a member", func() {
		dao.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	})

	testing.SetOriginCaller(m1)
	testing.SetRealm(std.NewUserRealm(m1))

	pid := dao.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, int(pid), 0)

	// m1 votes yes because that member is interested on it
	dao.MustVoteOnProposal(dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(0),
	})

	testing.SetOriginCaller(m11)

	dao.MustVoteOnProposal(dao.VoteRequest{
		Option:     dao.NoVote,
		ProposalID: dao.ProposalID(0),
	})

	testing.SetOriginCaller(m2)

	dao.MustVoteOnProposal(dao.VoteRequest{
		Option:     dao.NoVote,
		ProposalID: dao.ProposalID(0),
	})

	testing.SetOriginCaller(m3)

	dao.MustVoteOnProposal(dao.VoteRequest{
		Option:     dao.NoVote,
		ProposalID: dao.ProposalID(0),
	})

	testing.SetOriginCaller(m4)

	urequire.PanicsWithMessage(t, "member on specified tier is not allowed to vote on this proposal", func() {
		dao.MustVoteOnProposal(dao.VoteRequest{
			Option:     dao.NoVote,
			ProposalID: dao.ProposalID(0),
		})
	})

	testing.SetOriginCaller(m111)

	// Same effect as:
	// dao.MustVoteOnProposal(dao.VoteRequest{
	// 	Option:     dao.NoVote,
	// 	ProposalID: dao.ProposalID(0),
	// })
	dao.MustVoteOnProposalSimple(0, "NO")

	urequire.Equal(t, true, contains(dao.Render(""), "Proposal open for votes"))
	urequire.Equal(t, true, contains(dao.Render(""), "15.789473684210526%"))
	urequire.Equal(t, true, contains(dao.Render(""), "52.63157894736842%"))

	urequire.PanicsWithMessage(t, "proposal didn't reach supermajority yet: 66", func() {
		dao.ExecuteProposal(dao.ProposalID(0))
	})

	testing.SetOriginCaller(m1111)

	dao.MustVoteOnProposal(dao.VoteRequest{
		Option:     dao.NoVote,
		ProposalID: dao.ProposalID(0),
	})

	accepted := dao.ExecuteProposal(dao.ProposalID(0))
	urequire.Equal(t, false, accepted)

	urequire.Equal(t, true, contains(dao.Render(""), "**PROPOSAL HAS BEEN DENIED**"))
	urequire.Equal(t, true, contains(dao.Render(""), "NO PERCENT: 68.42105263157895%"))
}

func TestProposalPagination(t *testing.T) {
	loadMembers()
	portfolio := "### This is my portfolio:\n\n- THINGS"

	testing.SetOriginCaller(m1)

	nm1 := testutils.TestAddress("nm1")

	var pid dao.ProposalID

	pid = dao.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	// TODO: tests keep the same vm state: https://github.com/gnolang/gno/issues/1982
	urequire.Equal(t, 1, int(pid))

	pid = dao.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, 2, int(pid))

	pid = dao.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, 3, int(pid))

	pid = dao.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, 4, int(pid))

	pid = dao.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, 5, int(pid))

	pid = dao.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, 6, int(pid))

	urequire.Equal(t, true, contains(dao.Render(""), "## Proposal with id: 6"))
	urequire.Equal(t, true, contains(dao.Render(""), "## Proposal with id: 5"))
	urequire.Equal(t, true, contains(dao.Render(""), "## Proposal with id: 4"))

	urequire.Equal(t, true, contains(dao.Render("/?page=2"), "## Proposal with id: 3"))
	urequire.Equal(t, true, contains(dao.Render("/?page=2"), "## Proposal with id: 2"))
	urequire.Equal(t, true, contains(dao.Render("/?page=2"), "## Proposal with id: 1"))

	urequire.Equal(t, true, contains(dao.Render("/?page=3"), "## Proposal with id: 0"))
}

func TestUpgradeDaoImplementation(t *testing.T) {
	loadMembers()

	testing.SetOriginCaller(noMember)

	nm1 := testutils.TestAddress("nm1")

	urequire.PanicsWithMessage(t, "proposer is not a member", func() {
		preq := NewUpgradeDaoImplRequest(govDAO, "gno.land/r/gov/dao/v4/impl", "Something happened and we have to fix it.")
		dao.MustCreateProposal(preq)
	})

	testing.SetOriginCaller(m1)
	testing.SetRealm(std.NewUserRealm(m1))

	preq := NewUpgradeDaoImplRequest(govDAO, "gno.land/r/gov/dao/v4/impl", "Something happened and we have to fix it.")
	pid := dao.MustCreateProposal(preq)
	urequire.Equal(t, int(pid), 7)

	// m1 votes yes because that member is interested on it
	dao.MustVoteOnProposal(dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(pid),
	})

	testing.SetOriginCaller(m11)

	dao.MustVoteOnProposal(dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(pid),
	})

	testing.SetOriginCaller(m2)

	dao.MustVoteOnProposal(dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(pid),
	})

	testing.SetOriginCaller(m3)

	dao.MustVoteOnProposal(dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(pid),
	})

	testing.SetOriginCaller(m4)

	testing.SetOriginCaller(m111)

	// Same effect as:
	// dao.MustVoteOnProposal(dao.VoteRequest{
	// 	Option:     dao.YesVote,
	// 	ProposalID: dao.ProposalID(pid),
	// })
	dao.MustVoteOnProposalSimple(int64(pid), "YES")

	urequire.Equal(t, true, contains(dao.Render("7"), "Proposal open for votes"))
	urequire.Equal(t, true, contains(dao.Render("7"), "68.42105263157895%"))
	urequire.Equal(t, true, contains(dao.Render("7"), "0%"))

	testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/loader")) // check if the proposal will be executed even if the execution was made from another package
	accepted := dao.ExecuteProposal(dao.ProposalID(pid))
	urequire.Equal(t, true, accepted)
	urequire.Equal(t, true, contains(dao.Render("7"), "**PROPOSAL HAS BEEN ACCEPTED**"))
	urequire.Equal(t, true, contains(dao.Render("7"), "YES PERCENT: 68.42105263157895%"))
}

func contains(s, substr string) bool {
	return strings.Index(s, substr) >= 0
}
