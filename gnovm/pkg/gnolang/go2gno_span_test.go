package gnolang

import "testing"

// parseOneBinaryExpr returns the first ValueDecl's first Value as a
// *BinaryExpr — used by the span-correctness tests below.
func parseOneBinaryExpr(t *testing.T, src string) *BinaryExpr {
	t.Helper()
	m := NewMachine("test", nil)
	t.Cleanup(m.Release)
	fn, err := m.ParseFile("test.gno", src)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	vd, ok := fn.Decls[0].(*ValueDecl)
	if !ok {
		t.Fatalf("Decls[0] = %T; want *ValueDecl", fn.Decls[0])
	}
	bx, ok := vd.Values[0].(*BinaryExpr)
	if !ok {
		t.Fatalf("Values[0] = %T; want *BinaryExpr", vd.Values[0])
	}
	return bx
}

// TestParseFile_BinaryChain_SpanCorrect verifies the children-span
// optimization preserves Span semantics on the basic chain shape:
// the outer chain's Span.Pos points at the leftmost leaf, Span.End
// at the rightmost leaf's End.
func TestParseFile_BinaryChain_SpanCorrect(t *testing.T) {
	t.Parallel()
	// `const x = 1 + 2 + 3 + 4`
	//  col:    1234567890123456789012345
	//                    ^         ^
	// outer BinaryExpr Span covers col 11–24 (End is exclusive).
	src := "package main\n\nconst x = 1 + 2 + 3 + 4\n\nfunc main(){}\n"
	bx := parseOneBinaryExpr(t, src)
	span := bx.GetSpan()
	if span.Pos.Line != 3 || span.Pos.Column != 11 {
		t.Errorf("Span.Pos = %v; want line 3 col 11", span.Pos)
	}
	if span.End.Line != 3 || span.End.Column != 24 {
		t.Errorf("Span.End = %v; want line 3 col 24", span.End)
	}
}

// TestParseFile_BinaryChain_SpanCorrect_LeftmostParen exercises the
// gate's interaction with *ast.ParenExpr. The outer BinaryExpr
// `((1+2)+3) + 4` has X = inner BinaryExpr `(1+2) + 3` (gate fires
// → fast path uses inner's Span). The inner BinaryExpr in turn has
// X = *ast.ParenExpr (gate does NOT fire → default setSpan runs,
// which uses gon.Pos() = ParenExpr.Pos() = Lparen). The final outer
// Span.Pos must therefore be the column of `(`, not the column of
// the unwrapped `1` inside.
func TestParseFile_BinaryChain_SpanCorrect_LeftmostParen(t *testing.T) {
	t.Parallel()
	// `const x = (1+2) + 3 + 4`
	//  col:    12345678901234567890123
	//                    ^          ^
	// outer Span: col 11 (`(`) through col 24 (after `4`, exclusive).
	src := "package main\n\nconst x = (1+2) + 3 + 4\n\nfunc main(){}\n"
	bx := parseOneBinaryExpr(t, src)
	span := bx.GetSpan()
	if span.Pos.Line != 3 || span.Pos.Column != 11 {
		t.Errorf("Span.Pos = %v; want line 3 col 11 (the `(`)", span.Pos)
	}
	if span.End.Line != 3 || span.End.Column != 24 {
		t.Errorf("Span.End = %v; want line 3 col 24", span.End)
	}
}

// TestParseFile_BinaryChain_SpanCorrect_MultiLine verifies that
// Span.End line tracking propagates correctly through the chain when
// the rightmost leaf is on a later line than the leftmost.
func TestParseFile_BinaryChain_SpanCorrect_MultiLine(t *testing.T) {
	t.Parallel()
	// const x = 1 +     ← line 3, col 11 = `1`
	//     2 +           ← line 4
	//     3             ← line 5, col 5 = `3`, End col 6 (exclusive)
	src := "package main\n\nconst x = 1 +\n    2 +\n    3\n\nfunc main(){}\n"
	bx := parseOneBinaryExpr(t, src)
	span := bx.GetSpan()
	if span.Pos.Line != 3 || span.Pos.Column != 11 {
		t.Errorf("Span.Pos = %v; want line 3 col 11", span.Pos)
	}
	if span.End.Line != 5 || span.End.Column != 6 {
		t.Errorf("Span.End = %v; want line 5 col 6", span.End)
	}
}
