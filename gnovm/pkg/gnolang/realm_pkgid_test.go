package gnolang

import "testing"

// TestPkgIDEq verifies the hand-unrolled PkgID.eq against the generic
// == operator, exercising a difference in each word segment of the
// comparison (bytes 0-7, 8-15, 16-19).
func TestPkgIDEq(t *testing.T) {
	t.Parallel()
	base := PkgIDFromPkgPath("gno.land/r/demo/users")
	same := PkgIDFromPkgPath("gno.land/r/demo/users")
	if !base.eq(same) || base != same {
		t.Fatalf("expected %v == %v", base, same)
	}
	if !base.eq(base) {
		t.Fatalf("expected self-equality for %v", base)
	}
	// Flip each byte in turn: covers every word segment of the
	// hand-unrolled compare and fails loudly if HashSize ever drifts
	// away from the 8+8+4 layout eq assumes.
	for seg := range HashSize {
		other := base
		other.Hashlet[seg] ^= 0xFF
		if base.eq(other) {
			t.Errorf("byte %d differs but eq returned true", seg)
		}
		if (base == other) != base.eq(other) {
			t.Errorf("byte %d: eq disagrees with ==", seg)
		}
	}
	var zero PkgID
	if base.eq(zero) || !zero.eq(zero) {
		t.Errorf("zero-value comparisons wrong")
	}
}

// TestHashletIsZero verifies the hand-unrolled Hashlet.IsZero against
// the generic comparison, with a non-zero byte in each word segment.
func TestHashletIsZero(t *testing.T) {
	t.Parallel()
	var zero Hashlet
	if !zero.IsZero() {
		t.Fatal("zero Hashlet reported non-zero")
	}
	for seg := range HashSize {
		var h Hashlet
		h[seg] = 1
		if h.IsZero() {
			t.Errorf("byte %d set but IsZero returned true", seg)
		}
		if h.IsZero() != (h == Hashlet{}) {
			t.Errorf("byte %d: IsZero disagrees with ==", seg)
		}
	}
}
