// PKGPATH: gno.land/r/test/valattack
package valattack

// A separate realm keeps the array its changesFn returns and rewrites it after
// the proposal is stored: init and main commit as separate transactions. The
// DAO must execute the entry voters approved, not the rewritten one.

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/sys/validators/v0"
	"gno.land/r/gov/dao"
	valset "gno.land/r/sys/validators/v2"
)

type mockDAO struct{}

func (d *mockDAO) PreCreateProposal(_ int, rlm realm, r dao.ProposalRequest) (address, error) {
	return testutils.TestAddress("proposer"), nil
}
func (d *mockDAO) PostCreateProposal(_ int, rlm realm, r dao.ProposalRequest, pid dao.ProposalID) {}
func (d *mockDAO) VoteOnProposal(_ int, rlm realm, r dao.VoteRequest) error                       { return nil }
func (d *mockDAO) PreExecuteProposal(_ int, rlm realm, pid dao.ProposalID) (bool, error) {
	return true, nil
}
func (d *mockDAO) IsMember(addr address) bool { return false }
func (d *mockDAO) ExecuteProposal(_ int, rlm realm, pid dao.ProposalID, e dao.Executor) error {
	return e.Execute(cross(rlm))
}
func (d *mockDAO) Render(cur realm, pkgpath string, path string) string { return "" }

var (
	approved = validators.Validator{Address: testutils.TestAddress("approved-val"), PubKey: "pk-a", VotingPower: 1}
	evil     = validators.Validator{Address: testutils.TestAddress("evil-val"), PubKey: "pk-e", VotingPower: 1}
	held     []validators.Validator // the attacker realm's own state
	pid      dao.ProposalID
)

func init(cur realm) {
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(&mockDAO{}, nil))
	held = []validators.Validator{approved}
	req := valset.NewPropRequest(cross(cur), func() []validators.Validator { return held }, "Add validator", "")
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("proposer")))
	pid = dao.MustCreateProposal(cross(cur), req)
}

func main(cur realm) {
	held[0] = evil // a write to this realm's own state, after the proposal is stored
	println("executed:", dao.ExecuteProposal(cross(cur), pid))
	println("approved:", valset.IsValidator(approved.Address))
	println("evil:", valset.IsValidator(evil.Address))
}

// Output:
// executed: true
// approved: true
// evil: false
