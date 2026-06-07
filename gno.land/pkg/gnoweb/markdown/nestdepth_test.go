package markdown

import (
	"testing"

	"github.com/yuin/goldmark/parser"
)

func TestGet_FreshContext(t *testing.T) {
	pc := parser.NewContext()
	if got := Get(pc); got != 0 {
		t.Errorf("Get on fresh context = %d, want 0", got)
	}
}

func TestPushPop_RoundTrip(t *testing.T) {
	pc := parser.NewContext()
	if !Push(pc) {
		t.Fatal("Push from 0 returned false; expected true")
	}
	if got := Get(pc); got != 1 {
		t.Errorf("after one Push, depth = %d, want 1", got)
	}
	Pop(pc)
	if got := Get(pc); got != 0 {
		t.Errorf("after Pop, depth = %d, want 0", got)
	}
}

func TestPush_AtCap(t *testing.T) {
	pc := parser.NewContext()
	for i := 0; i < MaxGnoNestDepth; i++ {
		if !Push(pc) {
			t.Fatalf("Push %d returned false; expected true within cap", i+1)
		}
	}
	if got := Get(pc); got != MaxGnoNestDepth {
		t.Errorf("after %d Pushes, depth = %d, want %d", MaxGnoNestDepth, got, MaxGnoNestDepth)
	}
	if Push(pc) {
		t.Error("Push at cap returned true; expected false")
	}
	// Depth must not have advanced past cap.
	if got := Get(pc); got != MaxGnoNestDepth {
		t.Errorf("after refused Push, depth = %d, want %d", got, MaxGnoNestDepth)
	}
}

func TestPop_BelowZero_Clamps(t *testing.T) {
	pc := parser.NewContext()
	// Pop on empty must not underflow.
	Pop(pc)
	Pop(pc)
	if got := Get(pc); got != 0 {
		t.Errorf("Pop on empty context produced depth %d, want 0", got)
	}
}

func TestSeed_OverridesDepth(t *testing.T) {
	pc := parser.NewContext()
	Seed(pc, 3)
	if got := Get(pc); got != 3 {
		t.Errorf("after Seed(3), depth = %d, want 3", got)
	}
	if !Push(pc) {
		t.Fatal("Push at depth 3 returned false; expected true (cap is 4)")
	}
	if got := Get(pc); got != 4 {
		t.Errorf("after Seed(3) + Push, depth = %d, want 4", got)
	}
	if Push(pc) {
		t.Error("Push at cap returned true; expected false")
	}
}

func TestSeed_AtCap_NextPushRefuses(t *testing.T) {
	pc := parser.NewContext()
	Seed(pc, MaxGnoNestDepth)
	if Push(pc) {
		t.Error("Push at seeded cap returned true; expected false")
	}
	// Depth must remain at cap, not advance.
	if got := Get(pc); got != MaxGnoNestDepth {
		t.Errorf("after refused Push at seeded cap, depth = %d, want %d", got, MaxGnoNestDepth)
	}
}

func TestGet_CorruptValue_ReturnsZero(t *testing.T) {
	pc := parser.NewContext()
	// Defensive: if someone misuses the key by storing a non-int,
	// Get must NOT panic — it returns 0, the same default as for
	// an absent key.
	pc.Set(gnoNestDepthKey, "not an int")
	if got := Get(pc); got != 0 {
		t.Errorf("Get on corrupt value = %d, want 0 (defensive default)", got)
	}
}

func TestIsolation_DistinctContexts(t *testing.T) {
	a := parser.NewContext()
	b := parser.NewContext()
	if !Push(a) {
		t.Fatal("Push on a failed")
	}
	if !Push(a) {
		t.Fatal("second Push on a failed")
	}
	if got := Get(b); got != 0 {
		t.Errorf("b leaked depth from a: Get(b) = %d, want 0", got)
	}
}
