package impl

import (
	"testing"

	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gov/dao"
)

// Named z_* so it sorts after govdao_test.gno, whose TestUpgradeDaoImplementation
// swaps the DAO implementation and asserts a hard-coded proposal id. This test
// creates no proposal, so it is safe to run against the shared state.

// An empty realmPkg would be stored verbatim as an allowlist entry, and
// InAllowedDAOs compares by exact string against a caller's PkgPath() — which
// is "" for a user realm. It would also render as a blank name in the grant
// sentence, so the proposal would not say who it was empowering. Rejected at
// proposal-construction time so the mistake surfaces to the author rather than
// at execution. UpdateImpl rejects it too; see r/gov/dao/allowlist_test.gno.
func TestUpgradeRequestRejectsBlankRealmPkg(cur realm, t *testing.T) {
	testing.SetOriginCaller(m1)
	testing.SetRealm(testing.NewUserRealm(m1))

	for _, padded := range []string{" gno.land/r/x", "gno.land/r/x "} {
		urequire.PanicsWithMessage(t, cur,
			"realmPkg must not have leading or trailing spaces",
			func() {
				NewUpgradeDaoImplRequest(cur, &mockDAO{}, padded, "reason")
			})
	}

	for _, blank := range []string{"", "   "} {
		// PanicsWithMessage, not AbortsWithMessage: this is a same-realm call,
		// so the guard raises a plain panic rather than a cross-realm abort.
		urequire.PanicsWithMessage(t, cur,
			"realmPkg must be the realm path being granted govDAO authority",
			func() {
				NewUpgradeDaoImplRequest(cur, &mockDAO{}, blank, "reason")
			})
	}
}

type mockDAO struct{}

func (d *mockDAO) PreCreateProposal(_ int, rlm realm, r dao.ProposalRequest) (address, error) {
	return m1, nil
}

func (d *mockDAO) PostCreateProposal(_ int, rlm realm, r dao.ProposalRequest, pid dao.ProposalID) {}

func (d *mockDAO) VoteOnProposal(_ int, rlm realm, r dao.VoteRequest) error { return nil }

func (d *mockDAO) PreExecuteProposal(_ int, rlm realm, pid dao.ProposalID) (bool, error) {
	return true, nil
}

func (d *mockDAO) ExecuteProposal(_ int, rlm realm, pid dao.ProposalID, e dao.Executor) error {
	return nil
}

func (d *mockDAO) Render(cur realm, pkgpath string, path string) string { return "" }
