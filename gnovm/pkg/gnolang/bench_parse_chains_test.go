package gnolang

import (
	"fmt"
	"strings"
	"testing"
)

// These benchmarks measure Go2Gno parse-time cost on chain-shaped
// AST inputs across the ten reachable affected types — Pos-recursive
// (BinaryExpr, CallExpr, IndexExpr, SelectorExpr, SliceExpr,
// TypeAssertExpr) and End-recursive (StarExpr, UnaryExpr, ArrayType,
// MapType). ChanType is also End-recursive in stdlib but Gno rejects
// `chan` at parse time, so the chain cannot be constructed.
//
// SpanFromGo calls gon.Pos() / gon.End() once per node during the
// Go2Gno walk; for stdlib AST types whose Pos()/End() recurses into a
// same-type child, total parse-time cost is O(N²) without the fix.
//
// Two additional benchmarks document non-DoS shapes for regression
// safety: BinaryChainRightLeaning (parens break End() recursion via
// ParenExpr boundary) and LabeledStmtChain (delegating case protected
// by IsZero short-circuit in setSpan).
//
// Run:
//
//	go test -run=NONE -bench=BenchmarkParseFile_ -benchtime=10x -count=3 ./gnovm/pkg/gnolang/
//
// Asymptotic diagnostic: ns/op ratio per 4× N
//   - perfectly linear     (O(N)):  ~4×
//   - perfectly quadratic (O(N²)): ~16×
//   - empirical "linear-ish" band:  3×–7× (Go parser allocation + GC
//     noise stack on top of the algorithmic cost)

func benchmarkChain(b *testing.B, build func(n int) string) {
	b.Helper()
	for _, n := range []int{1_000, 4_000, 16_000} {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			src := build(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m := NewMachine("test", nil)
				_, err := m.ParseFile("test.gno", src)
				if err != nil {
					b.Fatalf("ParseFile failed: %v", err)
				}
				m.Release()
			}
		})
	}
}

// (*ast.BinaryExpr).Pos() = X.Pos(). Chain shape: `1 + 1 + ... + 1`
// (left-associative). This is the original PR #5648 benchmark, the
// adversarial source that motivated the entire DoS fix.
func BenchmarkParseFile_BinaryChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nconst x = 1")
		for range n {
			sb.WriteString(" + 1")
		}
		sb.WriteString("\n\nfunc main() { _ = x }\n")
		return sb.String()
	})
}

// (*ast.CallExpr).Pos() = Fun.Pos(). To trigger the leftward
// recursion the chain must nest through Fun, not Args. Shape:
// `f()()()...()` — each call's Fun is the previous call.
func BenchmarkParseFile_CallChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nfunc main() { _ = f")
		for range n {
			sb.WriteString("()")
		}
		sb.WriteString(" }\n")
		return sb.String()
	})
}

// (*ast.IndexExpr).Pos() = X.Pos(). Chain shape: a[0][0][0]...[0].
func BenchmarkParseFile_IndexChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nfunc main() { _ = a")
		for range n {
			sb.WriteString("[0]")
		}
		sb.WriteString(" }\n")
		return sb.String()
	})
}

// (*ast.SelectorExpr).Pos() = X.Pos(). Chain shape: a.b.b.b...b.
func BenchmarkParseFile_SelectorChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nfunc main() { _ = a")
		for range n {
			sb.WriteString(".b")
		}
		sb.WriteString(" }\n")
		return sb.String()
	})
}

// (*ast.SliceExpr).Pos() = X.Pos(). Chain shape: s[:][:][:]...[:].
func BenchmarkParseFile_SliceChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nfunc main() { _ = s")
		for range n {
			sb.WriteString("[:]")
		}
		sb.WriteString(" }\n")
		return sb.String()
	})
}

// (*ast.TypeAssertExpr).Pos() = X.Pos(). Chain shape: x.(I).(I).(I)...(I).
func BenchmarkParseFile_TypeAssertChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nfunc main() { _ = x")
		for range n {
			sb.WriteString(".(I)")
		}
		sb.WriteString(" }\n")
		return sb.String()
	})
}

// End-recursive shapes — stdlib defines End() as <field>.End() for these
// types, so a chain on the rightward-recursing field is O(N²) via the
// deferred setSpan's gon.End() call. Pos() is O(1) for all of these
// (opening token / OpPos / Star / Begin / Map), so this isn't a Pos-
// recursion problem like the six above.

// (*ast.StarExpr).End() = X.End(). Chain shape: `*****...*int`.
// Declared as a var type since `***int` as an expression is parsed
// differently — type context is required for the parser to produce
// nested StarExpr.
func BenchmarkParseFile_StarChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nvar x ")
		sb.WriteString(strings.Repeat("*", n))
		sb.WriteString("int\n")
		return sb.String()
	})
}

// (*ast.UnaryExpr).End() = X.End(). Chain shape: `!!!...!x`.
func BenchmarkParseFile_UnaryChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nfunc main() { _ = ")
		sb.WriteString(strings.Repeat("!", n))
		sb.WriteString("x }\n")
		return sb.String()
	})
}

// (*ast.UnaryExpr).End() = X.End() with Op == token.AND — the
// `&`-operator path in Go2Gno produces *RefExpr (separate gate at
// go2gno.go:404-413). Chain shape: `& & & ... &x` (whitespace because
// `&&` would lex as a single logical-AND token).
func BenchmarkParseFile_RefChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nfunc main() { _ = ")
		for range n {
			sb.WriteString("& ")
		}
		sb.WriteString("x }\n")
		return sb.String()
	})
}

// (*ast.ArrayType).End() = Elt.End(). Chain shape: `[1][1]...int`.
func BenchmarkParseFile_ArrayTypeChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nvar x ")
		sb.WriteString(strings.Repeat("[1]", n))
		sb.WriteString("int\n")
		return sb.String()
	})
}

// (*ast.MapType).End() = Value.End(). Chain shape: `map[int]map[int]...T`.
func BenchmarkParseFile_MapTypeChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nvar x ")
		sb.WriteString(strings.Repeat("map[int]", n))
		sb.WriteString("int\n")
		return sb.String()
	})
}

// LabeledStmt chain: `L1: L2: L3: ... LN: ;`. ast.LabeledStmt.End()
// recurses into its Stmt field (which is itself a LabeledStmt in a
// chain). Audit-flagged as a potential O(N²) surface. Verify here.
func BenchmarkParseFile_LabeledStmtChain(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nfunc main() {\n")
		for i := range n {
			fmt.Fprintf(&sb, "L%d:\n", i)
		}
		sb.WriteString("\t;\n}\n")
		return sb.String()
	})
}

// Right-leaning BinaryExpr: `1 + (1 + (1 + ... + 1))`. The only way to
// produce a right-leaning chain in Go is with explicit parens at every
// level (the `+` operator is left-associative by default). Each outer
// level's gon.Y is therefore an *ast.ParenExpr, NOT an *ast.BinaryExpr —
// and ParenExpr.End() = Rparen+1 is O(1), so the End() recursion bug
// I claimed exists actually does NOT occur in Go syntax. This bench
// verifies that claim empirically.
func BenchmarkParseFile_BinaryChainRightLeaning(b *testing.B) {
	benchmarkChain(b, func(n int) string {
		var sb strings.Builder
		sb.WriteString("package main\n\nfunc main() { _ = ")
		for range n {
			sb.WriteString("1 + (")
		}
		sb.WriteString("1")
		sb.WriteString(strings.Repeat(")", n))
		sb.WriteString(" }\n")
		return sb.String()
	})
}
