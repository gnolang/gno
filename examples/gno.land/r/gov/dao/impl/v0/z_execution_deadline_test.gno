package impl

import (
	"strconv"
	"testing"

	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gov/dao"
)

// Named z_* so it sorts after govdao_test.gno, whose tests assert hard-coded
// proposal ids. Each test keeps its calls in the test body: testing.SetRealm
// does not reach a cross-call made from inside a closure.

func TestProposalExecutionDeadline(cur realm, t *testing.T) {
	loadMembers(cur)

	executed := 0
	execute := func(realm) error {
		executed++
		return nil
	}
	atDeadline := dao.NewSimpleExecutor(0, cur, execute, "")
	pastDeadline := dao.NewSimpleExecutor(0, cur, execute, "")
	voters := []address{m1, m11, m111, m1111, m2, m3, m4, m5, m6}

	// At the deadline the proposal still executes.
	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))
	pid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest("At the deadline", "", atDeadline))
	for _, m := range voters {
		testing.SetOriginCaller(m)
		dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	}
	testing.SkipHeights(executionDeadlineBlocks)
	urequire.Equal(t, true, dao.ExecuteProposal(cross(cur), pid))
	urequire.Equal(t, 1, executed)

	// One block past it, the proposal closes as denied without executing.
	testing.SetOriginCaller(m1)
	pid = dao.MustCreateProposal(cross(cur), dao.NewProposalRequest("Past the deadline", "", pastDeadline))
	for _, m := range voters {
		testing.SetOriginCaller(m)
		dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	}
	testing.SkipHeights(executionDeadlineBlocks + 1)
	urequire.Equal(t, false, dao.ExecuteProposal(cross(cur), pid))
	urequire.Equal(t, 1, executed)
	page := dao.Render(cross(cur), strconv.FormatInt(int64(pid), 10))
	urequire.Equal(t, true, contains(page, "**PROPOSAL HAS BEEN DENIED**"))
	urequire.Equal(t, true, contains(page, "expired: not executed within 170000 blocks of passing"))

	urequire.AbortsWithMessage(t, cur, "proposal already executed. Accepted: false", func() {
		dao.ExecuteProposal(cross(cur), pid)
	})
}

func TestProposalDeadlineRunsFromPassingVote(cur realm, t *testing.T) {
	loadMembers(cur)

	executed := 0
	slow := dao.NewSimpleExecutor(0, cur, func(realm) error {
		executed++
		return nil
	}, "")

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))
	pid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest("Slow vote", "", slow))
	dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))

	// The deciding votes arrive more than a deadline after creation.
	testing.SkipHeights(executionDeadlineBlocks + 1)
	for _, m := range []address{m11, m111, m1111, m2, m3, m4, m5, m6} {
		testing.SetOriginCaller(m)
		dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	}
	urequire.Equal(t, true, dao.ExecuteProposal(cross(cur), pid))
	urequire.Equal(t, 1, executed)
}

func TestExpiredProposalIsClosedToVotes(cur realm, t *testing.T) {
	loadMembers(cur)

	idle := dao.NewSimpleExecutor(0, cur, func(realm) error { return nil }, "")

	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))
	pid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest("Never executed", "", idle))
	for _, m := range []address{m1, m11, m111, m1111, m2, m3} {
		testing.SetOriginCaller(m)
		dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	}
	testing.SkipHeights(executionDeadlineBlocks + 1)

	// Expired before any execution attempt: votes are refused, and the page
	// says so instead of showing the proposal as open.
	testing.SetOriginCaller(m6)
	urequire.AbortsWithMessage(t, cur, "proposal expired", func() {
		dao.MustVoteOnProposal(cross(cur), dao.NewVoteRequest(dao.YesVote, pid))
	})
	page := dao.Render(cross(cur), strconv.FormatInt(int64(pid), 10))
	urequire.Equal(t, true, contains(page, "**PROPOSAL HAS EXPIRED**: not executed within 170000 blocks of passing"))
	urequire.Equal(t, false, contains(page, "Proposal is open for votes"))
}
