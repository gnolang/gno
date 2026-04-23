package bptree

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestInlineValueThresholdOption_ClampsAboveMax pins the option-level
// clamp: passing an out-of-range threshold yields an effective
// InlineValueThreshold no larger than MaxInlineValueThreshold so a
// future Set cannot produce a leaf the reader rejects.
func TestInlineValueThresholdOption_ClampsAboveMax(t *testing.T) {
	cases := []struct {
		name string
		in   InlineThreshold
		want InlineThreshold
	}{
		{"at_max", MaxInlineValueThreshold, MaxInlineValueThreshold},
		{"one_above_max", MaxInlineValueThreshold + 1, MaxInlineValueThreshold},
		{"giant", 1 << 30, MaxInlineValueThreshold},
		{"default", DefaultInlineValueThreshold, DefaultInlineValueThreshold},
		{"one", 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := DefaultOptions()
			InlineValueThresholdOption(tc.in)(&opts)
			if opts.InlineValueThreshold != tc.want {
				t.Fatalf("threshold = %d, want %d", opts.InlineValueThreshold, tc.want)
			}
		})
	}
}

// TestResolveInlineThreshold_ClampsStructLiteralBypass pins the
// second-layer clamp: a caller writing Options.InlineValueThreshold
// directly (bypassing the option helper) still gets clamped at
// resolution time.
func TestResolveInlineThreshold_ClampsStructLiteralBypass(t *testing.T) {
	cases := []struct {
		name string
		in   InlineThreshold
		want InlineThreshold
	}{
		{"disabled_zero", 0, InlineDisabled},
		{"disabled_negative", -5, InlineDisabled},
		{"explicit_disabled", InlineDisabled, InlineDisabled},
		{"under_max", DefaultInlineValueThreshold, DefaultInlineValueThreshold},
		{"at_max", MaxInlineValueThreshold, MaxInlineValueThreshold},
		{"one_above_max", MaxInlineValueThreshold + 1, MaxInlineValueThreshold},
		{"giant", 1 << 30, MaxInlineValueThreshold},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveInlineThreshold(tc.in)
			if got != tc.want {
				t.Fatalf("resolveInlineThreshold(%d) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

// TestSet_DemotesValuesAboveMaxInlineThreshold pins the third-layer
// defence: even with t.inlineThreshold artificially driven past
// MaxInlineValueThreshold (simulating a hypothetical future bypass of
// the two clamp layers above), a Set with a value larger than
// MaxInlineValueThreshold still demotes to external storage. This
// keeps the on-disk leaf under the reader's per-leaf budget so
// LoadVersion will not fail to mount the tree.
func TestSet_DemotesValuesAboveMaxInlineThreshold(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 0, NewNopLogger(),
		InlineValueThresholdOption(DefaultInlineValueThreshold))

	// Force the tree's internal threshold past the safety cap. This
	// simulates a path that would otherwise bypass the clamp layers —
	// e.g. a future test or downstream consumer manipulating the field.
	tree.inlineThreshold = MaxInlineValueThreshold * 4

	// A value larger than the cap. Allocate explicitly so we can
	// verify it round-trips byte-for-byte after the Set.
	big := bytes.Repeat([]byte{0xab}, int(MaxInlineValueThreshold)+1)
	if _, err := tree.Set([]byte("big"), big); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Inspect the leaf: the slot at index 0 must be external (no
	// inline bit set) so we know the demotion happened.
	leaf, ok := tree.root.(*LeafNode)
	if !ok {
		t.Fatalf("expected single-leaf root, got %T", tree.root)
	}
	if leaf.inlineMask&1 != 0 {
		t.Fatalf("oversize value was inlined despite exceeding MaxInlineValueThreshold; inlineMask = %#x", leaf.inlineMask)
	}
	if leaf.inlineValues[0] != nil {
		t.Fatalf("inlineValues[0] should be nil for external slot, got %d bytes", len(leaf.inlineValues[0]))
	}
	if leaf.valueKeys[0] == nil {
		t.Fatalf("valueKeys[0] should be set for external slot")
	}

	// Round-trip: the value must read back identically.
	got, err := tree.Get([]byte("big"))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(got, big) {
		t.Fatalf("Get returned %d bytes, want %d bytes", len(got), len(big))
	}
}
