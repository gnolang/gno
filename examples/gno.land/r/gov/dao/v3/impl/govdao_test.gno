package impl

import (
	"fmt"
	"strings"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/urequire"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
)

func init() {
	loadMembers()

	dao.UpdateImpl(cross, dao.UpdateRequest{
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
	mstore := memberstore.Get()
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
	loadMembers()

	portfolio := "# This is my portfolio:\n\n- THINGS"

	testing.SetOriginCaller(noMember)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	nm1 := testutils.TestAddress("nm1")

	urequire.AbortsWithMessage(t, "Only T1 and T2 members can be added by proposal. To add a T3 member use AddMember function directly.", func(cur realm) {
		dao.MustCreateProposal(cross, NewAddMemberRequest(cur, nm1, memberstore.T3, portfolio))
	})

	urequire.AbortsWithMessage(t, "proposer is not a member", func(cur realm) {
		dao.MustCreateProposal(cross, NewAddMemberRequest(cur, nm1, memberstore.T2, portfolio))
	})

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	proposalRequest := NewAddMemberRequest(cur, nm1, memberstore.T2, portfolio)

	testing.SetOriginCaller(m1)
	pid := dao.MustCreateProposal(cross, proposalRequest)
	urequire.Equal(t, int(pid), 0)

	// m1 votes yes because that member is interested on it
	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(0),
	})

	testing.SetOriginCaller(m11)

	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.NoVote,
		ProposalID: dao.ProposalID(0),
	})

	testing.SetOriginCaller(m2)

	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.NoVote,
		ProposalID: dao.ProposalID(0),
	})

	testing.SetOriginCaller(m3)

	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.NoVote,
		ProposalID: dao.ProposalID(0),
	})

	testing.SetOriginCaller(m4)

	urequire.AbortsWithMessage(t, "member on specified tier is not allowed to vote on this proposal", func() {
		dao.MustVoteOnProposal(cross, dao.VoteRequest{
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
	dao.MustVoteOnProposalSimple(cross, 0, "NO")

	urequire.Equal(t, true, strings.Contains(dao.Render(""), "Prop #0 - New T2 Member Proposal"))
	// urequire.Equal(t, true, strings.Contains(dao.Render(""), "Author: "+m1.String()))

	urequire.AbortsWithMessage(t, "proposal didn't reach supermajority yet: 66.66", func() {
		dao.ExecuteProposal(cross, dao.ProposalID(0))
	})

	testing.SetOriginCaller(m1111)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.NoVote,
		ProposalID: dao.ProposalID(0),
	})

	accepted := dao.ExecuteProposal(cross, dao.ProposalID(0))
	urequire.Equal(t, false, accepted)

	urequire.Equal(t, true, contains(dao.Render("0"), "**PROPOSAL HAS BEEN DENIED**"))
	urequire.Equal(t, true, contains(dao.Render("0"), "NO PERCENT: 68.42105263157895%"))
}

func TestProposalPagination(cur realm, t *testing.T) {
	loadMembers()
	portfolio := "### This is my portfolio:\n\n- THINGS"

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	nm1 := testutils.TestAddress("nm1")

	var pid dao.ProposalID

	proposalRequest := NewAddMemberRequest(cur, nm1, memberstore.T2, portfolio)

	testing.SetOriginCaller(m1)
	pid = dao.MustCreateProposal(cross, proposalRequest)

	// TODO: tests keep the same vm state: https://github.com/gnolang/gno/issues/1982
	urequire.Equal(t, 1, int(pid))

	pid = dao.MustCreateProposal(cross, proposalRequest)
	urequire.Equal(t, 2, int(pid))

	pid = dao.MustCreateProposal(cross, proposalRequest)
	urequire.Equal(t, 3, int(pid))

	pid = dao.MustCreateProposal(cross, proposalRequest)
	urequire.Equal(t, 4, int(pid))

	pid = dao.MustCreateProposal(cross, proposalRequest)
	urequire.Equal(t, 5, int(pid))

	pid = dao.MustCreateProposal(cross, proposalRequest)
	urequire.Equal(t, 6, int(pid))

	fmt.Println(dao.Render(""))
	urequire.Equal(t, true, contains(dao.Render(""), "### [Prop #6 - New T2 Member Proposal](/r/gov/dao:6)"))
	urequire.Equal(t, true, contains(dao.Render(""), "### [Prop #5 - New T2 Member Proposal](/r/gov/dao:5)"))
	urequire.Equal(t, true, contains(dao.Render(""), "### [Prop #4 - New T2 Member Proposal](/r/gov/dao:4)"))
	urequire.Equal(t, true, contains(dao.Render(""), "### [Prop #3 - New T2 Member Proposal](/r/gov/dao:3)"))
	urequire.Equal(t, true, contains(dao.Render(""), "### [Prop #2 - New T2 Member Proposal](/r/gov/dao:2)"))

	urequire.Equal(t, true, contains(dao.Render("?page=2"), "### [Prop #1 - New T2 Member Proposal](/r/gov/dao:1)"))
	urequire.Equal(t, true, contains(dao.Render("?page=2"), "### [Prop #0 - New T2 Member Proposal](/r/gov/dao:0)"))
}

func TestUpgradeDaoImplementation(t *testing.T) {
	loadMembers()

	testing.SetOriginCaller(noMember)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	urequire.PanicsWithMessage(t, "proposer is not a member", func() {
		NewUpgradeDaoImplRequest(govDAO, "gno.land/r/gov/dao/v4/impl", "Something happened and we have to fix it.")
	})

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	preq := NewUpgradeDaoImplRequest(govDAO, "gno.land/r/gov/dao/v4/impl", "Something happened and we have to fix it.")

	testing.SetOriginCaller(m1)
	pid := dao.MustCreateProposal(cross, preq)
	urequire.Equal(t, int(pid), 7)

	// m1 votes yes because that member is interested on it
	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(pid),
	})

	testing.SetOriginCaller(m11)

	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(pid),
	})

	testing.SetOriginCaller(m2)

	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(pid),
	})

	testing.SetOriginCaller(m3)

	dao.MustVoteOnProposal(cross, dao.VoteRequest{
		Option:     dao.YesVote,
		ProposalID: dao.ProposalID(pid),
	})

	testing.SetOriginCaller(m111)

	// Same effect as:
	// dao.MustVoteOnProposal(dao.VoteRequest{
	// 	Option:     dao.YesVote,
	// 	ProposalID: dao.ProposalID(pid),
	// })
	dao.MustVoteOnProposalSimple(cross, int64(pid), "YES")

	urequire.Equal(t, true, contains(dao.Render("7"), "**Proposal is open for votes**"))
	urequire.Equal(t, true, contains(dao.Render("7"), "68.42105263157895%"))
	urequire.Equal(t, true, contains(dao.Render("7"), "0%"))

	accepted := dao.ExecuteProposal(cross, dao.ProposalID(pid))
	urequire.Equal(t, true, accepted)
	urequire.Equal(t, true, contains(dao.Render("7"), "**PROPOSAL HAS BEEN ACCEPTED**"))
	urequire.Equal(t, true, contains(dao.Render("7"), "YES PERCENT: 68.42105263157895%"))
}

func contains(s, substr string) bool {
	return strings.Index(s, substr) >= 0
}
