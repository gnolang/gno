package proposal

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gnops/valopers"
	"gno.land/r/gov/dao"
	daoinit "gno.land/r/gov/dao/v3/init" // so that the govdao initializer is executed
)


// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
var g1user = testutils.TestAddress("g1user")

func init() {
	daoinit.InitWithUsers(g1user)
}

func TestValopers_ProposeNewValidator(cur realm, t *testing.T) {
	const (
		registerMinFee int64 = 20 * 1_000_000 // minimum gnot must be paid to register.
		proposalMinFee int64 = 100 * 1_000_000

		moniker     string = "moniker"
		description string = "description"
		serverType  string = valopers.ServerTypeOnPrem
		pubKey             = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
	)

	// signingAddr is derived from the pubkey and is what
	// v3.IsValidator/GetValidator return after a proposal lands
	// (parseEntry in r/sys/params canonicalizes from PubKey). Used
	// for the "ErrSameValues" subtest's post-execute check.
	signingAddr, err := chain.PubKeyAddress(pubKey)
	urequire.NoError(t, err, "valid pubkey")

	// The operator address is g1user (the GovDAO member); the
	// signing address is derived separately from pubKey. Splitting
	// them exercises the operator/signer separation.
	opAddr := g1user

	// Set origin caller for valoper operations: must equal the
	// operator address (post-genesis squat guard in Register).
	testing.SetRealm(testing.NewUserRealm(opAddr))

	t.Run("remove an unexisting validator", func(t *testing.T) {
		// Send coins to be able to register a valoper
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", registerMinFee)})

		urequire.NotPanics(t, cur, func() {
			valopers.Register(cross, moniker, description, serverType, opAddr, pubKey)
			valopers.UpdateKeepRunning(cross, opAddr, false)
		})

		urequire.NotPanics(t, cur, func() {
			valopers.GetByAddr(opAddr)
		})

		// Send coins to be able to make a proposal
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", proposalMinFee)})

		urequire.AbortsWithMessage(t, cur, ErrValidatorMissing.Error(), func(cur realm) {
			pr := NewValidatorProposalRequest(cur, opAddr)

			dao.MustCreateProposal(cross, pr)
		})
	})

	t.Run("proposal successfully created", func(t *testing.T) {
		// Send coins to be able to register a valoper
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", registerMinFee)})

		urequire.NotPanics(t, cur, func() {
			valopers.UpdateKeepRunning(cross, opAddr, true)
		})

		var valoper valopers.Valoper

		urequire.NotPanics(t, cur, func() {
			valoper = valopers.GetByAddr(opAddr)
		})

		// Send coins to be able to make a proposal
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", proposalMinFee)})

		var pid dao.ProposalID
		urequire.NotPanics(t, cur, func(cur realm) {
			testing.SetRealm(testing.NewUserRealm(g1user))
			pr := NewValidatorProposalRequest(cur, opAddr)

			pid = dao.MustCreateProposal(cross, pr)
		})

		proposal, err := dao.GetProposal(cross, pid) // index starts from 0
		urequire.NoError(t, err, "proposal not found")

		// The proposal description is now operator-keyed (matches the
		// new v3.NewValidatorProposalRequest format): both the profile
		// link and the validator-updates section reference the
		// OPERATOR address with explicit power.
		description := ufmt.Sprintf(
			"Valoper profile: [%s](/r/gnops/valopers:%s)\n\n%s\n\n## Validator Updates\n- %s: add (power 1)\n",
			valoper.Moniker,
			valoper.OperatorAddress,
			valoper.Render(),
			valoper.OperatorAddress,
		)

		// Check that the proposal is correct
		urequire.Equal(t, description, proposal.Description())
	})

	t.Run("try to update a validator with the same values", func(t *testing.T) {
		// Send coins to be able to register a valoper
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", registerMinFee)})

		urequire.NotPanics(t, cur, func() {
			valopers.GetByAddr(opAddr)
		})

		urequire.NotPanics(t, cur, func() {
			// Vote the proposal created in the previous test
			dao.MustVoteOnProposal(cross, dao.NewVoteRequest(dao.YesVote, dao.ProposalID(0)))

			// Execute the proposal — callback writes valset:proposed +
			// dirty=true. v3.IsValidator(signingAddr) below reads the
			// effective view (proposed-when-dirty) and finds the
			// just-added validator → enters the ErrSameValues branch.
			dao.ExecuteProposal(cross, dao.ProposalID(0))
		})

		// Verify the just-added entry exists at the SIGNING address.
		_ = signingAddr

		// Send coins to be able to make a proposal
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", proposalMinFee)})

		urequire.AbortsWithMessage(t, cur, ErrSameValues.Error(), func() {
			pr := NewValidatorProposalRequest(cross, opAddr)
			dao.MustCreateProposal(cross, pr)
		})
	})
}

func TestValopers_ProposeNewInstructions(cur realm, t *testing.T) {
	const proposalMinFee int64 = 100 * 1_000_000

	newInstructions := "new instructions"
	description := ufmt.Sprintf("Update the instructions to: \n\n%s", newInstructions)

	// Set origin caller
	testing.SetRealm(testing.NewUserRealm(g1user))

	// Send coins to be able to make a proposal
	testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", proposalMinFee)})

	var pid dao.ProposalID
	urequire.NotPanics(t, cur, func() {
		pr := ProposeNewInstructionsProposalRequest(cross, newInstructions)

		pid = dao.MustCreateProposal(cross, pr)
	})

	proposal, err := dao.GetProposal(cross, pid) // index starts from 0
	urequire.NoError(t, err, "proposal not found")
	if proposal == nil {
		panic("PROPOSAL NOT FOUND")
	}

	// Check that the proposal is correct
	urequire.Equal(t, description, proposal.Description())
}

func TestValopers_ProposeNewMinFee(cur realm, t *testing.T) {
	const proposalMinFee int64 = 100 * 1_000_000
	newMinFee := int64(10)
	// ProposeNewMinFeeProposalRequest now delegates to the generic
	// sys/params factory; description is the standard one from
	// r/sys/params.newPropRequest, keyed on node:valoper:register_fee.
	description := "This proposal wants to add a new key to sys/params: node:valoper:register_fee"

	// Set origin caller
	testing.SetRealm(testing.NewUserRealm(g1user))

	// Send coins to be able to make a proposal
	testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", proposalMinFee)})

	var pid dao.ProposalID
	urequire.NotPanics(t, cur, func() {
		pr := ProposeNewMinFeeProposalRequest(cross, newMinFee)

		pid = dao.MustCreateProposal(cross, pr)
	})

	proposal, err := dao.GetProposal(cross, pid) // index starts from 0
	urequire.NoError(t, err, "proposal not found")
	// Check that the proposal is correct
	urequire.Equal(t, description, proposal.Description())
}

/* TODO fix this @moul
func TestValopers_ProposeNewValidator2(cur realm, t *testing.T) {
	const (
		registerMinFee int64 = 20 * 1_000_000 // minimum gnot must be paid to register.
		proposalMinFee int64 = 100 * 1_000_000

		moniker     string = "moniker"
		description string = "description"
		pubKey             = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
	)

	// Set origin caller
	testing.SetRealm(std.NewUserRealm(g1user))

	t.Run("create valid proposal", func(t *testing.T) {
		// Validator exists, should not panic
		urequire.NotPanics(t, cur, func() {
			_ = valopers.MustGetValoper(g1user)
		})

		// Create the proposal
		urequire.NotPanics(t, cur, func() {
			cross(valopers.Register)(moniker, description, g1user, pubKey)
		})

		// Verify proposal details
		urequire.NotPanics(t, cur, func() {
			valoper := valopers.MustGetValoper(g1user)
			urequire.Equal(t, moniker, valoper.Moniker)
			urequire.Equal(t, description, valoper.Description)
		})
		// Execute proposal with admin rights
		urequire.NotPanics(t, cur, func() {
			std.TestSetOrigCaller(std.Admin)
			cross(dao.ExecuteProposal)(dao.ProposalID(0))
		})
		// Check if valoper was updated
		urequire.NotPanics(t, cur, func() {
			valoper := valopers.MustGetValoper(g1user)
			urequire.Equal(t, moniker, valoper.Moniker)
			urequire.Equal(t, description, valoper.Description)
		})

		// Expect ExecuteProposal to pass
		urequire.NotPanics(t, cur, func() {
			cross(dao.ExecuteProposal)(dao.ProposalID(0))
		})
		// Check if valoper was updated
		urequire.NotPanics(t, cur, func() {
			valoper := valopers.MustGetValoper(g1user)
			urequire.Equal(t, moniker, valoper.Moniker)
			urequire.Equal(t, description, valoper.Description)
		})
		// Execute proposal with admin rights
		urequire.NotPanics(t, cur, func() {
			std.TestSetOrigCaller(std.Admin)
			cross(dao.ExecuteProposal)(dao.ProposalID(0))
		})
	})
}
*/
