package validators

import (
	"strings"
	"testing"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/poa/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
	"gno.land/p/sys/validators/v0"
	"gno.land/r/gov/dao"
)

// snapshotDAO executes every proposal, standing in for impl/v0.
type snapshotDAO struct{}

func (d *snapshotDAO) PreCreateProposal(_ int, rlm realm, r dao.ProposalRequest) (address, error) {
	return testutils.TestAddress("proposer"), nil
}

func (d *snapshotDAO) PostCreateProposal(_ int, rlm realm, r dao.ProposalRequest, pid dao.ProposalID) {
}

func (d *snapshotDAO) VoteOnProposal(_ int, rlm realm, r dao.VoteRequest) error {
	return nil
}

func (d *snapshotDAO) PreExecuteProposal(_ int, rlm realm, pid dao.ProposalID) (bool, error) {
	return true, nil
}

func (d *snapshotDAO) IsMember(addr address) bool { return false }

func (d *snapshotDAO) ExecuteProposal(_ int, rlm realm, pid dao.ProposalID, e dao.Executor) error {
	return e.Execute(cross(rlm))
}

func (d *snapshotDAO) Render(cur realm, pkgpath string, path string) string {
	return ""
}

// TestNewPropRequest_ExecutesApprovedSnapshot covers the approved entry being
// executed after the proposer rewrites its array.
func TestNewPropRequest_ExecutesApprovedSnapshot(cur realm, t *testing.T) {
	changes = bptree.NewBPTree32()
	vp = poa.NewPoA()

	// allowedDAOs is empty, so the proxy accepts this implementation.
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(&snapshotDAO{}, nil))

	approved := validators.Validator{
		Address:     testutils.TestAddress("approved-val"),
		PubKey:      "pk-approved",
		VotingPower: 1,
	}
	evil := validators.Validator{
		Address:     testutils.TestAddress("evil-val"),
		PubKey:      "pk-evil",
		VotingPower: 1,
	}

	proposer := []validators.Validator{approved}
	changesFn := func() []validators.Validator {
		return proposer
	}

	request := NewPropRequest(cur, changesFn, "Add validator", "routine valset upkeep")

	desc := request.Description()
	uassert.True(t, strings.Contains(desc, "- "+approved.Address.String()+": add"))

	pid := dao.MustCreateProposal(cross(cur), request)

	proposer[0] = evil // rewritten during the vote

	urequire.Equal(t, true, dao.ExecuteProposal(cross(cur), pid))

	uassert.True(t, IsValidator(approved.Address))
	uassert.False(t, IsValidator(evil.Address))
}
