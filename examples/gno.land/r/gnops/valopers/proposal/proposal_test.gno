package proposal

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/ufmt"
	"gno.land/p/nt/urequire"
	"gno.land/r/gnops/valopers"
	"gno.land/r/gov/dao"
	daoinit "gno.land/r/gov/dao/v3/init" // so that the govdao initializer is executed
)

var g1user = testutils.TestAddress("g1user")

func init() {
	daoinit.InitWithUsers(g1user)
}

func TestValopers_ProposeNewValidator(t *testing.T) {
	const (
		registerMinFee int64 = 20 * 1_000_000 // minimum gnot must be paid to register.
		proposalMinFee int64 = 100 * 1_000_000

		moniker     string = "moniker"
		description string = "description"
		serverType  string = valopers.ServerTypeOnPrem
		pubKey             = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
	)

	// Set origin caller
	testing.SetRealm(testing.NewUserRealm(g1user))

	t.Run("remove an unexisting validator", func(t *testing.T) {
		// Send coins to be able to register a valoper
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", registerMinFee)})

		urequire.NotPanics(t, func() {
			valopers.Register(cross, moniker, description, serverType, g1user, pubKey)
			valopers.UpdateKeepRunning(cross, g1user, false)
		})

		urequire.NotPanics(t, func() {
			valopers.GetByAddr(g1user)
		})

		// Send coins to be able to make a proposal
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", proposalMinFee)})

		urequire.AbortsWithMessage(t, ErrValidatorMissing.Error(), func(cur realm) {
			pr := NewValidatorProposalRequest(cur, g1user)

			dao.MustCreateProposal(cross, pr)
		})
	})

	t.Run("proposal successfully created", func(t *testing.T) {
		// Send coins to be able to register a valoper
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", registerMinFee)})

		urequire.NotPanics(t, func() {
			valopers.UpdateKeepRunning(cross, g1user, true)
		})

		var valoper valopers.Valoper

		urequire.NotPanics(t, func() {
			valoper = valopers.GetByAddr(g1user)
		})

		// Send coins to be able to make a proposal
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", proposalMinFee)})

		var pid dao.ProposalID
		urequire.NotPanics(t, func(cur realm) {
			testing.SetRealm(testing.NewUserRealm(g1user))
			pr := NewValidatorProposalRequest(cur, g1user)

			pid = dao.MustCreateProposal(cross, pr)
		})

		proposal, err := dao.GetProposal(cross, pid) // index starts from 0
		urequire.NoError(t, err, "proposal not found")

		description := ufmt.Sprintf(
			"Valoper profile: [%s](/r/gnops/valopers:%s)\n\n%s\n\n## Validator Updates\n- %s: add\n",
			valoper.Moniker,
			valoper.Address,
			valoper.Render(),
			valoper.Address,
		)

		// Check that the proposal is correct
		urequire.Equal(t, description, proposal.Description())
	})

	t.Run("try to update a validator with the same values", func(t *testing.T) {
		// Send coins to be able to register a valoper
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", registerMinFee)})

		urequire.NotPanics(t, func() {
			valopers.GetByAddr(g1user)
		})

		urequire.NotPanics(t, func() {
			// Vote the proposal created in the previous test
			dao.MustVoteOnProposal(cross, dao.VoteRequest{
				Option:     dao.YesVote,
				ProposalID: dao.ProposalID(0),
			})

			// Execute the proposal
			dao.ExecuteProposal(cross, dao.ProposalID(0))
		})

		// Send coins to be able to make a proposal
		testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", proposalMinFee)})

		urequire.AbortsWithMessage(t, ErrSameValues.Error(), func() {
			pr := NewValidatorProposalRequest(cross, g1user)
			dao.MustCreateProposal(cross, pr)
		})
	})
}

func TestValopers_ProposeNewInstructions(t *testing.T) {
	const proposalMinFee int64 = 100 * 1_000_000

	newInstructions := "new instructions"
	description := ufmt.Sprintf("Update the instructions to: \n\n%s", newInstructions)

	// Set origin caller
	testing.SetRealm(testing.NewUserRealm(g1user))

	// Send coins to be able to make a proposal
	testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", proposalMinFee)})

	var pid dao.ProposalID
	urequire.NotPanics(t, func() {
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

func TestValopers_ProposeNewMinFee(t *testing.T) {
	const proposalMinFee int64 = 100 * 1_000_000
	newMinFee := int64(10)
	description := ufmt.Sprintf("Update the minimum register fee to: %d ugnot", newMinFee)

	// Set origin caller
	testing.SetRealm(testing.NewUserRealm(g1user))

	// Send coins to be able to make a proposal
	testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", proposalMinFee)})

	var pid dao.ProposalID
	urequire.NotPanics(t, func() {
		pr := ProposeNewMinFeeProposalRequest(cross, newMinFee)

		pid = dao.MustCreateProposal(cross, pr)
	})

	proposal, err := dao.GetProposal(cross, pid) // index starts from 0
	urequire.NoError(t, err, "proposal not found")
	// Check that the proposal is correct
	urequire.Equal(t, description, proposal.Description())
}

/* TODO fix this @moul
func TestValopers_ProposeNewValidator2(t *testing.T) {
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
		urequire.NotPanics(t, func() {
			_ = valopers.MustGetValoper(g1user)
		})

		// Create the proposal
		urequire.NotPanics(t, func() {
			cross(valopers.Register)(moniker, description, g1user, pubKey)
		})

		// Verify proposal details
		urequire.NotPanics(t, func() {
			valoper := valopers.MustGetValoper(g1user)
			urequire.Equal(t, moniker, valoper.Moniker)
			urequire.Equal(t, description, valoper.Description)
		})
		// Execute proposal with admin rights
		urequire.NotPanics(t, func() {
			std.TestSetOrigCaller(std.Admin)
			cross(dao.ExecuteProposal)(dao.ProposalID(0))
		})
		// Check if valoper was updated
		urequire.NotPanics(t, func() {
			valoper := valopers.MustGetValoper(g1user)
			urequire.Equal(t, moniker, valoper.Moniker)
			urequire.Equal(t, description, valoper.Description)
		})

		// Expect ExecuteProposal to pass
		urequire.NotPanics(t, func() {
			cross(dao.ExecuteProposal)(dao.ProposalID(0))
		})
		// Check if valoper was updated
		urequire.NotPanics(t, func() {
			valoper := valopers.MustGetValoper(g1user)
			urequire.Equal(t, moniker, valoper.Moniker)
			urequire.Equal(t, description, valoper.Description)
		})
		// Execute proposal with admin rights
		urequire.NotPanics(t, func() {
			std.TestSetOrigCaller(std.Admin)
			cross(dao.ExecuteProposal)(dao.ProposalID(0))
		})
	})
}
*/
