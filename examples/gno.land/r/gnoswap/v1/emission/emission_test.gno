package emission

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestMintAndDistributeGns(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		shouldPanic bool
		panicMsg    string
		verify      func(distributed int64)
	}{
		{
			name: "no block passed",
			verify: func(distributed int64) {
				uassert.Equal(t, int64(0), distributed)
			},
		},
		{
			name: "block passed",
			setup: func() {
				testing.SkipHeights(123)
			},
			verify: func(distributed int64) {
				uassert.True(t, distributed > 0)
			},
		},
		{
			name: "same block",
			verify: func(distributed int64) {
				uassert.Equal(t, int64(0), distributed)
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
					MintAndDistributeGns(cross)
				})
			} else {
				distributed := MintAndDistributeGns(cross)
				if tt.verify != nil {
					tt.verify(distributed)
				}
			}
		})

		resetObject(t)
	}
}
