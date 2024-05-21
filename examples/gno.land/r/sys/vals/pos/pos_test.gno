package pos

import (
	"std"
	"testing"

	"gno.land/p/demo/ufmt"
	"gno.land/r/sys/vals/types"
)

func generateTestValidators(count int) []*types.Validator {
	vals := make([]*types.Validator, 0, count)

	for i := 0; i < count; i++ {
		val := &types.Validator{
			Address:     std.Address(ufmt.Sprintf("%d", i)),
			PubKey:      "public-key",
			VotingPower: 0,
		}

		vals = append(vals, val)
	}

	return vals
}

func TestPoS_AddValidator_Invalid(t *testing.T) {
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
		p := NewPoS(WithInitialSet(initialSet))

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)

		// Attempt to add the validator
		testing.PanicsWithError(t, errValidatorExists, func() {
			p.AddValidator(callerAddress, callerKey)
		})
	})

	t.Run("validator set full", func(t *testing.T) {
		t.Parallel()

		var (
			callerAddress = std.Address("caller")
			callerKey     = "public-key"

			initialSet = generateTestValidators(maxValidatorCount)
		)

		// Create the protocol with an initial set
		p := NewPoS(WithInitialSet(initialSet))

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)

		// Attempt to add the validator
		testing.PanicsWithError(t, errMaximumValidatorCount, func() {
			p.AddValidator(callerAddress, callerKey)
		})
	})

	t.Run("validator address is not the origin", func(t *testing.T) {
		t.Parallel()

		var (
			callerAddress = std.Address("caller")
			callerKey     = "public-key"

			initialSet = generateTestValidators(10)
		)

		// Create the protocol with an initial set
		p := NewPoS(WithInitialSet(initialSet))

		// Set the origin caller
		std.TestSetOrigCaller(std.Address("random address"))

		// Attempt to add the validator
		testing.PanicsWithError(t, errValidatorNotOrigin, func() {
			p.AddValidator(callerAddress, callerKey)
		})
	})
}

func TestPoS_AddValidator(t *testing.T) {
	t.Parallel()

	t.Run("user becomes validator in a few tries", func(t *testing.T) {
		t.Parallel()

		var (
			callerAddress = std.Address("caller")
			callerKey     = "public-key"

			numTries = int64(10)

			value = std.Coins{{"ugnot", stakeThreshold.Amount / numTries}}
		)

		// Create the protocol with an initial set
		p := NewPoS()

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)
		std.TestSetOrigSend(value, std.Coins{})

		for i := int64(0); i < numTries; i++ {
			if p.IsValidator(callerAddress) {
				t.Fatalf("should not have a validator at try %d", i)
			}

			// Attempt to add the validator
			testing.NotPanics(t, func() {
				p.AddValidator(callerAddress, callerKey)
			})
		}

		// Make sure the user became a validator
		if !p.IsValidator(callerAddress) {
			t.Fatalf("should be a validator")
		}
	})

	t.Run("user becomes validator in single stake (threshold)", func(t *testing.T) {
		t.Parallel()

		var (
			callerAddress = std.Address("caller")
			callerKey     = "public-key"

			value = std.Coins{{"ugnot", stakeThreshold.Amount}}
		)

		// Create the protocol with an initial set
		p := NewPoS()

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)
		std.TestSetOrigSend(value, std.Coins{})

		// Make sure the caller is not a validator
		if p.IsValidator(callerAddress) {
			t.Fatalf("should not be a validator")
		}

		// Attempt to add the validator
		testing.NotPanics(t, func() {
			p.AddValidator(callerAddress, callerKey)
		})

		// Make sure the user became a validator
		if !p.IsValidator(callerAddress) {
			t.Fatalf("should be a validator")
		}
	})
}
