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
