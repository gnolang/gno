package impl

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
	"gno.land/r/gov/dao/proxy"
	"gno.land/r/gov/dao/v3/memberstore"
)

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

// TODO: This is not needed right now. Tests keep the same vm state: https://github.com/gnolang/gno/issues/1982
func loadMembers() {
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

	std.TestSetOriginCaller(noMember)
	std.TestSetRealm(std.NewUserRealm(noMember))

	nm1 := testutils.TestAddress("nm1")

	urequire.PanicsWithMessage(t, "Only T1 and T2 members must be added by proposal. To add a T3 member use AddMember function directly.", func() {
		proxy.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T3, portfolio))
	})

	urequire.PanicsWithMessage(t, "proposer is not a member", func() {
		proxy.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	})

	std.TestSetOriginCaller(m1)
	std.TestSetRealm(std.NewUserRealm(m1))

	pid := proxy.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, int(pid), 0)

	// m1 votes yes because that member is interested on it
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.YesVote,
		ProposalID: proxy.ProposalID(0),
	})

	std.TestSetOriginCaller(m11)
	std.TestSetRealm(std.NewUserRealm(m11))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	std.TestSetOriginCaller(m2)
	std.TestSetRealm(std.NewUserRealm(m2))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	std.TestSetOriginCaller(m3)
	std.TestSetRealm(std.NewUserRealm(m3))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	std.TestSetOriginCaller(m4)
	std.TestSetRealm(std.NewUserRealm(m4))
	urequire.PanicsWithMessage(t, "member on specified tier is not allowed to vote on this proposal", func() {
		proxy.MustVoteOnProposal(proxy.VoteRequest{
			Option:     proxy.NoVote,
			ProposalID: proxy.ProposalID(0),
		})
	})

	std.TestSetOriginCaller(m111)
	std.TestSetRealm(std.NewUserRealm(m111))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	urequire.Equal(t, true, contains(proxy.Render(""), "Proposal open for votes"))
	urequire.Equal(t, true, contains(proxy.Render(""), "15.789473684210526%"))
	urequire.Equal(t, true, contains(proxy.Render(""), "52.63157894736842%"))

	urequire.PanicsWithMessage(t, "proposal didn't reached supermajority yet: 66", func() {
		ExecuteProposal(proxy.ProposalID(0))
	})

	std.TestSetOriginCaller(m1111)
	std.TestSetRealm(std.NewUserRealm(m1111))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	accepted := ExecuteProposal(proxy.ProposalID(0))
	urequire.Equal(t, false, accepted)

	urequire.Equal(t, true, contains(proxy.Render(""), "**PROPOSAL HAS BEEN DENIED**"))
	urequire.Equal(t, true, contains(proxy.Render(""), "NO PERCENT: 68.42105263157895%"))
}

func TestProposalPagination(t *testing.T) {
	loadMembers()
	portfolio := "### This is my portfolio:\n\n- THINGS"

	std.TestSetOriginCaller(m1)
	std.TestSetRealm(std.NewUserRealm(m1))

	nm1 := testutils.TestAddress("nm1")

	var pid proxy.ProposalID

	pid = proxy.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	// TODO: tests keep the same vm state: https://github.com/gnolang/gno/issues/1982
	urequire.Equal(t, int(pid), 1)

	pid = proxy.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, int(pid), 2)

	pid = proxy.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, int(pid), 3)

	pid = proxy.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, int(pid), 4)

	pid = proxy.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, int(pid), 5)

	pid = proxy.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T2, portfolio))
	urequire.Equal(t, int(pid), 6)

	urequire.Equal(t, true, contains(proxy.Render(""), "## Proposal with id: 6"))
	urequire.Equal(t, true, contains(proxy.Render(""), "## Proposal with id: 5"))
	urequire.Equal(t, true, contains(proxy.Render(""), "## Proposal with id: 4"))

	urequire.Equal(t, true, contains(proxy.Render("/?page=2"), "## Proposal with id: 3"))
	urequire.Equal(t, true, contains(proxy.Render("/?page=2"), "## Proposal with id: 2"))
	urequire.Equal(t, true, contains(proxy.Render("/?page=2"), "## Proposal with id: 1"))

	urequire.Equal(t, true, contains(proxy.Render("/?page=3"), "## Proposal with id: 0"))
}

func contains(s, substr string) bool {
	return strings.Index(s, substr) >= 0
}
