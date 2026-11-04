package main

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

func TestResolveAppHash(t *testing.T) {
	explicit := []byte{0xAA, 0xBB, 0xCC}
	fromBlock := []byte{0x11, 0x22, 0x33}
	nextMeta := &types.BlockMeta{Header: types.Header{AppHash: fromBlock}}

	t.Run("explicit flag wins over auto-detect", func(t *testing.T) {
		got := resolveAppHash(explicit, nextMeta)
		if !bytes.Equal(got, explicit) {
			t.Errorf("got %X, want %X", got, explicit)
		}
	})

	t.Run("auto-detect from next block header when no explicit hash", func(t *testing.T) {
		got := resolveAppHash(nil, nextMeta)
		if !bytes.Equal(got, fromBlock) {
			t.Errorf("got %X, want %X", got, fromBlock)
		}
	})

	t.Run("nil when neither explicit nor next block available", func(t *testing.T) {
		got := resolveAppHash(nil, nil)
		if got != nil {
			t.Errorf("got %X, want nil", got)
		}
	})
}

// TestCheckRollbackable pins that gnobr refuses a target above the state when
// the block store is more than one block ahead of it, and rolls back any other
// shape: the store at the state's height or one above, as a node leaves it in
// consensus, or a target at or below the state.
func TestCheckRollbackable(t *testing.T) {
	for _, tc := range []struct {
		store, state, target int64
		ok                   bool
	}{
		{10, 10, 5, true},
		{11, 10, 11, true},
		{12, 10, 11, false},
		{12, 10, 12, false},
		{12, 10, 10, true},
		{12, 10, 5, true},
		{0, 0, 0, true},
	} {
		err := checkRollbackable(tc.store, tc.state, tc.target)
		if (err == nil) != tc.ok {
			t.Errorf("store %d, state %d, target %d: got %v, want ok=%v", tc.store, tc.state, tc.target, err, tc.ok)
		}
	}
}
