package gnolang

import "testing"

// TestSpan verifies setSpanFromLeftChild / setSpanFromRightChild
// produce Spans identical to what the default SpanFromGo path would
// have set — covering all 11 affected AST types plus the gate-
// boundary cases (paren-wrapped operands at either end).
//
// Per-case columns are hand-counted off the source. Each source has
// `package main\n\n` as a 14-char prefix so the interesting line is
// always line 3.
func TestSpan(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		src  string
		// expected Span on the top-level chain expression
		posLine, posCol int
		endLine, endCol int
	}{
		// ── *ast.BinaryExpr (Pos-recursive) ─────────────────────────
		// Gate: gon.X is *ast.BinaryExpr.
		// Basic left-leaning chain — gate fires at outer and every
		// inner level. Fast path uses translated Left's Pos +
		// gon.End() (which equals Y.End() = '4'.End() for a leaf Y).
		{
			"binary/chain",
			"package main\n\nconst x = 1 + 2 + 3 + 4\n",
			3, 11, 3, 24,
		},
		// Leftmost X is *ast.ParenExpr — outer gate DOES NOT fire.
		// Default path runs; gon.Pos() = ParenExpr.Pos = '('.
		{
			"binary/leftmost_paren",
			"package main\n\nconst x = (1+2) + 3 + 4\n",
			3, 11, 3, 24,
		},
		// **@thehowl's regression case.** Y is *ast.ParenExpr while
		// the outer gate fires (X is BinaryExpr). Our helper reads
		// End from gon.End() = ParenExpr.End() = position after ')'.
		// PR #5648's helper reads End from bx.Right.GetSpan().End,
		// which after ParenExpr-unwrap (go2gno.go:264) is the
		// position after 'd' — off by one column. This test FAILS
		// with #5648's approach and PASSES with ours.
		{
			"binary/rightmost_paren",
			"package main\n\nconst x = a + b + (c + d)\n",
			3, 11, 3, 26,
		},
		// Chain across multiple lines — End line/col must propagate
		// through the chain.
		{
			"binary/multi_line",
			"package main\n\nconst x = 1 +\n    2 +\n    3\n",
			3, 11, 5, 6,
		},
		// Exercises the Y-side gate via precedence: `1 + 2 * 3`
		// parses as BinaryExpr{+, X: 1, Y: BinaryExpr{*, 2, 3}}.
		// gon.X is a leaf (X gate doesn't fire); gon.Y is a
		// BinaryExpr (Y gate fires). Fast path uses gon.Pos()
		// (= 1.Pos()) + translated Right's End (= 3.End()+1).
		{
			"binary/precedence_y_chain",
			"package main\n\nconst x = 1 + 2 * 3\n",
			3, 11, 3, 20,
		},
		// Both X and Y are BinaryExpr — `1 * 2 + 3 * 4` parses as
		// BinaryExpr{+, X: BinaryExpr{*, 1, 2}, Y: BinaryExpr{*, 3, 4}}.
		// The X gate wins (else-if ordering); the Y child still has
		// its Span set correctly via its own recursive walk.
		{
			"binary/both_sides_chain",
			"package main\n\nconst x = 1 * 2 + 3 * 4\n",
			3, 11, 3, 24,
		},

		// ── *ast.CallExpr (Pos-recursive) ───────────────────────────
		// Gate: gon.Fun is *ast.CallExpr. Chain: each call's Fun is
		// the previous call.
		{
			"call/chain",
			"package main\n\nvar x = f()()()\n",
			3, 9, 3, 16,
		},
		// Leftmost Fun is *ast.ParenExpr — innermost gate doesn't
		// fire. Outer levels still chain because each level's Fun
		// is a CallExpr, so the fast path fires there.
		{
			"call/paren_fun",
			"package main\n\nvar x = (f)()()\n",
			3, 9, 3, 16,
		},

		// ── *ast.IndexExpr (Pos-recursive) ──────────────────────────
		// Gate: gon.X is *ast.IndexExpr.
		{
			"index/chain",
			"package main\n\nvar x = a[0][1][2]\n",
			3, 9, 3, 19,
		},
		{
			"index/paren_x",
			"package main\n\nvar x = (a[0])[1][2]\n",
			3, 9, 3, 21,
		},

		// ── *ast.SelectorExpr (Pos-recursive) ───────────────────────
		// Gate: gon.X is *ast.SelectorExpr.
		{
			"selector/chain",
			"package main\n\nvar x = a.b.c.d\n",
			3, 9, 3, 16,
		},
		{
			"selector/paren_x",
			"package main\n\nvar x = (a.b).c.d\n",
			3, 9, 3, 18,
		},

		// ── *ast.SliceExpr (Pos-recursive) ──────────────────────────
		// Gate: gon.X is *ast.SliceExpr.
		{
			"slice/chain",
			"package main\n\nvar x = s[:][:][:]\n",
			3, 9, 3, 19,
		},
		{
			"slice/paren_x",
			"package main\n\nvar x = (s[:])[:][:]\n",
			3, 9, 3, 21,
		},

		// ── *ast.TypeAssertExpr (Pos-recursive) ─────────────────────
		// Gate: gon.X is *ast.TypeAssertExpr.
		{
			"typeassert/chain",
			"package main\n\nvar x = y.(I).(I).(I)\n",
			3, 9, 3, 22,
		},
		{
			"typeassert/paren_x",
			"package main\n\nvar x = (y.(I)).(I).(I)\n",
			3, 9, 3, 24,
		},

		// ── *ast.StarExpr (End-recursive) ───────────────────────────
		// Gate: gon.X is *ast.StarExpr. Fast path uses gon.Pos()
		// (Star, O(1)) + translated rightChild.End.
		{
			"star/chain",
			"package main\n\nvar x ***int\n",
			3, 7, 3, 13,
		},
		// Innermost X is *ast.ParenExpr — innermost gate doesn't
		// fire; outer levels still chain via StarExpr inner.
		{
			"star/paren_x",
			"package main\n\nvar x **(*int)\n",
			3, 7, 3, 15,
		},

		// ── *ast.UnaryExpr (End-recursive) ──────────────────────────
		// Gate: gon.X is *ast.UnaryExpr.
		{
			"unary/chain",
			"package main\n\nvar x = !!!y\n",
			3, 9, 3, 13,
		},
		{
			"unary/paren_x",
			"package main\n\nvar x = !!(!y)\n",
			3, 9, 3, 15,
		},
		// ── *ast.UnaryExpr with Op == token.AND (produces *RefExpr).
		// The Op == AND branch (go2gno.go:404-413) is a separate
		// gated path from the regular UnaryExpr branch. Source uses
		// `& &y` with whitespace because `&&` would lex as a single
		// logical-AND token.
		{
			"ref/chain",
			"package main\n\nvar x = & & &y\n",
			3, 9, 3, 15,
		},

		// ── *ast.ArrayType (End-recursive) ──────────────────────────
		// Gate: gon.Elt is *ast.ArrayType. No paren-wrap test because
		// parens around type expressions aren't permitted by the
		// parser in this position.
		{
			"arraytype/chain",
			"package main\n\nvar x [1][2][3]int\n",
			3, 7, 3, 19,
		},

		// ── *ast.ChanType — n/a on this branch: rejected at parse
		// time via `panicWithPos("channels are not permitted")`,
		// so a ChanType chain cannot be constructed by Go2Gno.

		// ── *ast.MapType (End-recursive) ────────────────────────────
		// Gate: gon.Value is *ast.MapType.
		{
			"maptype/chain",
			"package main\n\nvar x map[int]map[int]int\n",
			3, 7, 3, 26,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			e := parseTopExpr(t, tc.src)
			s := e.GetSpan()
			if s.Pos.Line != tc.posLine || s.Pos.Column != tc.posCol {
				t.Errorf("Span.Pos = %d:%d; want %d:%d",
					s.Pos.Line, s.Pos.Column, tc.posLine, tc.posCol)
			}
			if s.End.Line != tc.endLine || s.End.Column != tc.endCol {
				t.Errorf("Span.End = %d:%d; want %d:%d",
					s.End.Line, s.End.Column, tc.endLine, tc.endCol)
			}
		})
	}
}

// parseTopExpr parses src and returns the top-level chain expression
// inside the first ValueDecl — either Values[0] (`var/const x = ...`)
// or Type (`var x <type>`). Tests assume a single ValueDecl; if a
// future test needs imports or other prior decls, scan for the first
// *ValueDecl instead.
func parseTopExpr(t *testing.T, src string) Expr {
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
	if len(vd.Values) > 0 && vd.Values[0] != nil {
		return vd.Values[0]
	}
	if vd.Type != nil {
		return vd.Type
	}
	t.Fatalf("ValueDecl has neither Value nor Type")
	return nil
}

// TestSpan_InnerNodes verifies the fast-path helpers produce correct
// Spans not just at the outermost chain node but at every inner level.
// Without this test, a regression where the helper computes the right
// outer Span but corrupt inner Spans would ship green. The induction
// the helpers rely on (`leftChild.GetSpan().Pos` /
// `rightChild.GetSpan().End` already correct from the recursive walk)
// only holds if every inner level is correct.
func TestSpan_InnerNodes(t *testing.T) {
	t.Parallel()

	// `const x = 1 + 2 + 3 + 4` parses as ((1+2)+3)+4. Verify span
	// at every left-chain level.
	//   col: 1234567890123456789012345
	//                  ^         ^
	//   - innermost (1+2):       cols 11..16
	//   - middle    (1+2+3):     cols 11..20
	//   - outer     (1+2+3+4):   cols 11..24
	src := "package main\n\nconst x = 1 + 2 + 3 + 4\n"
	e := parseTopExpr(t, src)
	outer, ok := e.(*BinaryExpr)
	if !ok {
		t.Fatalf("outer = %T; want *BinaryExpr", e)
	}
	mid, ok := outer.Left.(*BinaryExpr)
	if !ok {
		t.Fatalf("outer.Left = %T; want *BinaryExpr", outer.Left)
	}
	innermost, ok := mid.Left.(*BinaryExpr)
	if !ok {
		t.Fatalf("mid.Left = %T; want *BinaryExpr", mid.Left)
	}
	check := func(name string, n Node, wantPosCol, wantEndCol int) {
		s := n.GetSpan()
		if s.Pos.Line != 3 || s.Pos.Column != wantPosCol {
			t.Errorf("%s Span.Pos = %d:%d; want 3:%d",
				name, s.Pos.Line, s.Pos.Column, wantPosCol)
		}
		if s.End.Line != 3 || s.End.Column != wantEndCol {
			t.Errorf("%s Span.End = %d:%d; want 3:%d",
				name, s.End.Line, s.End.Column, wantEndCol)
		}
	}
	check("innermost (1+2)", innermost, 11, 16)
	check("middle (1+2+3)", mid, 11, 20)
	check("outer (1+2+3+4)", outer, 11, 24)
}
