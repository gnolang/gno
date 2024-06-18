package poc

import (
	"testing"

	"std"

	"gno.land/p/demo/uassert"
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

func TestPoC_AddValidator_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("validator already in set", func(t *testing.T) {
		t.Parallel()

		var (
			proposalAddress = std.Address("caller")
			proposalKey     = "public-key"

			initialSet = generateTestValidators(1)
		)

		initialSet[0].Address = proposalAddress
		initialSet[0].PubKey = proposalKey

		// Create the protocol with an initial set
		p := NewPoC(WithInitialSet(initialSet))

		// Attempt to add the validator
		uassert.PanicsWithError(t, types.ErrValidatorExists, func() {
			p.AddValidator(proposalAddress, proposalKey)
		})
	})
}

func TestPoC_AddValidator(t *testing.T) {
	t.Parallel()

	var (
		proposalAddress = std.Address("caller")
		proposalKey     = "public-key"
	)

	// Create the protocol with no initial set
	p := NewPoC()

	// Attempt to add the validator
	uassert.NotPanics(t, func() {
		p.AddValidator(proposalAddress, proposalKey)
	})

	// Make sure the validator is added
	if !p.IsValidator(proposalAddress) || len(p.validators) != 1 {
		t.Fatal("address is not validator")
	}
}

func TestPoC_RemoveValidator_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("proposed removal not in set", func(t *testing.T) {
		t.Parallel()

		var (
			proposalAddress = std.Address("caller")
			initialSet      = generateTestValidators(1)
		)

		initialSet[0].Address = proposalAddress

		// Create the protocol with an initial set
		p := NewPoC(WithInitialSet(initialSet))

		// Attempt to remove the validator
		uassert.PanicsWithError(t, types.ErrValidatorMissing, func() {
			p.RemoveValidator(std.Address("totally random"))
		})
	})
}

func TestPoC_RemoveValidator(t *testing.T) {
	t.Parallel()

	var (
		proposalAddress = std.Address("caller")
		initialSet      = generateTestValidators(1)
	)

	initialSet[0].Address = proposalAddress

	// Create the protocol with an initial set
	p := NewPoC(WithInitialSet(initialSet))

	// Attempt to remove the validator
	uassert.NotPanics(t, func() {
		p.RemoveValidator(proposalAddress)
	})

	// Make sure the validator is removed
	if p.IsValidator(proposalAddress) || len(p.validators) != 0 {
		t.Fatal("address is validator")
	}
}
