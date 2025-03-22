package valopers_proposal

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/r/gnoland/valopers"
	"gno.land/r/gov/dao/bridge"
	_ "gno.land/r/gov/dao/init" // so that the govdao initializer is executed
)

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

		uassert.NotPanics(t, func() {
			valopers.Register(moniker, description, test1Address, pubKey)
			valopers.UpdateKeepRunning(test1Address, false)
		})

		var valoper valopers.Valoper

		uassert.NotPanics(t, func() {
			valoper = valopers.GetByAddr(test1Address)
		})

		// Send coins to be able to make a proposal
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", proposalMinFee)})

		uassert.PanicsWithMessage(t, ErrValidatorMissing.Error(), func() {
			ProposeNewValidator(test1Address)
		})
	})

	t.Run("proposal successfully created", func(t *testing.T) {
		// Send coins to be able to register a valoper
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", registerMinFee)})

		uassert.NotPanics(t, func() {
			valopers.UpdateKeepRunning(test1Address, true)
		})

		var valoper valopers.Valoper

		uassert.NotPanics(t, func() {
			valoper = valopers.GetByAddr(test1Address)
		})

		dao := bridge.GovDAO()
		propStore := dao.GetPropStore()
		currentSize := propStore.Size()

		// Send coins to be able to make a proposal
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", proposalMinFee)})

		uassert.NotPanics(t, func() {
			ProposeNewValidator(test1Address)
		})

		uassert.Equal(t, currentSize+1, propStore.Size())

		proposal, err := propStore.ProposalByID(uint64(currentSize)) // index starts from 0
		uassert.NoError(t, err, "proposal not found")

		description := ufmt.Sprintf("Valoper profile: [%s](/r/gnoland/valopers:%s)\n\n%s",
			valoper.Moniker,
			valoper.Address,
			valoper.Render(),
		)

		// Check that the proposal is correct
		uassert.Equal(t, description, proposal.Description())
	})

	t.Run("try to update a validator with the same values", func(t *testing.T) {
		// Send coins to be able to register a valoper
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", registerMinFee)})

		var valoper valopers.Valoper

		uassert.NotPanics(t, func() {
			valoper = valopers.GetByAddr(test1Address)
		})

		dao := bridge.GovDAO()
		propStore := dao.GetPropStore()
		currentSize := propStore.Size()

		// Set a GovDAO member as the caller
		testing.SetOriginCaller(std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"))

		uassert.NotPanics(t, func() {
			// Vote the proposal created in the previous test
			dao.VoteOnProposal(uint64(currentSize-1), "YES")

			// Execute the proposal
			dao.ExecuteProposal(uint64(currentSize - 1))
		})

		// Send coins to be able to make a proposal
		testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", proposalMinFee)})

		uassert.PanicsWithMessage(t, ErrSameValues.Error(), func() {
			ProposeNewValidator(test1Address)
		})
	})
}

func TestValopers_ProposeNewInstructions(t *testing.T) {
	const proposalMinFee int64 = 100 * 1_000_000

	newInstructions := "new instructions"
	description := ufmt.Sprintf("Update the instructions to: \n\n%s", newInstructions)

	dao := bridge.GovDAO()
	propStore := dao.GetPropStore()
	currentSize := propStore.Size()

	// Send coins to be able to make a proposal
	testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", proposalMinFee)})

	uassert.NotPanics(t, func() {
		ProposeNewInstructions(newInstructions)
	})

	uassert.Equal(t, currentSize+1, propStore.Size())

	proposal, err := propStore.ProposalByID(uint64(currentSize)) // index starts from 0
	uassert.NoError(t, err, "proposal not found")

	// Check that the proposal is correct
	uassert.Equal(t, description, proposal.Description())
}

func TestValopers_ProposeNewMinFee(t *testing.T) {
	const proposalMinFee int64 = 100 * 1_000_000
	newMinFee := int64(10)
	description := ufmt.Sprintf("Update the minimum register fee to: %d ugnot", newMinFee)

	dao := bridge.GovDAO()
	propStore := dao.GetPropStore()
	currentSize := propStore.Size()

	// Send coins to be able to make a proposal
	testing.SetOriginSend(std.Coins{std.NewCoin("ugnot", proposalMinFee)})

	uassert.NotPanics(t, func() {
		ProposeNewMinFee(newMinFee)
	})

	uassert.Equal(t, currentSize+1, propStore.Size())

	proposal, err := propStore.ProposalByID(uint64(currentSize)) // index starts from 0
	uassert.NoError(t, err, "proposal not found")

	// Check that the proposal is correct
	uassert.Equal(t, description, proposal.Description())
}
