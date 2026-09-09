package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestConfigRejectsANonPositivePrepareBudget: the prepare budget bounds the
// child until it reports ready, and zero would kill every child at spawn while
// reporting the node as unavailable.
func TestConfigRejectsANonPositivePrepareBudget(t *testing.T) {
	valid := config{
		chainID:       "test",
		key:           "approver",
		gnoRoot:       "/gno",
		verifyBudget:  10 * time.Second,
		prepareBudget: time.Minute,
		gasWanted:     1,
	}
	require.NoError(t, valid.validate())

	for _, budget := range []time.Duration{0, -time.Second} {
		cfg := valid
		cfg.prepareBudget = budget
		err := cfg.validate()
		require.Error(t, err)
		require.ErrorContains(t, err, "prepare-budget must be positive")
	}
}
