package validators

import (
	"chain/runtime"
	"math"
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/poa/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/ufmt/v0"
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
		chs := GetChanges(i, initialHeight+int64(len(vals)))

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
		chs := GetChanges(i, initialRemoveHeight+int64(len(vals)))

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

// TestGetChanges_BoundedRange verifies that GetChanges(from, to) correctly
// returns only changes within the [from, to] block range.
func TestGetChanges_BoundedRange(t *testing.T) {
	changes = avl.NewTree()
	vp = poa.NewPoA()

	vals := generateTestValidators(3)

	// Store additions at block h1
	h1 := runtime.ChainHeight()
	for _, val := range vals {
		addValidator(val)
	}
	testing.SkipHeights(1)

	// Store removals at block h2
	h2 := runtime.ChainHeight()
	for _, val := range vals {
		removeValidator(val.Address)
	}
	testing.SkipHeights(1)

	// Query spanning both blocks returns all changes
	all := GetChanges(h1, h2)
	uassert.Equal(t, 6, len(all))

	// Query for h1 only returns additions
	atH1 := GetChanges(h1, h1)
	uassert.Equal(t, 3, len(atH1))
	for i, ch := range atH1 {
		uassert.Equal(t, vals[i].Address, ch.Address)
		uassert.True(t, ch.VotingPower > 0)
	}

	// Query for h2 only returns removals
	atH2 := GetChanges(h2, h2)
	uassert.Equal(t, 3, len(atH2))
	for i, ch := range atH2 {
		uassert.Equal(t, vals[i].Address, ch.Address)
		uassert.Equal(t, uint64(0), ch.VotingPower)
	}

	// Query beyond stored range returns empty
	uassert.Equal(t, 0, len(GetChanges(h2+1, h2+1)))
}

func TestGetChanges_PanicsOnInvalidRange(t *testing.T) {
	uassert.PanicsWithMessage(t, "invalid range: from must be <= to", func() {
		GetChanges(10, 5)
	})
}

func TestGetChanges_ClampsMaxInt64(t *testing.T) {
	changes = avl.NewTree()

	vals := generateTestValidators(1)

	// Simulate a validator change at block math.MaxInt64-1 (the boundary value).
	changes.Set(getBlockID(math.MaxInt64-1), []change{
		{blockNum: math.MaxInt64 - 1, validator: vals[0]},
	})

	// Passing math.MaxInt64 as "to" means "get all updates from here onwards".
	// The clamp (to = MaxInt64-1) must still include the boundary block.
	result := GetChanges(math.MaxInt64-1, math.MaxInt64)
	uassert.Equal(t, 1, len(result))
	uassert.Equal(t, vals[0].Address, result[0].Address)
}
