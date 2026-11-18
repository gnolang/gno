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
