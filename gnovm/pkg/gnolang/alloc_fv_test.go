package gnolang

import (
	"fmt"
	"testing"
	"time"
)

// TestForeignHeldTypeLinear: a type chain where each level holds the next
// twice (struct{ a, b [n]Next }) must not cost 2^depth in the held-realm walk.
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
		ft := firstForeignHeld(next, rid)
		d := time.Since(start)
		t.Logf("rid %v: %v (found %v)", rid, d, ft != nil)
		if d > 50*time.Millisecond {
			t.Fatalf("held-realm walk took %v", d)
		}
	}
}

// BenchmarkCopyNested copies an 85-node nested struct, like zrealm_fv_cost:
// the type check runs once per nested struct copy.
func BenchmarkCopyNested(b *testing.B) {
	const pkg = "gno.land/r/zchalcpu"
	mk := func(name string, fields []FieldType) Type {
		return &DeclaredType{PkgPath: pkg, Name: Name(name), Base: &StructType{PkgPath: pkg, Fields: fields}}
	}
	four := func(t Type) []FieldType {
		return []FieldType{{Name: "X", Type: t}, {Name: "Y", Type: t}, {Name: "Z", Type: t}, {Name: "W", Type: t}}
	}
	ints := []FieldType{{Name: "A", Type: IntType}, {Name: "B", Type: IntType}, {Name: "C", Type: IntType}, {Name: "D", Type: IntType}}
	l0 := mk("L0", ints)
	l1 := mk("L1", four(l0))
	l2 := mk("L2", four(l1))
	l3 := mk("L3", four(l2))
	alloc := &Allocator{
		maxBytes:         1 << 40,
		currentRealmID:   PkgIDFromPkgPath(pkg),
		currentRealmPath: pkg,
	}
	root := buildNestedTV(l3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.Copy(alloc)
	}
}

func buildNestedTV(t Type) TypedValue {
	st, ok := baseOf(t).(*StructType)
	if !ok {
		return TypedValue{T: t}
	}
	fields := make([]TypedValue, len(st.Fields))
	for i, f := range st.Fields {
		fields[i] = buildNestedTV(f.Type)
	}
	return TypedValue{T: t, V: &StructValue{Fields: fields}}
}

// TestHeldRealmsEmptyCached: a /p/ chain where each level holds the next in
// two fields. The first walk must be linear (each type visited once) and the
// empty answer must be cached, so a repeat is a lookup.
func TestHeldRealmsEmptyCached(t *testing.T) {
	const pkg = "gno.land/p/zchalcpu"
	var next Type = &DeclaredType{PkgPath: pkg, Name: "L0", Base: &StructType{
		PkgPath: pkg, Fields: []FieldType{{Name: "N", Type: IntType}},
	}}
	for i := 1; i <= 26; i++ {
		next = &DeclaredType{PkgPath: pkg, Name: Name(fmt.Sprintf("L%d", i)), Base: &StructType{
			PkgPath: pkg, Fields: []FieldType{{Name: "X", Type: next}, {Name: "Y", Type: next}},
		}}
	}
	rid := PkgIDFromPkgPath("gno.land/r/other")
	start := time.Now()
	ft := firstForeignHeld(next, rid)
	first := time.Since(start)
	start = time.Now()
	firstForeignHeld(next, rid)
	second := time.Since(start)
	t.Logf("first %v, second %v", first, second)
	if heldRealmsOf(next) == nil {
		t.Fatal("computed empty answer is not cached")
	}
	if ft != nil {
		t.Fatalf("unexpected foreign type %v", ft)
	}
	if first > 50*time.Millisecond {
		t.Fatalf("first walk took %v", first)
	}
	if second > time.Millisecond {
		t.Fatalf("cached walk took %v", second)
	}
}

// BenchmarkCopyNestedP is BenchmarkCopyNested with /p/ types: every type
// holds no realm, so the held-realm answer is the empty one.
func BenchmarkCopyNestedP(b *testing.B) {
	const pkg = "gno.land/p/zchalcpu"
	mk := func(name string, fields []FieldType) Type {
		return &DeclaredType{PkgPath: pkg, Name: Name(name), Base: &StructType{PkgPath: pkg, Fields: fields}}
	}
	four := func(t Type) []FieldType {
		return []FieldType{{Name: "X", Type: t}, {Name: "Y", Type: t}, {Name: "Z", Type: t}, {Name: "W", Type: t}}
	}
	ints := []FieldType{{Name: "A", Type: IntType}, {Name: "B", Type: IntType}, {Name: "C", Type: IntType}, {Name: "D", Type: IntType}}
	l0 := mk("L0", ints)
	l1 := mk("L1", four(l0))
	l2 := mk("L2", four(l1))
	l3 := mk("L3", four(l2))
	alloc := &Allocator{
		maxBytes:         1 << 40,
		currentRealmID:   PkgIDFromPkgPath("gno.land/r/zchalb"),
		currentRealmPath: "gno.land/r/zchalb",
	}
	root := buildNestedTV(l3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.Copy(alloc)
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
