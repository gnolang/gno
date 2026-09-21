package impl

import (
	"testing"

	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gov/dao"
)

// Named z_* so it sorts after govdao_test.gno, whose TestUpgradeDaoImplementation
// swaps the DAO implementation and asserts a hard-coded proposal id. This test
// creates no proposal, so it is safe to run against the shared state.

// A proposal id this implementation has no status for — an unknown id, or a
// proposal that outlived the GovDAO instance that created it (statuses live on
// the instance). The nil deref this used to hit surfaced as "runtime error:
// nil pointer dereference", which reads like a VM fault rather than a bad
// request.
//
// Note what this does NOT assert: that the proposal becomes rejectable.
// dao.executeProposal panics on ANY error from PreExecuteProposal, ahead of
// its execErrorRejects branch, so the guard buys a named error and nothing
// more.
func TestExecuteUnknownProposalNamesTheError(cur realm, t *testing.T) {
	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))

	urequire.AbortsWithMessage(t, cur, "proposal not found", func() {
		dao.ExecuteProposal(cross(cur), dao.ProposalID(999999))
	})
}
