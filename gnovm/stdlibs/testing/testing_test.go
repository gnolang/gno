package testing

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

func TestExpectEmit(t *testing.T) {
	tests := []struct {
		name        string
		emitType    string
		emitAttrs   []string
		expectType  string
		expectAttrs []string
		shouldMatch bool
		skipEmit    bool
	}{
		{
			name:        "exact match",
			emitType:    "transfer",
			emitAttrs:   []string{"from", "addr1", "to", "addr2", "amount", "100"},
			expectType:  "transfer",
			expectAttrs: []string{"from", "addr1", "to", "addr2", "amount", "100"},
			shouldMatch: true,
		},
		{
			name:        "different event type",
			emitType:    "transfer",
			emitAttrs:   []string{"from", "addr1", "to", "addr2"},
			expectType:  "withdraw",
			expectAttrs: []string{"from", "addr1", "to", "addr2"},
			shouldMatch: false,
		},
		{
			name:        "different attributes",
			emitType:    "transfer",
			emitAttrs:   []string{"from", "addr1", "to", "addr2"},
			expectType:  "transfer",
			expectAttrs: []string{"from", "addr2", "to", "addr1"},
			shouldMatch: false,
		},
		{
			name:        "different attribute count",
			emitType:    "transfer",
			emitAttrs:   []string{"from", "addr1", "to", "addr2", "amount", "100"},
			expectType:  "transfer",
			expectAttrs: []string{"from", "addr1", "to", "addr2"},
			shouldMatch: false,
		},
		{
			name:        "no event",
			expectType:  "transfer",
			expectAttrs: []string{"from", "addr1"},
			shouldMatch: false,
			skipEmit:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := gno.NewMachine("test", nil)
			m.Context = std.ExecContext{
				EventLogger: sdk.NewEventLogger(),
			}

			if !tt.skipEmit {
				std.X_emit(m, tt.emitType, tt.emitAttrs)
			}

			result := X_expectEmit(m, tt.expectType, tt.expectAttrs)
			if result != tt.shouldMatch {
				t.Errorf("X_expectEmit() = %v, want %v", result, tt.shouldMatch)
			}
		})
	}
}
