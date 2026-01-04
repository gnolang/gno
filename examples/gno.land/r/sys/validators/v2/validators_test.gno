package validators

import (
	"chain/runtime"
	"testing"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
	"gno.land/p/sys/validators"
)

// generateTestValidators generates a dummy validator set
func generateTestValidators(count int) []validators.Validator {
	vals := make([]validators.Validator, 0, count)

	for i := 0; i < count; i++ {
		val := validators.Validator{
			Address:     testutils.TestAddress(ufmt.Sprintf("%d", i)),
			PubKey:      "public-key",
			VotingPower: 10,
		}

		vals = append(vals, val)
	}

	return vals
}

func TestValidators_AddRemove(t *testing.T) {
	// Clear any changes
	changes = avl.NewTree()

	var (
		vals          = generateTestValidators(100)
		initialHeight = int64(123)
	)

	// Add in the validators
	for _, val := range vals {
		addValidator(val)

		// Make sure the validator is added
		uassert.True(t, vp.IsValidator(val.Address))

		testing.SkipHeights(1)
	}

	for i := initialHeight; i < initialHeight+int64(len(vals)); i++ {
		// Make sure the changes are saved
		chs := GetChanges(i)

		// We use the funky index calculation to make sure
		// changes are properly handled for each block span
		uassert.Equal(t, initialHeight+int64(len(vals))-i, int64(len(chs)))

		for index, val := range vals[i-initialHeight:] {
			// Make sure the changes are equal to the additions
			ch := chs[index]

			uassert.Equal(t, val.Address, ch.Address)
			uassert.Equal(t, val.PubKey, ch.PubKey)
			uassert.Equal(t, val.VotingPower, ch.VotingPower)
		}
	}

	// Save the beginning height for the removal
	initialRemoveHeight := runtime.ChainHeight()

	// Clear any changes
	changes = avl.NewTree()

	// Remove the validators
	for _, val := range vals {
		removeValidator(val.Address)

		// Make sure the validator is removed
		uassert.False(t, vp.IsValidator(val.Address))

		testing.SkipHeights(1)
	}

	for i := initialRemoveHeight; i < initialRemoveHeight+int64(len(vals)); i++ {
		// Make sure the changes are saved
		chs := GetChanges(i)

		// We use the funky index calculation to make sure
		// changes are properly handled for each block span
		uassert.Equal(t, initialRemoveHeight+int64(len(vals))-i, int64(len(chs)))

		for index, val := range vals[i-initialRemoveHeight:] {
			// Make sure the changes are equal to the additions
			ch := chs[index]

			uassert.Equal(t, val.Address, ch.Address)
			uassert.Equal(t, val.PubKey, ch.PubKey)
			uassert.Equal(t, uint64(0), ch.VotingPower)
		}
	}
}
