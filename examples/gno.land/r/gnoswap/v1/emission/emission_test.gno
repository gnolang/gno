package emission

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestMintAndDistributeGns(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		shouldPanic bool
		panicMsg    string
		verify      func(distributed uint64)
	}{
		{
			name: "no block passed",
			verify: func(distributed uint64) {
				uassert.Equal(t, uint64(0), distributed)
			},
		},
		{
			name: "block passed",
			setup: func() {
				std.TestSkipHeights(123)
			},
			verify: func(distributed uint64) {
				uassert.True(t, distributed > 0)
			},
		},
		{
			name: "same block",
			verify: func(distributed uint64) {
				uassert.Equal(t, uint64(0), distributed)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					MintAndDistributeGns()
				})
			} else {
				distributed := MintAndDistributeGns()
				if tt.verify != nil {
					tt.verify(distributed)
				}
			}
		})
	}
}
