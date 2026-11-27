// Package gen builds type-correct Go programs inside the Gno subset.
// Catalog is the enumerated seed corpus. Generate is a deterministic
// AST generator used by the fuzzer.
package gen

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

type kind uint8

const (
	kBool kind = iota
	kInt
	kInt64
	kUint64
	kString
)

func (k kind) String() string {
	switch k {
	case kBool:
		return "bool"
	case kInt:
		return "int"
	case kInt64:
		return "int64"
	case kUint64:
		return "uint64"
	case kString:
		return "string"
	default:
		return "int"
	}
}

type node struct {
	k    kind
	code string
}

// Generate returns one program for seed. The same seed always returns the
// same source. Programs use only the Gno subset: no goroutines, channels,
// generics, or imports. Division uses a non-zero literal divisor except for
// an occasional trailing panic that both toolchains must report.
func Generate(seed uint64) string {
	//nolint:gosec // G404: deterministic PRNG, the seed is the fuzzer input
	r := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	g := &builder{r: r}
	// id* keeps arithmetic off the constant evaluator. gc rejects a typed
	// constant that overflows (uint64(8)-uint64(40)); Gno wraps it. The
	// generator wants the runtime wrap, which both sides share.
	g.b.WriteString("package main\n\nfunc idInt(x int) int { return x }\nfunc id64(x int64) int64 { return x }\nfunc idU(x uint64) uint64 { return x }\n\nfunc main() {\n\t_ = idInt(0)\n\t_ = id64(0)\n\t_ = idU(0)\n")
	n := 5 + g.r.IntN(8)
	for range n {
		g.statement()
	}
	switch g.r.IntN(10) {
	case 0:
		g.b.WriteString("\tprintln(\"before\")\n\tvar z int = 0\n\tprintln(1 / z)\n")
	case 1:
		g.b.WriteString("\tprintln(\"before\")\n\txs := []int{1, 2}\n\tprintln(xs[5])\n")
	}
	g.b.WriteString("}\n")
	return g.b.String()
}

type builder struct {
	r    *rand.Rand
	vals []node
	b    strings.Builder
	seq  int
}

func (g *builder) statement() {
	switch g.r.IntN(6) {
	case 0, 1:
		g.decl()
	case 2:
		g.println()
	case 3:
		g.ifStmt()
	case 4:
		g.forStmt()
	default:
		g.aggregate()
	}
}

func (g *builder) decl() {
	k := kind(g.r.IntN(int(kString) + 1))
	e := g.expr(k, 0)
	name := fmt.Sprintf("v%d", g.seq)
	g.seq++
	// println uses the variable. gc rejects a local that is never read;
	// leaving it unused would be a compile-error divergence on every seed.
	fmt.Fprintf(&g.b, "\t%s := %s\n\tprintln(%s)\n", name, e.code, name)
	g.vals = append(g.vals, node{k: k, code: name})
}

func (g *builder) println() {
	if len(g.vals) == 0 {
		g.decl()
		return
	}
	v := g.vals[g.r.IntN(len(g.vals))]
	fmt.Fprintf(&g.b, "\tprintln(%s)\n", v.code)
}

func (g *builder) ifStmt() {
	cond := g.expr(kBool, 0)
	fmt.Fprintf(&g.b, "\tif %s {\n", cond.code)
	g.decl()
	g.b.WriteString("\t}\n")
}

func (g *builder) forStmt() {
	n := 1 + g.r.IntN(3)
	name := fmt.Sprintf("i%d", g.seq)
	g.seq++
	fmt.Fprintf(&g.b, "\tfor %s := 0; %s < %d; %s++ {\n\t\tprintln(%s)\n\t}\n", name, name, n, name, name)
}

func (g *builder) aggregate() {
	id := g.seq
	g.seq++
	fmt.Fprintf(&g.b, "\ts%d := []int{1, 2, 3}\n\tprintln(len(s%d), s%d[0], s%d[2])\n", id, id, id, id)
	fmt.Fprintf(&g.b, "\tm%d := map[int]int{1: 10, 2: 20}\n\tsum%d := 0\n\tfor _, mv := range m%d {\n\t\tsum%d += mv\n\t}\n\tprintln(sum%d, len(m%d), m%d[1])\n", id, id, id, id, id, id, id)
}

func (g *builder) expr(k kind, depth int) node {
	if depth > 3 || g.r.IntN(3) == 0 {
		if v, ok := g.pick(k); ok && g.r.IntN(2) == 0 {
			return v
		}
		return g.literal(k)
	}
	switch k {
	case kBool:
		l := g.expr(kInt, depth+1)
		r := g.expr(kInt, depth+1)
		op := []string{"==", "!=", "<", "<=", ">", ">="}[g.r.IntN(6)]
		return node{k: kBool, code: fmt.Sprintf("(%s %s %s)", l.code, op, r.code)}
	case kInt, kInt64, kUint64:
		if g.r.IntN(5) == 0 {
			l := g.expr(k, depth+1)
			return node{k: k, code: fmt.Sprintf("(%s << %d)", l.code, g.r.IntN(4))}
		}
		if g.r.IntN(5) == 0 {
			l := g.expr(k, depth+1)
			return node{k: k, code: fmt.Sprintf("(%s / %s)", l.code, nonzero(k))}
		}
		if k != kInt && g.r.IntN(4) == 0 {
			inner := g.expr(kInt, depth+1)
			return node{k: k, code: fmt.Sprintf("%s(%s)", k, inner.code)}
		}
		l := g.expr(k, depth+1)
		r := g.expr(k, depth+1)
		op := []string{"+", "-", "*", "|", "&", "^"}[g.r.IntN(6)]
		return node{k: k, code: fmt.Sprintf("(%s %s %s)", l.code, op, r.code)}
	case kString:
		if g.r.IntN(2) == 0 {
			l := g.expr(kString, depth+1)
			r := g.expr(kString, depth+1)
			return node{k: kString, code: fmt.Sprintf("(%s + %s)", l.code, r.code)}
		}
		return g.literal(kString)
	default:
		return g.literal(kBool)
	}
}

func (g *builder) literal(k kind) node {
	switch k {
	case kBool:
		if g.r.IntN(2) == 0 {
			return node{k: kBool, code: "true"}
		}
		return node{k: kBool, code: "false"}
	case kInt:
		vals := []string{"0", "1", "2", "-1", "7", "40"}
		return node{k: kInt, code: "idInt(" + vals[g.r.IntN(len(vals))] + ")"}
	case kInt64:
		return node{k: kInt64, code: fmt.Sprintf("id64(int64(%d))", g.r.IntN(21)-10)}
	case kUint64:
		return node{k: kUint64, code: fmt.Sprintf("idU(uint64(%d))", g.r.IntN(20))}
	default:
		vals := []string{`""`, `"a"`, `"hi"`, `"héllo"`}
		return node{k: kString, code: vals[g.r.IntN(len(vals))]}
	}
}

func (g *builder) pick(k kind) (node, bool) {
	var opts []node
	for _, v := range g.vals {
		if v.k == k {
			opts = append(opts, v)
		}
	}
	if len(opts) == 0 {
		return node{}, false
	}
	return opts[g.r.IntN(len(opts))], true
}

func nonzero(k kind) string {
	switch k {
	case kInt64:
		return "int64(2)"
	case kUint64:
		return "uint64(2)"
	default:
		return "2"
	}
}
