package poa

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
	"gno.land/p/sys/validators"

	"gno.land/p/nt/ufmt/v0"
)

// generateTestValidators generates a dummy validator set
func generateTestValidators(count int) []validators.Validator {
	vals := make([]validators.Validator, 0, count)

	for i := 0; i < count; i++ {
		val := validators.Validator{
			Address:     testutils.TestAddress(ufmt.Sprintf("%d", i)),
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
			proposalAddress = testutils.TestAddress("caller")
			proposalKey     = "public-key"

			initialSet = generateTestValidators(1)
		)

		initialSet[0].Address = proposalAddress
		initialSet[0].PubKey = proposalKey

		// Create the protocol with an initial set
		p := NewPoA(WithInitialSet(initialSet))

		// Attempt to add the validator
		_, err := p.AddValidator(proposalAddress, proposalKey, 1)
		uassert.ErrorIs(t, err, validators.ErrValidatorExists)
	})

	t.Run("invalid voting power", func(t *testing.T) {
		t.Parallel()

		var (
			proposalAddress = testutils.TestAddress("caller")
			proposalKey     = "public-key"
		)

		// Create the protocol with no initial set
		p := NewPoA()

		// Attempt to add the validator
		_, err := p.AddValidator(proposalAddress, proposalKey, 0)
		uassert.ErrorIs(t, err, ErrInvalidVotingPower)
	})
}

func TestPoA_AddValidator(t *testing.T) {
	t.Parallel()

	var (
		proposalAddress = testutils.TestAddress("caller")
		proposalKey     = "public-key"
	)

	// Create the protocol with no initial set
	p := NewPoA()

	// Attempt to add the validator
	_, err := p.AddValidator(proposalAddress, proposalKey, 1)
	uassert.NoError(t, err)

	// Make sure the validator is added
	if !p.IsValidator(proposalAddress) || p.validators.Size() != 1 {
		t.Fatal("address is not validator")
	}
}

func TestPoA_RemoveValidator_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("proposed removal not in set", func(t *testing.T) {
		t.Parallel()

		var (
			proposalAddress = testutils.TestAddress("caller")
			initialSet      = generateTestValidators(1)
		)

		initialSet[0].Address = proposalAddress

		// Create the protocol with an initial set
		p := NewPoA(WithInitialSet(initialSet))

		// Attempt to remove the validator
		_, err := p.RemoveValidator(testutils.TestAddress("totally random"))
		uassert.ErrorIs(t, err, validators.ErrValidatorMissing)
	})
}

func TestPoA_RemoveValidator(t *testing.T) {
	t.Parallel()

	var (
		proposalAddress = testutils.TestAddress("caller")
		initialSet      = generateTestValidators(1)
	)

	initialSet[0].Address = proposalAddress

	// Create the protocol with an initial set
	p := NewPoA(WithInitialSet(initialSet))

	// Attempt to remove the validator
	_, err := p.RemoveValidator(proposalAddress)
	urequire.NoError(t, err)

	// Make sure the validator is removed
	if p.IsValidator(proposalAddress) || p.validators.Size() != 0 {
		t.Fatal("address is validator")
	}
}

func TestPoA_GetValidator(t *testing.T) {
	t.Parallel()

	t.Run("validator not in set", func(t *testing.T) {
		t.Parallel()

		// Create the protocol with no initial set
		p := NewPoA()

		// Attempt to get the voting power
		_, err := p.GetValidator(testutils.TestAddress("caller"))
		uassert.ErrorIs(t, err, validators.ErrValidatorMissing)
	})

	t.Run("validator fetched", func(t *testing.T) {
		t.Parallel()

		var (
			address_XXX = testutils.TestAddress("caller")
			pubKey      = "public-key"
			votingPower = uint64(10)

			initialSet = generateTestValidators(1)
		)

		initialSet[0].Address = address_XXX
		initialSet[0].PubKey = pubKey
		initialSet[0].VotingPower = votingPower

		// Create the protocol with an initial set
		p := NewPoA(WithInitialSet(initialSet))

		// Get the validator
		val, err := p.GetValidator(address_XXX)
		urequire.NoError(t, err)

		// Validate the address
		if val.Address != address_XXX {
			t.Fatal("invalid address")
		}

		// Validate the voting power
		if val.VotingPower != votingPower {
			t.Fatal("invalid voting power")
		}

		// Validate the public key
		if val.PubKey != pubKey {
			t.Fatal("invalid public key")
		}
	})
}

func TestPoA_GetValidators(t *testing.T) {
	t.Parallel()

	t.Run("empty set", func(t *testing.T) {
		t.Parallel()

		// Create the protocol with no initial set
		p := NewPoA()

		// Attempt to get the voting power
		vals := p.GetValidators()

		if len(vals) != 0 {
			t.Fatal("validator set is not empty")
		}
	})

	t.Run("validator set fetched", func(t *testing.T) {
		t.Parallel()

		initialSet := generateTestValidators(10)

		// Create the protocol with an initial set
		p := NewPoA(WithInitialSet(initialSet))

		// Get the validator set
		vals := p.GetValidators()

		if len(vals) != len(initialSet) {
			t.Fatal("returned validator set mismatch")
		}

		for _, val := range vals {
			for _, initialVal := range initialSet {
				if val.Address != initialVal.Address {
					continue
				}

				// Validate the voting power
				uassert.Equal(t, val.VotingPower, initialVal.VotingPower)

				// Validate the public key
				uassert.Equal(t, val.PubKey, initialVal.PubKey)
			}
		}
	})
}
