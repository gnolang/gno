package gnolang

import (
	"fmt"
	"testing"
	"time"
)

// TestForeignHeldTypeLinear: a type chain where each level holds the next
// twice (struct{ a, b [n]Next }) must not cost 2^depth in foreignHeldType.
func TestForeignHeldTypeLinear(t *testing.T) {
	for _, n := range []int{0, 1} {
		t.Run(fmt.Sprintf("len%d", n), func(t *testing.T) {
			testForeignHeldTypeChain(t, n)
		})
	}
}

func testForeignHeldTypeChain(t *testing.T, n int) {
	const pkg = "gno.land/r/zchalcpu"
	var next Type = &DeclaredType{PkgPath: pkg, Name: "A26", Base: &StructType{
		PkgPath: pkg, Fields: []FieldType{{Name: "a", Type: IntType}},
	}}
	for i := 25; i >= 0; i-- {
		arr := &ArrayType{Len: n, Elt: next}
		next = &DeclaredType{PkgPath: pkg, Name: Name(fmt.Sprintf("A%d", i)), Base: &StructType{
			PkgPath: pkg, Fields: []FieldType{{Name: "a", Type: arr}, {Name: "b", Type: arr}},
		}}
	}
	// The own realm, so the walk never stops early; then another realm.
	for _, rid := range []PkgID{PkgIDFromPkgPath(pkg), PkgIDFromPkgPath("gno.land/r/other")} {
		start := time.Now()
		ft := foreignHeldType(next, rid)
		d := time.Since(start)
		t.Logf("rid %v: %v (found %v)", rid, d, ft != nil)
		if d > 50*time.Millisecond {
			t.Fatalf("foreignHeldType took %v", d)
		}
	}
}

const handleIterateSrc = `package test

var sum int

func visit(i int) bool { sum += i; return false }

func iterate(n int, f func(int) bool) {
	for i := 0; i < n; i++ {
		if f(i) {
			return
		}
	}
}

func Loop() {
	for k := 0; k < 1000; k++ {
		iterate(10, visit)
	}
}
`

// BenchmarkHandleIterate passes a declared callback to an iterate helper 1000
// times per op; each pass makes one handle.
func BenchmarkHandleIterate(b *testing.B) {
	m := NewMachine("test", nil)
	m.RunFiles(m.MustParseFile("main.gno", handleIterateSrc))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Eval(Call("Loop"))
	}
}

const stripBenchSrc = `package bench

var sum int

type P struct {
	A, B int
	S    string
}

var ps [8]P

type In struct{ A, B int }

type Out struct {
	I In
	S string
}

var os [8]Out

func visit(i int) bool { sum += i; return false }

func iterate(n int, f func(int) bool) {
	for i := 0; i < n; i++ {
		if f(i) {
			return
		}
	}
}

// Iterate hands the realm's own callback (a closure made here) on 1000 times.
func Iterate() {
	f := func(i int) bool { return visit(i) }
	for k := 0; k < 1000; k++ {
		iterate(10, f)
	}
}

// Copies copies a struct without function fields 1000 times.
func Copies() {
	for k := 0; k < 1000; k++ {
		x := ps[k%8]
		x.A++
		ps[(k+1)%8] = x
	}
}

// NestedCopies copies a struct nesting a struct by value 1000 times.
func NestedCopies() {
	for k := 0; k < 1000; k++ {
		x := os[k%8]
		x.I.A++
		os[(k+1)%8] = x
	}
}
`

// BenchmarkStripOwnIterate and BenchmarkStripStructCopy run in a realm with
// its own callback and with structs holding no function: nothing is stripped.
func BenchmarkStripOwnIterate(b *testing.B)  { benchStrip(b, "Iterate") }
func BenchmarkStripStructCopy(b *testing.B)  { benchStrip(b, "Copies") }
func BenchmarkNestedStructCopy(b *testing.B) { benchStrip(b, "NestedCopies") }

func benchStrip(b *testing.B, fn string) {
	b.Helper()
	m := NewMachine("gno.land/r/bench", nil)
	m.RunFiles(m.MustParseFile("bench.gno", stripBenchSrc))
	if m.Realm == nil {
		b.Fatal("no realm")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Eval(Call(Nx(fn)))
	}
}
