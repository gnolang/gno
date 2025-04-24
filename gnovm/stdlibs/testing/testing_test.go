package testing

import (
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

func TestExpectEmit(t *testing.T) {
	tests := []struct {
		name         string
		emitType     string
		emitAttrs    []string
		expectType   string
		expectAttrs  []string
		eventIndex   int
		partialMatch bool
		shouldMatch  bool
		skipEmit     bool
	}{
		{
			name:         "exact match last event",
			emitType:     "transfer",
			emitAttrs:    []string{"from", "addr1", "to", "addr2", "amount", "100"},
			expectType:   "transfer",
			expectAttrs:  []string{"from", "addr1", "to", "addr2", "amount", "100"},
			eventIndex:   -1,
			partialMatch: false,
			shouldMatch:  true,
		},
		{
			name:         "partial match last event",
			emitType:     "transfer",
			emitAttrs:    []string{"from", "addr1", "to", "addr2", "amount", "100"},
			expectType:   "transfer",
			expectAttrs:  []string{"from", "addr1"},
			eventIndex:   -1,
			partialMatch: true,
			shouldMatch:  true,
		},
		{
			name:         "partial match with extra attributes",
			emitType:     "transfer",
			emitAttrs:    []string{"from", "addr1", "to", "addr2", "amount", "100", "timestamp", "123"},
			expectType:   "transfer",
			expectAttrs:  []string{"from", "addr1", "to", "addr2"},
			eventIndex:   -1,
			partialMatch: true,
			shouldMatch:  true,
		},
		{
			name:        "different event type",
			emitType:    "transfer",
			emitAttrs:   []string{"from", "addr1", "to", "addr2"},
			expectType:  "withdraw",
			expectAttrs: []string{"from", "addr1", "to", "addr2"},
			eventIndex:  -1,
			shouldMatch: false,
		},
		{
			name:        "different attributes",
			emitType:    "transfer",
			emitAttrs:   []string{"from", "addr1", "to", "addr2"},
			expectType:  "transfer",
			expectAttrs: []string{"from", "addr2", "to", "addr1"},
			eventIndex:  -1,
			shouldMatch: false,
		},
		{
			name:        "different attribute count",
			emitType:    "transfer",
			emitAttrs:   []string{"from", "addr1", "to", "addr2", "amount", "100"},
			expectType:  "transfer",
			expectAttrs: []string{"from", "addr1", "to", "addr2"},
			eventIndex:  -1,
			shouldMatch: false,
		},
		{
			name:        "no event",
			expectType:  "transfer",
			expectAttrs: []string{"from", "addr1"},
			eventIndex:  -1,
			shouldMatch: false,
			skipEmit:    true,
		},
		{
			name:        "out of range index",
			emitType:    "transfer",
			emitAttrs:   []string{"from", "addr1"},
			expectType:  "transfer",
			expectAttrs: []string{"from", "addr1"},
			eventIndex:  1,
			shouldMatch: false,
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

			result := X_expectEmit(m, tt.expectType, tt.expectAttrs, tt.eventIndex, tt.partialMatch)
			if result != tt.shouldMatch {
				t.Errorf("X_expectEmit() = %v, want %v", result, tt.shouldMatch)
			}
		})
	}
}

func TestExpectEmit_MultipleEvents(t *testing.T) {
	m := gno.NewMachine("test", nil)
	m.Context = std.ExecContext{
		EventLogger: sdk.NewEventLogger(),
	}

	// partial match cases

	// emit multiple events
	std.X_emit(m, "init", []string{"version", "1.0", "timestamp", "123"})
	std.X_emit(m, "transfer", []string{"from", "addr1", "to", "addr2", "amount", "100", "timestamp", "456"})
	std.X_emit(m, "transfer", []string{"from", "addr2", "to", "addr3", "amount", "50", "timestamp", "789"})

	// 1st event
	if !X_expectEmit(m, "init", []string{"version", "1.0"}, 0, true) {
		t.Error("failed to verify first event with partial match")
	}

	// 2nd event
	if !X_expectEmit(m, "transfer", []string{"from", "addr1", "amount", "100"}, 1, true) {
		t.Error("failed to verify second event with partial match")
	}

	// last event
	if !X_expectEmit(m, "transfer", []string{"to", "addr3"}, -1, true) {
		t.Error("failed to verify last event with partial match")
	}

	// wrong index
	if X_expectEmit(m, "transfer", []string{"from", "addr1"}, 3, false) {
		t.Error("should fail for out of range index")
	}
}
