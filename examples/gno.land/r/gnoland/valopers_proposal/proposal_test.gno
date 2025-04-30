package valopers_proposal

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
	"gno.land/r/gnoland/valopers"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/init" // so that the govdao initializer is executed
)

func init() {
	c := std.OriginCaller()
	init.InitWithUsers(c)
}

func TestValopers_ProposeNewValidator(t *testing.T) {
	const (
		registerMinFee int64 = 20 * 1_000_000 // minimum gnot must be paid to register.
		proposalMinFee int64 = 100 * 1_000_000

		moniker     string = "moniker"
		description string = "description"
		pubKey             = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
	)

	test1Address := testutils.TestAddress("test1")

	t.Run("remove an unexisting validator", func(t *testing.T) {
		// Send coins to be able to register a valoper
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", registerMinFee)})

		urequire.NotPanics(t, func() {
			valopers.Register(moniker, description, test1Address, pubKey)
			valopers.UpdateKeepRunning(test1Address, false)
		})

		var valoper valopers.Valoper

		urequire.NotPanics(t, func() {
			valoper = valopers.GetByAddr(test1Address)
		})

		// Send coins to be able to make a proposal
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", proposalMinFee)})

		urequire.PanicsWithMessage(t, ErrValidatorMissing.Error(), func() {
			pr := NewValidatorProposalRequest(test1Address)

			dao.MustCreateProposal(pr)
		})
	})

	t.Run("proposal successfully created", func(t *testing.T) {
		// Send coins to be able to register a valoper
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", registerMinFee)})

		urequire.NotPanics(t, func() {
			valopers.UpdateKeepRunning(test1Address, true)
		})

		var valoper valopers.Valoper

		urequire.NotPanics(t, func() {
			valoper = valopers.GetByAddr(test1Address)
		})

		// Send coins to be able to make a proposal
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", proposalMinFee)})

		var pid dao.ProposalID
		urequire.NotPanics(t, func() {
			pr := NewValidatorProposalRequest(test1Address)

			pid := dao.MustCreateProposal(pr)
		})

		proposal, err := dao.GetProposal(pid) // index starts from 0
		urequire.NoError(t, err, "proposal not found")

		description := ufmt.Sprintf("Valoper profile: [%s](/r/gnoland/valopers:%s)\n\n%s",
			valoper.Moniker,
			valoper.Address,
			valoper.Render(),
		)

		// Check that the proposal is correct
		urequire.Equal(t, description, proposal.Description())
	})

	t.Run("try to update a validator with the same values", func(t *testing.T) {
		// Send coins to be able to register a valoper
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", registerMinFee)})

		var valoper valopers.Valoper

		urequire.NotPanics(t, func() {
			valoper = valopers.GetByAddr(test1Address)
		})

		urequire.NotPanics(t, func() {
			// Vote the proposal created in the previous test
			dao.MustVoteOnProposal(dao.VoteRequest{
				Option:     dao.YesVote,
				ProposalID: dao.ProposalID(0),
			})

			// Execute the proposal
			dao.ExecuteProposal(dao.ProposalID(0))
		})

		// Send coins to be able to make a proposal
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", proposalMinFee)})

		urequire.PanicsWithMessage(t, ErrSameValues.Error(), func() {
			pr := NewValidatorProposalRequest(test1Address)

			dao.MustCreateProposal(pr)
		})
	})
}

func TestValopers_ProposeNewInstructions(t *testing.T) {
	const proposalMinFee int64 = 100 * 1_000_000

	newInstructions := "new instructions"
	description := ufmt.Sprintf("Update the instructions to: \n\n%s", newInstructions)

	// Send coins to be able to make a proposal
	testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", proposalMinFee)})

	var pid dao.ProposalID
	urequire.NotPanics(t, func() {
		pr := ProposeNewInstructionsProposalRequest(newInstructions)

		pid = dao.MustCreateProposal(pr)
	})

	proposal, err := dao.GetProposal(pid) // index starts from 0
	urequire.NoError(t, err, "proposal not found")
	if proposal == nil {
		panic("PROPOSAL NOT FOUND")
	}

	// Check that the proposal is correct
	urequire.Equal(t, description, proposal.Description())
}

func TestValopers_ProposeNewMinFee(t *testing.T) {
	const proposalMinFee int64 = 100 * 1_000_000
	newMinFee := int64(10)
	description := ufmt.Sprintf("Update the minimum register fee to: %d ugnot", newMinFee)

	// Send coins to be able to make a proposal
	testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", proposalMinFee)})

	var pid dao.ProposalID
	urequire.NotPanics(t, func() {
		pr := ProposeNewMinFeeProposalRequest(newMinFee)

		pid = dao.MustCreateProposal(pr)
	})

	proposal, err := dao.GetProposal(pid) // index starts from 0
	urequire.NoError(t, err, "proposal not found")
	// Check that the proposal is correct
	urequire.Equal(t, description, proposal.Description())
}
