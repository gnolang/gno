package impl

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
	"gno.land/r/gov/dao/proxy"
	"gno.land/r/gov/dao/v3/memberstore"
)

var (
	m1   = testutils.TestAddress("m1")
	m11  = testutils.TestAddress("m1.1")
	m111 = testutils.TestAddress("m1.1.1")
	m2   = testutils.TestAddress("m2")
	m3   = testutils.TestAddress("m3")
	m4   = testutils.TestAddress("m4")
	m5   = testutils.TestAddress("m5")
	m6   = testutils.TestAddress("m6")

	noMember = testutils.TestAddress("nm1")
)

func TestCreateProposalAndVote(t *testing.T) {
	proxy.UpdateImpl(proxy.UpdateRequest{
		DAO:         govDAO,
		AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl"},
	})

	memberstore.Get().SetMember(memberstore.T1, m1, memberByTier(memberstore.T1))
	memberstore.Get().SetMember(memberstore.T1, m11, memberByTier(memberstore.T1))
	memberstore.Get().SetMember(memberstore.T1, m111, memberByTier(memberstore.T1))

	memberstore.Get().SetMember(memberstore.T2, m2, memberByTier(memberstore.T2))
	memberstore.Get().SetMember(memberstore.T2, m3, memberByTier(memberstore.T2))
	memberstore.Get().SetMember(memberstore.T3, m4, memberByTier(memberstore.T3))
	memberstore.Get().SetMember(memberstore.T3, m5, memberByTier(memberstore.T3))
	memberstore.Get().SetMember(memberstore.T3, m6, memberByTier(memberstore.T3))

	std.TestSetOrigCaller(noMember)
	std.TestSetRealm(std.NewUserRealm(noMember))

	nm1 := testutils.TestAddress("nm1")

	urequire.PanicsWithMessage(t, "proposer is not a member", func() {
		proxy.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T3))
	})

	println(proxy.Render(""))

	std.TestSetOrigCaller(m1)
	std.TestSetRealm(std.NewUserRealm(m1))

	pid := proxy.MustCreateProposal(NewAddMemberRequest(nm1, memberstore.T3))
	urequire.Equal(t, int(pid), 0)

	// m1 votes yes because that member is interested on it
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.YesVote,
		ProposalID: proxy.ProposalID(0),
	})

	println(proxy.Render(""))

	std.TestSetOrigCaller(m11)
	std.TestSetRealm(std.NewUserRealm(m11))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	std.TestSetOrigCaller(m2)
	std.TestSetRealm(std.NewUserRealm(m2))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	println(proxy.Render(""))

	std.TestSetOrigCaller(m3)
	std.TestSetRealm(std.NewUserRealm(m3))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	std.TestSetOrigCaller(m4)
	std.TestSetRealm(std.NewUserRealm(m4))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	std.TestSetOrigCaller(m5)
	std.TestSetRealm(std.NewUserRealm(m5))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	println(proxy.Render(""))

	std.TestSetOrigCaller(m6)
	std.TestSetRealm(std.NewUserRealm(m6))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	println(proxy.Render(""))


	std.TestSetOrigCaller(m111)
	std.TestSetRealm(std.NewUserRealm(m111))
	proxy.MustVoteOnProposal(proxy.VoteRequest{
		Option:     proxy.NoVote,
		ProposalID: proxy.ProposalID(0),
	})

	println(proxy.Render(""))

	ExecuteProposal(proxy.ProposalID(0))

	println(proxy.Render(""))

}

var executed bool

func callback(proxy.Metadata) error {
	println("PROPOSAL EXECUTED!")
	executed = true
	return nil
}
