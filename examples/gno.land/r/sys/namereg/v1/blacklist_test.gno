package namereg

import "testing"

// TestReservedNamesSorted enforces lexicographic ordering on reservedNames.
// The list is intended to be auditable at a glance and to grow over time;
// keeping it sorted makes diffs minimal and makes manual scanning for a
// given entry feasible. If this fails after a maintainer adds an entry,
// re-sort the slice rather than weakening the test.
func TestReservedNamesSorted(t *testing.T) {
	for i := 1; i < len(reservedNames); i++ {
		if reservedNames[i-1] >= reservedNames[i] {
			t.Errorf(
				"reservedNames not sorted at index %d: %q should come after %q",
				i, reservedNames[i-1], reservedNames[i],
			)
		}
	}
}
