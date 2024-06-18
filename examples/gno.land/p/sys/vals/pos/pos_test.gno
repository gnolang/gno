package pos

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

		// Create the protocol with no initial set
		p := NewPoS(WithInitialSet(initialSet))

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)

		// Attempt to add the validator
		testing.PanicsWithError(t, types.ErrValidatorExists, func() {
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

		// Create the protocol with no initial set
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

func TestPoS_IsValidator(t *testing.T) {
	t.Parallel()

	t.Run("address is not in the set", func(t *testing.T) {
		t.Parallel()

		// Empty PoS set
		p := NewPoS()

		if p.IsValidator(std.Address("random address")) {
			t.Fatal("should not be a validator")
		}
	})

	t.Run("address is in the set", func(t *testing.T) {
		t.Parallel()

		initialSet := generateTestValidators(10)

		// Existing PoS set
		p := NewPoS(WithInitialSet(initialSet))

		if !p.IsValidator(initialSet[0].Address) {
			t.Fatal("should be a validator")
		}
	})
}

func TestPoS_GetValidators(t *testing.T) {
	t.Parallel()

	t.Run("empty set", func(t *testing.T) {
		t.Parallel()

		// Empty PoS set
		p := NewPoS()

		if len(p.GetValidators()) != 0 {
			t.Fatal("should be empty set")
		}
	})

	t.Run("existing set", func(t *testing.T) {
		t.Parallel()

		initialSet := generateTestValidators(10)

		// Existing PoS set
		p := NewPoS(WithInitialSet(initialSet))

		vals := p.GetValidators()

		// Make sure the set has a valid length
		if len(vals) != len(initialSet) {
			t.Fatal("invalid set length")
		}

		// Make sure each validator is valid
		for i, val := range vals {
			ref := initialSet[i]

			if val.Address != ref.Address {
				t.Fatal("address mismatch")
			}

			if val.PubKey != ref.PubKey {
				t.Fatal("address mismatch")
			}

			if val.VotingPower != 0 {
				t.Fatal("invalid voting power")
			}
		}
	})
}

func TestPoS_RemoveValidator_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("validator not in set", func(t *testing.T) {
		t.Parallel()

		callerAddress := std.Address("caller")

		// Create the protocol with no initial set
		p := NewPoS()

		// Set the origin caller
		std.TestSetOrigCaller(callerAddress)

		// Attempt to remove the validator
		testing.PanicsWithError(t, types.ErrValidatorMissing, func() {
			p.RemoveValidator(callerAddress)
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
		std.TestSetOrigCaller(callerAddress)

		// Attempt to remove the validator
		testing.PanicsWithError(t, errValidatorNotOrigin, func() {
			p.RemoveValidator(initialSet[0].Address)
		})
	})
}

func TestPoS_RemoveValidator(t *testing.T) {
	t.Parallel()

	var (
		callerAddress = std.Address("g1d4y807day8r5z54a54rkjheq3l50l5e0xmceem")
		callerKey     = "public-key"

		value = std.Coins{{"ugnot", stakeThreshold.Amount}}
	)

	// Create an initial PoS protocol with no set
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

	// Attempt to remove the validator
	testing.NotPanics(t, func() {
		p.RemoveValidator(callerAddress)
	})

	// Make sure the validator is removed
	if p.IsValidator(callerAddress) {
		t.Fatalf("should not be a validator")
	}
}
