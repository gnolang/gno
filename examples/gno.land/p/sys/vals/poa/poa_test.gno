package poa

import (
	"std"
	"testing"

	"gno.land/p/demo/ufmt"
	"gno.land/p/sys/vals/types"
)

// generateTestValidators generates a dummy validator set
func generateTestValidators(count int) []*types.Validator {
	vals := make([]*types.Validator, 0, count)

	for i := 0; i < count; i++ {
		val := &types.Validator{
			Address:     std.Address(ufmt.Sprintf("%d", i)),
			PubKey:      "public-key",
			VotingPower: 1,
		}

		vals = append(vals, val)
	}

	return vals
}

func TestPoA_AddValidator_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("validator already in set", func(t *testing.T) {
		t.Parallel()

		var (
			callerAddress = std.Address("caller")
			callerKey     = "public-key"

			initialSet = generateTestValidators(1)
		)

		initialSet[0].Address = callerAddress
		initialSet[0].PubKey = callerKey

		// Create the protocol with an initial set
		p := NewPoA(WithInitialSet(initialSet))

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)

		// Attempt to add the validator
		testing.PanicsWithError(t, types.ErrValidatorExists, func() {
			p.AddValidator(callerAddress, callerKey)
		})
	})

	t.Run("caller not validator", func(t *testing.T) {
		t.Parallel()

		var (
			callerAddress = std.Address("caller")
			callerKey     = "public-key"
		)

		// Create the protocol with no initial set
		p := NewPoA()

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)

		// Attempt to add the validator
		testing.PanicsWithError(t, types.ErrValidatorMissing, func() {
			p.AddValidator(callerAddress, callerKey)
		})
	})
}

func TestPoA_AddValidator(t *testing.T) {
	t.Parallel()

	t.Run("validator voted in by single node", func(t *testing.T) {
		t.Parallel()

		var (
			callerAddress = std.Address("caller")
			callerKey     = "public-key"

			proposedAddress = std.Address("proposal")
			proposedKey     = "proposed-key"

			initialSet = generateTestValidators(1)
		)

		initialSet[0].Address = callerAddress
		initialSet[0].PubKey = callerKey
		initialSet[0].VotingPower = 1

		// Create the protocol with an initial set
		p := NewPoA(WithInitialSet(initialSet))

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)

		// Attempt to add the validator
		testing.NotPanics(t, func() {
			p.AddValidator(proposedAddress, proposedKey)
		})

		// Make sure the validator is added
		if !p.IsValidator(proposedAddress) {
			t.Fatal("address is not validator")
		}

		// Make sure the total voting power is changed
		if p.totalVotingPower != 2 {
			t.Fatal("invalid total voting power")
		}

		// Make sure the majority voting power is changed
		if p.majorityPower != 1 { // 50% is the 2/3 majority for 2 nodes
			t.Fatal("invalid majority voting power")
		}
	})

	t.Run("validator voted in by majority", func(t *testing.T) {
		t.Parallel()

		var (
			callerAddress = std.Address("caller")
			callerKey     = "public-key"

			proposedAddress = std.Address("proposal")
			proposedKey     = "proposed-key"

			initialSet = generateTestValidators(3)
		)

		// Prepare the initial validator set
		for index, validator := range initialSet {
			validator.Address = std.Address(ufmt.Sprintf("caller-%d", index))
			validator.PubKey = ufmt.Sprintf("caller-key-%d", index)
			validator.VotingPower = 1
		}

		// Create the protocol with an initial set
		p := NewPoA(WithInitialSet(initialSet))

		// Add in the votes
		for _, validator := range initialSet {
			// Set the origin caller
			std.TestSetOrigCaller(validator.Address)

			// Attempt to add the validator
			testing.NotPanics(t, func() {
				p.AddValidator(proposedAddress, proposedKey)
			})
		}

		// Make sure the validator is added
		if !p.IsValidator(proposedAddress) {
			t.Fatal("address is not validator")
		}

		// Make sure the total voting power is changed
		if p.totalVotingPower != 4 {
			t.Fatal("invalid total voting power")
		}

		// Make sure the majority voting power is changed
		if p.majorityPower != 2 { // 2 * 4 / 3
			t.Fatal("invalid majority voting power")
		}
	})
}

func TestPoA_RemoveValidator_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("validator not in set", func(t *testing.T) {
		t.Parallel()

		callerAddress := std.Address("caller")

		// Create the protocol with no initial set
		p := NewPoA()

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)

		// Attempt to remove the validator
		testing.PanicsWithError(t, errCallerNotValidator, func() {
			p.RemoveValidator(callerAddress)
		})
	})

	t.Run("proposed removal not in set", func(t *testing.T) {
		t.Parallel()

		var (
			callerAddress = std.Address("caller")
			initialSet    = generateTestValidators(1)
		)

		initialSet[0].Address = callerAddress

		// Create the protocol with an initial set
		p := NewPoA(WithInitialSet(initialSet))

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)

		// Attempt to remove the validator
		testing.PanicsWithError(t, types.ErrValidatorMissing, func() {
			p.RemoveValidator(std.Address("totally random"))
		})
	})
}

func TestPoA_RemoveValidator(t *testing.T) {
	t.Parallel()

	t.Run("validator voted out by single node", func(t *testing.T) {
		t.Parallel()

		var (
			callerAddress = std.Address("caller")
			callerKey     = "public-key"

			initialSet = generateTestValidators(2)
		)

		initialSet[0].Address = callerAddress
		initialSet[0].PubKey = callerKey
		initialSet[0].VotingPower = 10

		proposedAddress := initialSet[1].Address

		// Create the protocol with an initial set
		p := NewPoA(WithInitialSet(initialSet))

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)

		// Attempt to remove the validator
		testing.NotPanics(t, func() {
			p.RemoveValidator(proposedAddress)
		})

		// Make sure the validator is removed
		if p.IsValidator(proposedAddress) {
			t.Fatal("address is still a validator")
		}

		// Make sure the total voting power is changed
		if p.totalVotingPower != 10 {
			t.Fatal("invalid total voting power")
		}

		// Make sure the majority voting power is changed
		if p.majorityPower != 6 {
			t.Fatal("invalid majority voting power")
		}
	})

	t.Run("validator voted in by majority", func(t *testing.T) {
		t.Parallel()

		initialSet := generateTestValidators(4)

		// Prepare the initial validator set
		for index, validator := range initialSet {
			validator.Address = std.Address(ufmt.Sprintf("caller-%d", index))
			validator.PubKey = ufmt.Sprintf("caller-key-%d", index)
			validator.VotingPower = 1
		}

		var (
			proposedAddress = initialSet[len(initialSet)-1].Address // last validator
			voters          = initialSet[:len(initialSet)-1]
		)

		// Create the protocol with an initial set
		p := NewPoA(WithInitialSet(initialSet))

		// Add in the votes
		for _, validator := range voters {
			// Set the origin caller
			std.TestSetOrigCaller(validator.Address)

			// Attempt to remove the validator
			testing.NotPanics(t, func() {
				p.RemoveValidator(proposedAddress)
			})
		}

		// Make sure the validator is removed
		if p.IsValidator(proposedAddress) {
			t.Fatal("address is validator")
		}

		// Make sure the total voting power is changed
		if p.totalVotingPower != 3 {
			t.Fatal("invalid total voting power")
		}

		// Make sure the majority voting power is changed
		if p.majorityPower != 2 { // 2 * 3 / 3
			t.Fatal("invalid majority voting power")
		}
	})
}
