package gnoland

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
)

// valsetUpdates returns one fresh validator update per given power.
func valsetUpdates(t *testing.T, powers ...int64) []abci.ValidatorUpdate {
	t.Helper()

	updates := generateValidatorUpdates(t, len(powers))
	for i := range updates {
		updates[i].Power = powers[i]
	}
	return updates
}

// TestEndBlockerVotingPowerBound covers rejecting every delta tm2's ApplyBlock would refuse.
func TestEndBlockerVotingPowerBound(t *testing.T) {
	t.Parallel()

	const maxPower = bft.MaxTotalVotingPower

	t.Run("single entry above the cap rejected", func(t *testing.T) {
		t.Parallel()

		current := valsetUpdates(t, 1)
		extra := valsetUpdates(t, maxPower+1)[0]
		proposed := append(append([]abci.ValidatorUpdate{}, current...), extra)

		st := &valsetState{
			current:  serializeUpdates(current),
			proposed: serializeUpdates(proposed),
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		assert.Equal(t, abci.ResponseEndBlock{}, res)
		assert.False(t, st.dirty, "a rejected proposal must clear the flag so a smaller one can land")
		assert.Empty(t, st.currentWrites, "current must not advance on a rejected proposal")
	})

	t.Run("two entries whose sum crosses the cap rejected", func(t *testing.T) {
		t.Parallel()

		current := valsetUpdates(t, 1)
		proposed := append(append([]abci.ValidatorUpdate{}, current...), valsetUpdates(t, maxPower/2+1, maxPower/2+1)...)

		st := &valsetState{
			current:  serializeUpdates(current),
			proposed: serializeUpdates(proposed),
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		assert.Equal(t, abci.ResponseEndBlock{}, res)
		assert.False(t, st.dirty)
		assert.Empty(t, st.currentWrites)
	})

	t.Run("replacement at the cap over the old set's power rejected", func(t *testing.T) {
		t.Parallel()

		// current {maxPower/2, maxPower/2} -> {new: maxPower}: verifyUpdates
		// starts from the current total and adds maxPower, so it refuses.
		current := valsetUpdates(t, maxPower/2, maxPower/2)
		proposed := valsetUpdates(t, maxPower)

		st := &valsetState{
			current:  serializeUpdates(current),
			proposed: serializeUpdates(proposed),
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		assert.Equal(t, abci.ResponseEndBlock{}, res)
		assert.False(t, st.dirty)
		assert.Empty(t, st.currentWrites)
	})

	t.Run("total exactly at the cap accepted", func(t *testing.T) {
		t.Parallel()

		current := valsetUpdates(t, 1)
		extra := valsetUpdates(t, maxPower-1)[0]
		proposed := append(append([]abci.ValidatorUpdate{}, current...), extra)

		st := &valsetState{
			current:  serializeUpdates(current),
			proposed: serializeUpdates(proposed),
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		require.Len(t, res.ValidatorUpdates, 1)
		assert.Equal(t, maxPower-1, res.ValidatorUpdates[0].Power)
		assert.False(t, st.dirty)
		require.Len(t, st.currentWrites, 1)
	})

	t.Run("removal-only proposal accepted", func(t *testing.T) {
		t.Parallel()

		current := valsetUpdates(t, maxPower/2, maxPower/2)
		a, b := current[0], current[1]
		proposed := []abci.ValidatorUpdate{a}

		st := &valsetState{
			current:  serializeUpdates(current),
			proposed: serializeUpdates(proposed),
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		require.Len(t, res.ValidatorUpdates, 1)
		assert.Equal(t, int64(0), res.ValidatorUpdates[0].Power, "the removed validator is emitted with power 0")
		assert.Equal(t, b.Address.String(), res.ValidatorUpdates[0].Address.String())
		assert.False(t, st.dirty)
		require.Len(t, st.currentWrites, 1)
	})

	// valset:current stores the proposal as written, as an unpatched node
	// does; any other bytes move the app hash.
	t.Run("removal entry stored as proposed", func(t *testing.T) {
		t.Parallel()

		v := valsetUpdates(t, 1, 1)
		a, b := v[0], v[1]
		bOff := b
		bOff.Power = 0
		proposed := serializeUpdates([]abci.ValidatorUpdate{a, bOff})

		st := &valsetState{
			current:  serializeUpdates([]abci.ValidatorUpdate{a, b}),
			proposed: proposed,
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		require.Len(t, res.ValidatorUpdates, 1)
		assert.Equal(t, int64(0), res.ValidatorUpdates[0].Power)
		proposedSet, err := abci.ParseValidatorUpdates(proposed)
		require.NoError(t, err)
		assert.Equal(t, abci.EncodeValidatorUpdates(proposedSet), st.current)
	})

	// A removal entry or a duplicate stored in valset:current must not panic
	// the next proposal's pre-check.
	t.Run("removal entry in current does not panic the next proposal", func(t *testing.T) {
		t.Parallel()

		v := valsetUpdates(t, 1, 1, 1)
		a, b, c := v[0], v[1], v[2]
		bOff := b
		bOff.Power = 0

		st := &valsetState{
			current:  serializeUpdates([]abci.ValidatorUpdate{a, b}),
			proposed: serializeUpdates([]abci.ValidatorUpdate{a, bOff}),
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		require.Len(t, res.ValidatorUpdates, 1)
		assert.Equal(t, int64(0), res.ValidatorUpdates[0].Power)

		// r/sys/validators/v0 builds its next proposal from the stored set.
		st.proposed, st.dirty = serializeUpdates([]abci.ValidatorUpdate{a, bOff, c}), true
		require.NotPanics(t, func() {
			res = runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")
		})
		require.Len(t, res.ValidatorUpdates, 1)
		assert.Equal(t, c.Address.String(), res.ValidatorUpdates[0].Address.String())
		assert.Equal(t, abci.EncodeValidatorUpdates(abci.ValidatorUpdates{a, bOff, c}), st.current)
	})

	t.Run("proposal omitting a stored power-0 address rejected", func(t *testing.T) {
		t.Parallel()

		// tm2 refuses to remove a validator it no longer holds, so the gate
		// rejects where a node without it halts.
		v := valsetUpdates(t, 1, 1)
		a, b := v[0], v[1]
		bOff := b
		bOff.Power = 0

		st := &valsetState{
			current:  serializeUpdates([]abci.ValidatorUpdate{a, bOff}),
			proposed: serializeUpdates([]abci.ValidatorUpdate{a}),
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		assert.Equal(t, abci.ResponseEndBlock{}, res)
		assert.False(t, st.dirty)
		assert.Empty(t, st.currentWrites)
	})

	t.Run("duplicate address in current does not panic the next proposal", func(t *testing.T) {
		t.Parallel()

		v := valsetUpdates(t, 1, 1, 1)
		a, b, c := v[0], v[1], v[2]
		b2 := b
		b2.Power = 2

		st := &valsetState{
			current:  serializeUpdates([]abci.ValidatorUpdate{a}),
			proposed: serializeUpdates([]abci.ValidatorUpdate{a, b, b2}),
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		require.Len(t, res.ValidatorUpdates, 1)
		assert.Equal(t, int64(2), res.ValidatorUpdates[0].Power)
		assert.Equal(t, abci.EncodeValidatorUpdates(abci.ValidatorUpdates{a, b, b2}), st.current)

		st.proposed, st.dirty = serializeUpdates([]abci.ValidatorUpdate{a, b2, c}), true
		require.NotPanics(t, func() {
			res = runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")
		})
		require.Len(t, res.ValidatorUpdates, 1)
		assert.Equal(t, c.Address.String(), res.ValidatorUpdates[0].Address.String())
		assert.Equal(t, abci.EncodeValidatorUpdates(abci.ValidatorUpdates{a, b2, c}), st.current)
	})

	// Storing sorts b's two entries lowest power first, so valset:current ends
	// on the power consensus did not apply.
	t.Run("duplicate address listed higher power first does not halt the next proposal", func(t *testing.T) {
		t.Parallel()

		v := valsetUpdates(t, 1, 2, 1)
		a, bLow, c := v[0], v[1], v[2]
		bHigh := bLow
		bHigh.Power = maxPower + 1

		st := &valsetState{
			current:  serializeUpdates([]abci.ValidatorUpdate{a}),
			proposed: serializeUpdates([]abci.ValidatorUpdate{a, bHigh, bLow}),
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")

		require.Len(t, res.ValidatorUpdates, 1)
		assert.Equal(t, int64(2), res.ValidatorUpdates[0].Power, "consensus applies the last entry")

		st.proposed, st.dirty = serializeUpdates([]abci.ValidatorUpdate{a, bLow, c}), true
		require.NotPanics(t, func() {
			res = runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")
		})
		currentSet, err := abci.ParseValidatorUpdates(serializeUpdates([]abci.ValidatorUpdate{a, bLow, bHigh}))
		require.NoError(t, err)
		proposedSet, err := abci.ParseValidatorUpdates(serializeUpdates([]abci.ValidatorUpdate{a, bLow, c}))
		require.NoError(t, err)
		assert.ElementsMatch(t, currentSet.UpdatesFrom(proposedSet), res.ValidatorUpdates,
			"the delta a node without the pre-check returns")
	})

	t.Run("malformed current reported as corruption", func(t *testing.T) {
		t.Parallel()

		// A power above the cap in valset:current is corruption, so the gate
		// reports it rather than letting the set constructor panic.
		over := valsetUpdates(t, maxPower+1)[0]
		current := append(valsetUpdates(t, 1), over)

		st := &valsetState{
			current:  serializeUpdates(current),
			proposed: serializeUpdates(valsetUpdates(t, 2)),
			dirty:    true,
		}
		want := fmt.Sprintf("valset:current corrupted (chain-internal): power %d for %s exceeds %d", over.Power, over.Address, maxPower)
		assert.PanicsWithValue(t, want, func() {
			runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")
		})
	})

	t.Run("current total above the cap reported as corruption", func(t *testing.T) {
		t.Parallel()

		// Each entry is within the cap and the sum is not, so only the total
		// check reports it rather than the set constructor panicking.
		current := valsetUpdates(t, maxPower, 1)

		st := &valsetState{
			current:  serializeUpdates(current),
			proposed: serializeUpdates(valsetUpdates(t, 2)),
			dirty:    true,
		}
		want := fmt.Sprintf("valset:current corrupted (chain-internal): total voting power exceeds %d", maxPower)
		assert.PanicsWithValue(t, want, func() {
			runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")
		})
	})

	t.Run("current total exactly at the cap read by the next proposal", func(t *testing.T) {
		t.Parallel()

		// A set the gate accepted at the cap is stored, and the next proposal
		// reads it back as a valid set rather than as corruption.
		current := valsetUpdates(t, 1)
		extra := valsetUpdates(t, maxPower-1)[0]
		atCap := append(append([]abci.ValidatorUpdate{}, current...), extra)

		st := &valsetState{
			current:  serializeUpdates(current),
			proposed: serializeUpdates(atCap),
			dirty:    true,
		}
		res := runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")
		require.Len(t, res.ValidatorUpdates, 1)

		lower := append([]abci.ValidatorUpdate{}, atCap...)
		lower[1].Power = 2
		st.proposed, st.dirty = serializeUpdates(lower), true
		require.NotPanics(t, func() {
			res = runEndBlocker(t, newValsetMock(st), "/tm.PubKeySecp256k1")
		})
		require.Len(t, res.ValidatorUpdates, 1)
	})
}
