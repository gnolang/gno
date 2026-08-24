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

	// Test that Height is not zero
	if genesis.Height == 0 {
		t.Error("Height should not be zero")
	}

	// Test that Domain is not empty
	if genesis.Domain == "" {
		t.Error("Domain should not be empty")
	}

	// Test that Uptime is not zero
	if genesis.Uptime() != 0 {
		t.Error("Uptime should be zero")
	}

	// Test that Upheight is not zero
	if genesis.Upheight() != 0 {
		t.Error("Upheight should be zero")
	}
}
