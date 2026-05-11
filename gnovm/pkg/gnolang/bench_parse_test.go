package gnolang

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkParseFile_BinaryChain measures Go2Gno performance on a
// left-leaning `1 + 1 + ... + 1` const expression of varying depth.
//
// Run:
//
//	go test -run=NONE -bench=BenchmarkParseFile_BinaryChain \
//	    -benchtime=10x -count=3 ./gnovm/pkg/gnolang/
//
// The ratio of ns/op between successive N values is the diagnostic:
//   - linear (O(N)):    4× N → ~4× ns/op
//   - quadratic (O(N²)): 4× N → ~16× ns/op
//
// The advisory shape behind this benchmark is a parse-time DoS:
// (*ast.BinaryExpr).Pos() in the standard library recurses leftward
// (Pos = X.Pos()), and gnolang.SpanFromGo calls Pos() once per AST
// node during Go2Gno. For a left-leaning chain of depth N, total
// parse-time cost is O(N²). The fix in Go2Gno's *ast.BinaryExpr case
// makes it O(N) by setting the Span from already-translated
// children's spans.
func BenchmarkParseFile_BinaryChain(b *testing.B) {
	for _, n := range []int{1_000, 4_000, 16_000} {
		b.Run(fmt.Sprintf("N=%d", n), func(b *testing.B) {
			var sb strings.Builder
			sb.WriteString("package main\n\nconst x = 1")
			for i := 0; i < n; i++ {
				sb.WriteString(" + 1")
			}
			sb.WriteString("\n\nfunc main() { _ = x }\n")
			src := sb.String()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m := NewMachine("test", nil)
				_, err := m.ParseFile("test.gno", src)
				if err != nil {
					b.Fatal(err)
				}
				m.Release()
			}
		})
	}
}
