package gnolang

import (
	"bytes"
	"testing"
)

func TestCoverageCollectorBasic(t *testing.T) {
	c := NewCoverageCollector()
	if c.Percentage() != 0 {
		t.Fatalf("expected 0%% coverage, got %.1f%%", c.Percentage())
	}
	if c.HasBlocks() {
		t.Fatal("expected no blocks")
	}
}

func TestCoverageCollectorRegistration(t *testing.T) {
	c := NewCoverageCollector()

	// Create some mock statements with spans.
	s1 := &ExprStmt{}
	s1.Span = Span4(1, 1, 1, 20)
	s2 := &ExprStmt{}
	s2.Span = Span4(2, 1, 2, 20)
	s3 := &ExprStmt{}
	s3.Span = Span4(3, 1, 3, 20)

	c.RegisterBlock(s1, "pkg/file.gno", 1, 1, 1, 20, 1)
	c.RegisterBlock(s2, "pkg/file.gno", 2, 1, 2, 20, 1)
	c.RegisterBlock(s3, "pkg/file.gno", 3, 1, 3, 20, 1)

	if !c.HasBlocks() {
		t.Fatal("expected blocks to be registered")
	}
	if c.Percentage() != 0 {
		t.Fatalf("expected 0%% coverage before hits, got %.1f%%", c.Percentage())
	}
}

func TestCoverageCollectorHits(t *testing.T) {
	c := NewCoverageCollector()

	s1 := &ExprStmt{}
	s1.Span = Span4(1, 1, 1, 20)
	s2 := &ExprStmt{}
	s2.Span = Span4(2, 1, 2, 20)
	s3 := &ExprStmt{}
	s3.Span = Span4(3, 1, 3, 20)

	c.RegisterBlock(s1, "pkg/file.gno", 1, 1, 1, 20, 1)
	c.RegisterBlock(s2, "pkg/file.gno", 2, 1, 2, 20, 1)
	c.RegisterBlock(s3, "pkg/file.gno", 3, 1, 3, 20, 1)

	// Hit 2 out of 3 statements.
	c.HitStmt(s1)
	c.HitStmt(s2)

	pct := c.Percentage()
	// 2/3 = 66.67%
	if pct < 66.6 || pct > 66.7 {
		t.Fatalf("expected ~66.7%% coverage, got %.1f%%", pct)
	}

	// Hit all 3.
	c.HitStmt(s3)
	if c.Percentage() != 100 {
		t.Fatalf("expected 100%% coverage, got %.1f%%", c.Percentage())
	}
}

func TestCoverageCollectorHitUnregistered(t *testing.T) {
	c := NewCoverageCollector()
	s1 := &ExprStmt{}
	s1.Span = Span4(1, 1, 1, 20)
	// Hitting an unregistered statement should not panic.
	c.HitStmt(s1)
}

func TestCoverageCollectorDuplicateRegistration(t *testing.T) {
	c := NewCoverageCollector()
	s1 := &ExprStmt{}
	s1.Span = Span4(1, 1, 1, 20)
	c.RegisterBlock(s1, "pkg/file.gno", 1, 1, 1, 20, 1)
	c.RegisterBlock(s1, "pkg/file.gno", 1, 1, 1, 20, 1) // duplicate
	if len(c.blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(c.blocks))
	}
}

func TestCoverageCollectorReset(t *testing.T) {
	c := NewCoverageCollector()
	s1 := &ExprStmt{}
	s1.Span = Span4(1, 1, 1, 20)
	c.RegisterBlock(s1, "pkg/file.gno", 1, 1, 1, 20, 1)
	c.HitStmt(s1)
	if c.Percentage() != 100 {
		t.Fatalf("expected 100%%, got %.1f%%", c.Percentage())
	}
	c.Reset()
	if c.Percentage() != 0 {
		t.Fatalf("expected 0%% after reset, got %.1f%%", c.Percentage())
	}
}

func TestCoverageCollectorString(t *testing.T) {
	c := NewCoverageCollector()
	if c.String() != "coverage: [no statements]" {
		t.Fatalf("unexpected string: %s", c.String())
	}
	s1 := &ExprStmt{}
	s1.Span = Span4(1, 1, 1, 20)
	c.RegisterBlock(s1, "pkg/file.gno", 1, 1, 1, 20, 1)
	c.HitStmt(s1)
	expected := "coverage: 100.0% of statements"
	if c.String() != expected {
		t.Fatalf("expected %q, got %q", expected, c.String())
	}
}

func TestCoverageCollectorWriteProfile(t *testing.T) {
	c := NewCoverageCollector()
	s1 := &ExprStmt{}
	s1.Span = Span4(1, 1, 1, 20)
	s2 := &ExprStmt{}
	s2.Span = Span4(5, 1, 8, 2)

	c.RegisterBlock(s1, "gno.land/p/demo/avl/avl.gno", 1, 1, 1, 20, 1)
	c.RegisterBlock(s2, "gno.land/p/demo/avl/avl.gno", 5, 1, 8, 2, 1)
	c.HitStmt(s1)

	var buf bytes.Buffer
	c.WriteCoverProfile(&buf, "set")
	out := buf.String()

	// Should start with mode line.
	if !bytes.HasPrefix([]byte(out), []byte("mode: set\n")) {
		t.Fatalf("profile should start with mode line, got: %s", out)
	}
	// Should contain block entries.
	if !bytes.Contains([]byte(out), []byte("gno.land/p/demo/avl/avl.gno:1.1,1.20 1 1")) {
		t.Fatalf("profile missing hit block entry, got: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("gno.land/p/demo/avl/avl.gno:5.1,8.2 1 0")) {
		t.Fatalf("profile missing unhit block entry, got: %s", out)
	}
}

func TestCoverageCollectorMerge(t *testing.T) {
	c1 := NewCoverageCollector()
	c2 := NewCoverageCollector()

	s1 := &ExprStmt{}
	s1.Span = Span4(1, 1, 1, 20)
	s2 := &ExprStmt{}
	s2.Span = Span4(2, 1, 2, 20)

	// Same blocks in both.
	c1.RegisterBlock(s1, "pkg/file.gno", 1, 1, 1, 20, 1)
	c1.RegisterBlock(s2, "pkg/file.gno", 2, 1, 2, 20, 1)

	// c2 uses different stmt pointers but same positions.
	s1b := &ExprStmt{}
	s1b.Span = Span4(1, 1, 1, 20)
	s2b := &ExprStmt{}
	s2b.Span = Span4(2, 1, 2, 20)
	c2.RegisterBlock(s1b, "pkg/file.gno", 1, 1, 1, 20, 1)
	c2.RegisterBlock(s2b, "pkg/file.gno", 2, 1, 2, 20, 1)

	c1.HitStmt(s1)
	c2.HitStmt(s2b)

	c1.Merge(c2)

	// Both blocks should now be hit.
	if c1.Percentage() != 100 {
		t.Fatalf("expected 100%% after merge, got %.1f%%", c1.Percentage())
	}
}

func TestCoverageCollectorFilterByPackage(t *testing.T) {
	c := NewCoverageCollector()

	s1 := &ExprStmt{}
	s1.Span = Span4(1, 1, 1, 20)
	s2 := &ExprStmt{}
	s2.Span = Span4(2, 1, 2, 20)

	c.RegisterBlock(s1, "gno.land/p/demo/avl/avl.gno", 1, 1, 1, 20, 1)
	c.RegisterBlock(s2, "gno.land/p/other/pkg/pkg.gno", 2, 1, 2, 20, 1)

	c.HitStmt(s1)

	pct := c.FilterByPackage("gno.land/p/demo/avl")
	if pct != 100 {
		t.Fatalf("expected 100%% for avl package, got %.1f%%", pct)
	}
	pct2 := c.FilterByPackage("gno.land/p/other/pkg")
	if pct2 != 0 {
		t.Fatalf("expected 0%% for other package, got %.1f%%", pct2)
	}
}
