package gnolang

import "testing"

func TestCheckAssignableTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		xt        Type
		dt        Type
		wantError string
	}{
		{
			name: "nil to nil",
			xt:   nil,
			dt:   nil,
		},
		{
			name: "nil and interface",
			xt:   nil,
			dt:   &InterfaceType{},
		},
		{
			name: "interface to nil",
			xt:   &InterfaceType{},
			dt:   nil,
		},
		{
			name:      "nil to non-nillable",
			xt:        nil,
			dt:        StringType,
			wantError: "cannot use nil as string value",
		},
		{
			name: "interface to interface",
			xt:   &InterfaceType{},
			dt:   &InterfaceType{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := checkAssignableTo(nil, tt.xt, tt.dt)
			if tt.wantError != "" {
				if err.Error() != tt.wantError {
					t.Errorf("checkAssignableTo() returned wrong error: want: %v got: %v", tt.wantError, err.Error())
				}
			} else if err != nil {
				t.Errorf("checkAssignableTo() returned unexpected wrong error: got: %v", err.Error())
			}
		})
	}
}

// buildFanout returns a chain of struct types where each level has two
// fields holding the same next-lower level, written as Gno source:
//
//	type L0 struct{ x int }
//	type L1 struct{ a, b [0]L0 }
//	type L2 struct{ a, b [0]L1 }
//	...                            // up to L<depth>, the returned root
//
// Without memoization, checking one level checks the level below twice, so
// the number of checks doubles per level — isComparable on the root would
// take 2^depth steps and never finish at the depths used here. Only the
// per-StructType cache makes the walk finish (one compute per level).
func buildFanout(depth int) *StructType {
	st := &StructType{Fields: []FieldType{{Name: "x", Type: IntType}}}
	for range depth {
		elt := &ArrayType{Len: 0, Elt: st}
		st = &StructType{Fields: []FieldType{
			{Name: "a", Type: elt},
			{Name: "b", Type: elt},
		}}
	}
	return st
}

func TestIsComparableMemoized(t *testing.T) {
	t.Parallel()

	// Depth 200 means 2^200 paths: the test completing at all asserts the
	// walk visits each struct type once, not once per path.
	const depth = 200

	good := buildFanout(depth)
	if !isComparable(good) {
		t.Error("isComparable(comparable fanout) = false, want true")
	}

	// The uncomparable field comes last: isComparable checks fields in
	// order and stops at the first failure, so it must finish walking the
	// entire fan-out subtree before the slice flips the verdict. With the
	// slice first, the walk would fail fast and never enter the fan-out.
	bad := &StructType{Fields: []FieldType{
		{Name: "t", Type: buildFanout(depth)},
		{Name: "s", Type: &SliceType{Elt: IntType}},
	}}
	if isComparable(bad) {
		t.Error("isComparable(uncomparable fanout) = true, want false")
	}

	// Cache reads are directly observable by poisoning a verdict with a lie:
	// the type below is structurally comparable, so a recomputing walk would
	// return true; only a walk that trusts the child's cached verdict can
	// return false.
	poisoned := buildFanout(depth)
	child := poisoned.Fields[0].Type.(*ArrayType).Elt.(*StructType)
	child.comparable = 2
	if isComparable(poisoned) {
		t.Error("walk recomputed a child verdict instead of reading the cache")
	}

	// Same for the root on a repeated call: flip the cached verdict and the
	// next call must return the flipped value, not recompute the truth.
	good.comparable = 2
	if isComparable(good) {
		t.Error("repeated call recomputed the root verdict instead of reading the cache")
	}
}
