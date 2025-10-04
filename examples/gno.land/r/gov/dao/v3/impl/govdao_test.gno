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
	testing.SetRealm(testing.NewUserRealm(m1))
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

func TestExecutorCreationRealm(cur realm, t *testing.T) {
	loadMembers()

	// Test that executor creation realm is captured correctly
	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/template/contract"))

	// Create executor in the template contract realm
	executor := dao.NewSimpleExecutor(func(realm) error { return nil }, "Test executor from template")

	proposalRequest := dao.NewProposalRequest(
		"Test Proposal",
		"This proposal tests executor creation realm tracking",
		executor,
	)

	// Create proposal from user realm (user can call DAO directly)
	testing.SetRealm(testing.NewUserRealm(m1))
	pid := dao.MustCreateProposal(cross, proposalRequest)

	// Get the proposal
	prop := dao.MustGetProposal(cross, pid)

	// Verify the author is m1
	urequire.Equal(t, m1, prop.Author())

	// Verify the executor creation realm is captured correctly
	urequire.Equal(t, "gno.land/r/template/contract", prop.ExecutorCreationRealm())

	// Check that it's displayed in the individual proposal render output
	individualRendered := dao.Render(pid.String())
	urequire.Equal(t, true, contains(individualRendered, "Executor created in: gno.land/r/template/contract"))
	urequire.Equal(t, true, contains(individualRendered, "Test executor from template"))

	// Also verify the main content is there
	urequire.Equal(t, true, contains(individualRendered, "Test Proposal"))
	urequire.Equal(t, true, contains(individualRendered, "This proposal tests executor creation realm tracking"))
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
	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross, proposalRequest)

	// TODO: tests keep the same vm state: https://github.com/gnolang/gno/issues/1982
	urequire.Equal(t, 2, int(pid))

	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross, proposalRequest)
	urequire.Equal(t, 3, int(pid))

	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross, proposalRequest)
	urequire.Equal(t, 4, int(pid))

	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross, proposalRequest)
	urequire.Equal(t, 5, int(pid))

	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross, proposalRequest)
	urequire.Equal(t, 6, int(pid))

	testing.SetRealm(testing.NewUserRealm(m1))
	pid = dao.MustCreateProposal(cross, proposalRequest)
	urequire.Equal(t, 7, int(pid))

	fmt.Println(dao.Render(""))
	urequire.Equal(t, true, contains(dao.Render(""), "### [Prop #7 - New T2 Member Proposal](/r/gov/dao:7)"))
	urequire.Equal(t, true, contains(dao.Render(""), "### [Prop #6 - New T2 Member Proposal](/r/gov/dao:6)"))
	urequire.Equal(t, true, contains(dao.Render(""), "### [Prop #5 - New T2 Member Proposal](/r/gov/dao:5)"))
	urequire.Equal(t, true, contains(dao.Render(""), "### [Prop #4 - New T2 Member Proposal](/r/gov/dao:4)"))
	urequire.Equal(t, true, contains(dao.Render(""), "### [Prop #3 - New T2 Member Proposal](/r/gov/dao:3)"))

	urequire.Equal(t, true, contains(dao.Render("?page=2"), "### [Prop #2 - New T2 Member Proposal](/r/gov/dao:2)"))
	urequire.Equal(t, true, contains(dao.Render("?page=2"), "### [Prop #1 - Test Proposal](/r/gov/dao:1)"))
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
	testing.SetRealm(testing.NewUserRealm(m1))
	pid := dao.MustCreateProposal(cross, preq)
	urequire.Equal(t, int(pid), 8)

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

	urequire.Equal(t, true, contains(dao.Render("8"), "**Proposal is open for votes**"))
	urequire.Equal(t, true, contains(dao.Render("8"), "68.42105263157895%"))
	urequire.Equal(t, true, contains(dao.Render("8"), "0%"))

	accepted := dao.ExecuteProposal(cross, dao.ProposalID(pid))
	urequire.Equal(t, true, accepted)
	urequire.Equal(t, true, contains(dao.Render("8"), "**PROPOSAL HAS BEEN ACCEPTED**"))
	urequire.Equal(t, true, contains(dao.Render("8"), "YES PERCENT: 68.42105263157895%"))
}

func TestGetMemberActivity(t *testing.T) {
	loadMembers()

	// First, test when no proposals exist
	activity, available := govDAO.GetMemberActivity(m1)
	urequire.Equal(t, uint32(0), activity)
	urequire.Equal(t, uint32(0), available) // No proposals exist yet

	// Test non-member (should not panic and return 0,0)
	activity, available = govDAO.GetMemberActivity(noMember)
	urequire.Equal(t, uint32(0), activity)
	urequire.Equal(t, uint32(0), available) // Non-member, no proposals

	// Get m1 member for time calculations
	m1Member, _ := memberstore.Get().GetMember(m1)

	// Create a mock proposal status that was created before member joined
	ps1 := &proposalStatus{
		YesVotes:           memberstore.NewMembersByTier(),
		NoVotes:            memberstore.NewMembersByTier(),
		AllVotes:           memberstore.NewMembersByTier(),
		Accepted:           false,
		Denied:             false,
		TiersAllowedToVote: []string{memberstore.T1, memberstore.T2},
		TotalPower:         100.0,
		CreatedAt:          m1Member.JoinTime.Add(-24 * 60 * 60 * 1000000000), // 1 day before member joined
		ClosedAt:           m1Member.JoinTime.Add(-12 * 60 * 60 * 1000000000), // 12 hours before member joined
	}
	ps1.YesVotes.SetTier(memberstore.T1)
	ps1.YesVotes.SetTier(memberstore.T2)
	ps1.NoVotes.SetTier(memberstore.T1)
	ps1.NoVotes.SetTier(memberstore.T2)
	ps1.AllVotes.SetTier(memberstore.T1)
	ps1.AllVotes.SetTier(memberstore.T2)

	// Create an open proposal created after member joined
	ps2 := &proposalStatus{
		YesVotes:           memberstore.NewMembersByTier(),
		NoVotes:            memberstore.NewMembersByTier(),
		AllVotes:           memberstore.NewMembersByTier(),
		Accepted:           false,
		Denied:             false,
		TiersAllowedToVote: []string{memberstore.T1, memberstore.T2},
		TotalPower:         100.0,
		CreatedAt:          m1Member.JoinTime.Add(24 * 60 * 60 * 1000000000), // 1 day after member joined
		ClosedAt:           m1Member.JoinTime.Add(48 * 60 * 60 * 1000000000), // 2 days after member joined (not closed yet)
	}
	ps2.YesVotes.SetTier(memberstore.T1)
	ps2.YesVotes.SetTier(memberstore.T2)
	ps2.NoVotes.SetTier(memberstore.T1)
	ps2.NoVotes.SetTier(memberstore.T2)
	ps2.AllVotes.SetTier(memberstore.T1)
	ps2.AllVotes.SetTier(memberstore.T2)

	// Add m1's vote to the second proposal
	ps2.AllVotes.SetMember(memberstore.T1, m1, m1Member)

	// Create a closed proposal (accepted) created after member joined
	ps3 := &proposalStatus{
		YesVotes:           memberstore.NewMembersByTier(),
		NoVotes:            memberstore.NewMembersByTier(),
		AllVotes:           memberstore.NewMembersByTier(),
		Accepted:           true,
		Denied:             false,
		TiersAllowedToVote: []string{memberstore.T1, memberstore.T2},
		TotalPower:         100.0,
		CreatedAt:          m1Member.JoinTime.Add(12 * 60 * 60 * 1000000000), // 12 hours after member joined
		ClosedAt:           m1Member.JoinTime.Add(36 * 60 * 60 * 1000000000), // 36 hours after member joined
	}
	ps3.YesVotes.SetTier(memberstore.T1)
	ps3.YesVotes.SetTier(memberstore.T2)
	ps3.NoVotes.SetTier(memberstore.T1)
	ps3.NoVotes.SetTier(memberstore.T2)
	ps3.AllVotes.SetTier(memberstore.T1)
	ps3.AllVotes.SetTier(memberstore.T2)

	// Add m1's vote to the third proposal
	ps3.AllVotes.SetMember(memberstore.T1, m1, m1Member)

	// Create another open proposal without m1's vote
	ps4 := &proposalStatus{
		YesVotes:           memberstore.NewMembersByTier(),
		NoVotes:            memberstore.NewMembersByTier(),
		AllVotes:           memberstore.NewMembersByTier(),
		Accepted:           false,
		Denied:             false,
		TiersAllowedToVote: []string{memberstore.T1, memberstore.T2},
		TotalPower:         100.0,
		CreatedAt:          m1Member.JoinTime.Add(6 * 60 * 60 * 1000000000),  // 6 hours after member joined
		ClosedAt:           m1Member.JoinTime.Add(72 * 60 * 60 * 1000000000), // 3 days after member joined (not closed yet)
	}
	ps4.YesVotes.SetTier(memberstore.T1)
	ps4.YesVotes.SetTier(memberstore.T2)
	ps4.NoVotes.SetTier(memberstore.T1)
	ps4.NoVotes.SetTier(memberstore.T2)
	ps4.AllVotes.SetTier(memberstore.T1)
	ps4.AllVotes.SetTier(memberstore.T2)

	// Add proposals to govDAO's proposal status store
	govDAO.pss.Set("0", *ps1) // Proposal created before member joined
	govDAO.pss.Set("1", *ps2) // Open proposal with m1's vote
	govDAO.pss.Set("2", *ps3) // Closed proposal with m1's vote
	govDAO.pss.Set("3", *ps4) // Open proposal without m1's vote

	// Test with 4 proposals:
	// ps1: before joining (ignored)
	// ps2: after joining, open, m1 voted (counts: available + activity)
	// ps3: after joining, closed, m1 voted (counts: available only)
	// ps4: after joining, open, no vote (counts: available only)
	// Expected: activity=1, available=3

	activity, available = govDAO.GetMemberActivity(m1)
	urequire.Equal(t, uint32(3), available) // 3 proposals created after member joined
	urequire.Equal(t, uint32(1), activity)  // 1 vote on open proposals

	// For activity, let's first test with a simpler expectation
	// Since we're having issues with vote detection, let's debug step by step
	// For now, let's accept any result for activity and focus on getting available right first
	// urequire.Equal(t, uint32(1), activity)

	// Test m2 (no votes): should see all proposals but no activity
	activity, available = govDAO.GetMemberActivity(m2)
	urequire.Equal(t, uint32(0), activity)
	urequire.Equal(t, uint32(3), available)

	// Test T3 member (can't vote on T1/T2 proposals)
	activity, available = govDAO.GetMemberActivity(m4)
	urequire.Equal(t, uint32(0), activity)
	urequire.Equal(t, uint32(3), available)

	// Test non-member (should not panic and return 0,0)
	activity, available = govDAO.GetMemberActivity(noMember)
	urequire.Equal(t, uint32(0), activity)
	urequire.Equal(t, uint32(0), available)

	// Test with an address that was never registered as a member
	unknownAddr := testutils.TestAddress("unknown")
	activity, available = govDAO.GetMemberActivity(unknownAddr)
	urequire.Equal(t, uint32(0), activity)
	urequire.Equal(t, uint32(0), available)

	// Test summary:
	// ✓ Works with no proposals
	// ✓ Handles non-members safely (no crashes)
	// ✓ Counts votes correctly for different scenarios
	// ✓ Works with all member tiers (T1, T2, T3)
	// ✓ Handles unknown addresses
	//
	// How GetMemberActivity works:
	// - activity = votes on open proposals
	// - available = all proposals since joining
}

func TestActivityPage(t *testing.T) {
	loadMembers()

	// Set the correct realm context for testing
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/v3/impl"))

	// Clean the proposal store before testing
	govDAO.pss = NewProposalsStatuses()

	// Test activity page with no proposals using the dao proxy (like other tests)
	activityPage := dao.Render("activity")

	urequire.Equal(t, true, contains(activityPage, "# GovDAO Activity"))

	// Check if members are being loaded properly - test some basic member activity display
	urequire.Equal(t, true, len(activityPage) > 50) // Should have some content

	// Create some mock proposals to test activity display
	m1Member, _ := memberstore.Get().GetMember(m1)

	// Create an open proposal with m1's vote
	ps1 := &proposalStatus{
		YesVotes:           memberstore.NewMembersByTier(),
		NoVotes:            memberstore.NewMembersByTier(),
		AllVotes:           memberstore.NewMembersByTier(),
		Accepted:           false,
		Denied:             false,
		TiersAllowedToVote: []string{memberstore.T1, memberstore.T2},
		TotalPower:         100.0,
		CreatedAt:          m1Member.JoinTime.Add(24 * 60 * 60 * 1000000000),  // After member joined
		ClosedAt:           m1Member.JoinTime.Add(-48 * 60 * 60 * 1000000000), // Not closed (set to past to indicate not closed yet)
	}
	ps1.YesVotes.SetTier(memberstore.T1)
	ps1.YesVotes.SetTier(memberstore.T2)
	ps1.NoVotes.SetTier(memberstore.T1)
	ps1.NoVotes.SetTier(memberstore.T2)
	ps1.AllVotes.SetTier(memberstore.T1)
	ps1.AllVotes.SetTier(memberstore.T2)

	// Add m1's vote
	ps1.AllVotes.SetMember(memberstore.T1, m1, m1Member)

	// Add to proposal store
	govDAO.pss.Set("test1", *ps1)

	// Test activity page with proposals
	activityPage = dao.Render("activity")

	println(activityPage)
	// Clean up after test
	govDAO.pss = NewProposalsStatuses()
}

func contains(s, substr string) bool {
	return strings.Index(s, substr) >= 0
}
