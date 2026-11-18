package gnolang

import (
	"testing"
	"time"
)

// TestForeignHeldTypeLinear: a type chain where each level holds the next
// twice (struct{ a, b [0]Next }) must not cost 2^depth in foreignHeldType.
func TestForeignHeldTypeLinear(t *testing.T) {
	const pkg = "gno.land/r/zchalcpu"
	var next Type = &DeclaredType{PkgPath: pkg, Name: "A26", Base: &StructType{
		PkgPath: pkg, Fields: []FieldType{{Name: "a", Type: IntType}},
	}}
	for i := 25; i >= 0; i-- {
		arr := &ArrayType{Len: 0, Elt: next}
		next = &DeclaredType{PkgPath: pkg, Name: Name("A" + string(rune('a'+i))), Base: &StructType{
			PkgPath: pkg, Fields: []FieldType{{Name: "a", Type: arr}, {Name: "b", Type: arr}},
		}}
	}
	// The own realm, so the walk never stops early; the other realm sees a
	// chain with nothing foreign below level 0 either.
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
