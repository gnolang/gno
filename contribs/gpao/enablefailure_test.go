package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRecordEnableFailure covers what happens to a package that passed
// verification but failed to enable on-chain.
//
// The old behaviour marked it seen before broadcasting, so the FIRST failure
// retired those bytes for the rest of the run. That matters more once
// submitting costs money: a creator pays the submission charge, the oracle
// approves, the enable fails for a reason that has nothing to do with their
// code -- an unfunded deposit, a dependency not live yet, a gas estimate
// measured against older state -- and the only record is this process's
// stderr. The creator sees a charge and then silence.
func TestRecordEnableFailure(t *testing.T) {
	newOracle := func() *oracle {
		return &oracle{
			seen:         make(map[string]struct{}),
			failedEnable: make(map[string]int),
		}
	}

	t.Run("an early failure leaves the content retryable", func(t *testing.T) {
		o := newOracle()
		n, giveUp := o.recordEnableFailure("pkg@abcd")
		assert.Equal(t, 1, n)
		assert.False(t, giveUp)
		assert.NotContains(t, o.seen, "pkg@abcd",
			"the bytes verified, so the failure is about the chain's state and may clear")
	})

	t.Run("the count runs out and the content is recorded", func(t *testing.T) {
		o := newOracle()
		for i := 1; i < maxEnableAttempts; i++ {
			_, giveUp := o.recordEnableFailure("pkg@abcd")
			require.False(t, giveUp, "attempt %d must not give up yet", i)
			require.NotContains(t, o.seen, "pkg@abcd")
		}
		n, giveUp := o.recordEnableFailure("pkg@abcd")
		assert.Equal(t, maxEnableAttempts, n)
		assert.True(t, giveUp)
		assert.Contains(t, o.seen, "pkg@abcd",
			"a path that keeps failing must stop costing a fee every pass")
	})

	t.Run("the allowance is per content, not per oracle", func(t *testing.T) {
		// Keyed on path plus content hash, so replacing parked bytes earns a
		// fresh allowance -- different bytes are different work, and the retry
		// path for a creator IS to resubmit.
		o := newOracle()
		for range maxEnableAttempts {
			o.recordEnableFailure("pkg@abcd")
		}
		require.Contains(t, o.seen, "pkg@abcd")

		n, giveUp := o.recordEnableFailure("pkg@ef01")
		assert.Equal(t, 1, n)
		assert.False(t, giveUp)
		assert.NotContains(t, o.seen, "pkg@ef01",
			"one exhausted path must not retire another")
	})
}
