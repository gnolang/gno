package genesis_test

import (
	"testing"

	"gno.land/p/sys/genesis"
)

func TestGenesisConstants(t *testing.T) {
	// Test that Time is not zero
	if genesis.Time.IsZero() {
		t.Error("Time should not be zero")
	}
}
