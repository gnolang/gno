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
