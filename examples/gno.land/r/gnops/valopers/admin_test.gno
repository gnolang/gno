package valopers

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

// TestAuthDescribesGovDAOGate pins init.gno's configuration, including the
// proposer half — which is the load-bearing half.
//
// Deliberately does NOT reassign `auth`: an earlier version of this test
// built its own ContractAuthority and then claimed to "mirror init.gno",
// which meant changing the contract path in init.gno left it green.
//
// Before ContractAuthority.String() rendered the proposer, this assertion
// was blind to the drift that matters: swapping init.gno for
// NewRestrictedContractAuthority(path, handler, NewAutoAcceptAuthority())
// — i.e. reopening the hole for this realm — left the rendered string
// byte-identical and the whole valopers suite green.
//
// It is also blind no longer to a proposer that merely CLAIMS to be the
// contract-identity default: authz renders any non-canonical Authority
// wrapped as custom_authority[...], so a spoofing impl cannot reproduce
// these bytes. See authz's TestSpoofedProposerCannotImpersonateContractIdentity.
func TestAuthDescribesGovDAOGate(cur realm, t *testing.T) {
	uassert.Equal(t,
		"contract_authority[contract=gno.land/r/gov/dao,proposer=contract-identity]",
		Auth())
}

// TestUpdateInstructionsRejectsNonGovDAO is the negative pin for the
// repointed authority, and it is the one that would have caught the
// regression class this realm fears.
//
// The authority is a ContractAuthority for gno.land/r/gov/dao driven by
// DoByPrevious, so the principal is whoever crossed into valopers. Here
// that is the test's own frame, not GovDAO's executor — so the privileged
// write must be refused and `instructions` must not move.
//
// Note what this test could NOT assert before the repoint: under the
// previous ContractAuthority(ownPath) + DoByCurrent pairing, rlm.Address()
// inside any valopers crossing frame is unconditionally valopers' own
// address, so this exact call SUCCEEDED and the old version of this test
// asserted NotPanics. That is why it never pinned anything: the gate it was
// exercising could not reject.
//
// The positive half — that GovDAO can still perform the write — lives in
// proposal/filetests/z_governed_instructions_filetest.gno, which runs
// propose -> vote -> execute end to end.
func TestUpdateInstructionsRejectsNonGovDAO(cur realm, t *testing.T) {
	before := instructions

	// An ERROR, not a panic: the refusal has to travel back to
	// impl.ExecuteProposal as a value so a passed-but-refused proposal can
	// be marked Denied instead of aborting the execute transaction. See
	// updateInstructions.
	err := updateInstructions(0, cur, "SHOULD-NOT-LAND")

	uassert.ErrorContains(t, err, "unauthorized",
		"a non-GovDAO principal must be refused")
	uassert.Equal(t, before, instructions,
		"instructions must not move when the gate refuses")
}

// The instructions proposal's on-chain title is rendered to GovDAO voters
// and is what off-chain tooling matches on, so it is pinned rather than
// left to drift silently. It changed in this realm's history from
// "/p/gnops/valopers: ..." (wrong — valopers is an /r/ realm) when
// proposal construction moved out of r/gnops/valopers/proposal.
func TestInstructionsProposalTitleIsPinned(cur realm, t *testing.T) {
	uassert.Equal(t, "/r/gnops/valopers: Update instructions", instructionsProposalTitle)

	req := NewInstructionsProposalRequest(cross(cur), "whatever")
	uassert.Equal(t, instructionsProposalTitle, req.Title())
}

// The authority must be rotatable by GovDAO, or a principal that stops
// being presentable freezes `instructions` with no on-chain recovery (and
// redeploying is expensive: r/sys/validators/v0/cache.gno hardcodes
// valopersRealmPath). This pins that the rotation path exists, is sealed
// the same way the instructions write is, and that a non-GovDAO principal
// cannot drive it.
func TestRotateAuthorityRejectsNonGovDAO(cur realm, t *testing.T) {
	before := Auth()

	err := rotateAuthority(0, cur, "gno.land/r/gov/dao/v4")

	uassert.ErrorContains(t, err, "unauthorized",
		"a non-GovDAO principal must not be able to rotate the authority")
	uassert.Equal(t, before, Auth(), "the authority must be unchanged")
}

// The rotation request is built here and returned sealed: dao.ProposalRequest
// exposes only Title/Description/Filter, so a holder cannot reach the
// executor. Anyone may build one; only GovDAO can adopt and run it.
func TestRotationProposalRequestIsSealed(cur realm, t *testing.T) {
	req := NewAuthorityRotationProposalRequest(cross(cur), "gno.land/r/gov/dao/v4")

	uassert.Equal(t, "/r/gnops/valopers: Rotate governance authority", req.Title())
	uassert.NotEqual(t, "", req.Description())

	// Building the request must not itself perform the rotation.
	uassert.Equal(t,
		"contract_authority[contract=gno.land/r/gov/dao,proposer=contract-identity]",
		Auth())
}

// TestUpdateMinFee was removed: register_fee now lives in sysparams
// (node:valoper:register_fee), governed via the generic
// NewSysParamUint64PropRequest factory. See proposal.gno for the
// replacement flow.
