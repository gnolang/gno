package gnolang

import (
	"strings"
	"testing"
)

// ft builds a concrete (non-embedded) method entry.
func mkMethod(name string, typ Type) FieldType {
	return FieldType{Name: Name(name), Type: typ}
}

// mkEmbed builds an embed entry: a field whose type is an (already flattened)
// interface from package pkg with the given methods.
func mkEmbed(pkg string, methods ...FieldType) FieldType {
	return FieldType{Type: &InterfaceType{PkgPath: pkg, Methods: methods}}
}

func TestFlattenInterfaceMethods(t *testing.T) {
	t.Parallel()

	fn := &FuncType{} // func()

	type want struct {
		name string
		pkg  string
	}
	cases := []struct {
		desc    string
		in      []FieldType
		pkgPath string
		want    []want
	}{
		{
			// No embed → fast path returns the slice unchanged: a direct
			// unexported method is NOT stamped (idName falls back to the
			// enclosing pkg at TypeID time), so PkgPath stays empty.
			desc:    "no-embed interface returned unchanged (fast path)",
			in:      []FieldType{mkMethod("M", fn), mkMethod("m", fn)},
			pkgPath: "q",
			want:    []want{{"M", ""}, {"m", ""}},
		},
		{
			// On the slow path a same-package unexported method is still left
			// unstamped (only cross-package methods are stamped), so it relies
			// on idName's fallback — matching the fast path.
			desc: "same-package unexported left unstamped even on the slow path",
			in: []FieldType{
				mkEmbed("p", mkMethod("E", fn)),
				mkMethod("m", fn),
			},
			pkgPath: "q",
			want:    []want{{"E", ""}, {"m", ""}},
		},
		{
			desc:    "embedded exported method flattens, pkg stays empty (package-independent identity)",
			in:      []FieldType{mkEmbed("p", mkMethod("E", fn))},
			pkgPath: "q",
			want:    []want{{"E", ""}},
		},
		{
			desc:    "embedded unexported method keeps its origin package, not the enclosing one",
			in:      []FieldType{mkEmbed("p", FieldType{Name: "sec", Type: fn, PkgPath: "p"})},
			pkgPath: "q",
			want:    []want{{"sec", "p"}},
		},
		{
			desc: "diamond: same unexported method via two embeds dedups to one",
			in: []FieldType{
				mkEmbed("p", FieldType{Name: "sec", Type: fn, PkgPath: "p"}),
				mkEmbed("p", FieldType{Name: "sec", Type: fn, PkgPath: "p"}),
			},
			pkgPath: "q",
			want:    []want{{"sec", "p"}},
		},
		{
			// cross-package sec (stamped "p") and a same-package sec (left
			// unstamped, qualified to "q" via fallback) are distinct and coexist.
			desc: "same name, different package: distinct methods coexist",
			in: []FieldType{
				mkEmbed("p", FieldType{Name: "sec", Type: fn, PkgPath: "p"}),
				mkMethod("sec", fn), // declared directly in enclosing q
			},
			pkgPath: "q",
			want:    []want{{"sec", "p"}, {"sec", ""}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			got := flattenInterfaceMethods(tc.in, tc.pkgPath)
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d, want %d (%+v)", len(got), len(tc.want), got)
			}
			for i, w := range tc.want {
				if string(got[i].Name) != w.name || got[i].PkgPath != w.pkg {
					t.Errorf("entry %d = (%s, %q), want (%s, %q)",
						i, got[i].Name, got[i].PkgPath, w.name, w.pkg)
				}
				if got[i].Embedded {
					t.Errorf("entry %d (%s) should not be marked Embedded", i, got[i].Name)
				}
			}
		})
	}
}

// Two embeds contributing the same method name with conflicting signatures is
// a should-not-happen (go/types rejects it upstream); flatten's dedup guards
// with a panic. (A pure direct-method duplicate takes the fast path and is
// instead caught later by sortForPackage at TypeID time.)
func TestFlattenInterfaceMethods_ConflictPanics(t *testing.T) {
	t.Parallel()
	fn := &FuncType{}
	fn2 := &FuncType{Results: []FieldType{{Type: BoolType}}}
	in := []FieldType{
		mkEmbed("p", mkMethod("M", fn)),
		mkEmbed("p", mkMethod("M", fn2)),
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on conflicting embedded method, got none")
		}
	}()
	flattenInterfaceMethods(in, "q")
}

// Regression for the sort/emit provenance mismatch (thehowl review). The same
// interface in package p can hold an unexported method with PkgPath either
// stamped to "p" (the method was hoisted from an embed — slow path) or empty
// (the method was declared directly, so the no-embed fast path leaves it
// unstamped). Both must yield one TypeID: the sort key and the emitted id
// both qualify via the enclosing package, so the order can't flip between
// the two representations.
func TestInterfaceTypeID_PkgPathProvenance(t *testing.T) {
	t.Parallel()
	fn := &FuncType{}
	stamped := &InterfaceType{PkgPath: "p", Methods: []FieldType{
		{Name: "M", Type: fn},
		{Name: "z", Type: fn, PkgPath: "p"}, // hoisted from an embed
	}}
	unstamped := &InterfaceType{PkgPath: "p", Methods: []FieldType{
		{Name: "M", Type: fn},
		{Name: "z", Type: fn}, // declared directly (fast path, unstamped)
	}}
	if stamped.TypeID() != unstamped.TypeID() {
		t.Fatalf("stamped vs empty PkgPath gave different TypeIDs:\n stamped:   %s\n unstamped: %s",
			stamped.TypeID(), unstamped.TypeID())
	}
}

// An InterfaceKind entry in Methods can only be state persisted before
// interface flattening (every construction path flattens), which is
// unsupported — identity already moved. Store bytes are external input, so
// the decode boundary (fillType, reached from both GetTypeSafe and
// fillTypesOfValue) rejects it unconditionally; the interior sites
// (FindEmbeddedFieldType/verifyImplementedBy/TypeID) assume the invariant
// and only assert under -tags debugAssert. Pins the drop of the legacy
// embedded-interface branches; see adr/pr5739.
func TestInterfaceType_UnflattenedIsHardError(t *testing.T) {
	t.Parallel()
	embedded := &InterfaceType{PkgPath: "p", Methods: []FieldType{{Name: "M", Type: &FuncType{}}}}
	legacy := &InterfaceType{PkgPath: "p", Methods: []FieldType{
		{Name: "E", Type: embedded, Embedded: true}, // as decoded from bytes persisted before flattening
	}}

	uses := map[string]func(){
		// no RefType inside, so a nil store never dereferences
		"fillType": func() { fillType(nil, legacy) },
	}
	if debugAssert {
		uses["FindEmbeddedFieldType"] = func() { legacy.FindEmbeddedFieldType("p", "M") }
		uses["verifyImplementedBy"] = func() { legacy.verifyImplementedBy(nil, 0, embedded) }
		uses["TypeID"] = func() { legacy.TypeID() }
	}
	for name, use := range uses {
		t.Run(name, func(t *testing.T) {
			t.Parallel() // uses is read-only on the shared types; TypeID panics before its cache write
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("expected panic on unflattened interface, got none")
				}
				if s, ok := r.(string); !ok || !strings.Contains(s, "unflattened embedded interface") {
					t.Fatalf("expected unflattened-interface panic, got: %v", r)
				}
			}()
			use()
		})
	}
}

// The runtime interface-construction path (doOpInterfaceType) must flatten its
// embeds just like the preprocess path. This path executes during filetests
// (e.g. v.(interface{ Embed })) but its result is never observed for identity
// there, so no filetest pins it; drive the op directly and assert the built
// InterfaceType holds the embed's methods, with no embedded-interface entry.
func TestDoOpInterfaceType_Flattens(t *testing.T) {
	m := NewMachine("p", nil)
	defer m.Release()

	// Two embeds that overlap on B (a diamond): {A,B} and {B,C}. Flattening
	// expands both and dedups B, so the result is exactly {A,B,C} — 3 methods.
	// Without flattening the interface would instead hold 2 embedded-interface
	// entries, so the count distinguishes real flatten+dedup from pass-through.
	fnB := &FuncType{} // shared signature so the overlapping B dedups (not conflicts)
	e1 := &InterfaceType{PkgPath: "p", Methods: []FieldType{{Name: "A", Type: &FuncType{}}, {Name: "B", Type: fnB}}}
	e2 := &InterfaceType{PkgPath: "p", Methods: []FieldType{{Name: "B", Type: fnB}, {Name: "C", Type: &FuncType{}}}}

	// interface{ e1; e2 }: doOpInterfaceType reads len(x.Methods) from the expr
	// and pops one resolved type per method off the value stack.
	m.PushValue(TypedValue{T: gTypeType, V: toTypeValue(FieldType{Type: e1})})
	m.PushValue(TypedValue{T: gTypeType, V: toTypeValue(FieldType{Type: e2})})
	m.PushExpr(&InterfaceTypeExpr{Methods: FieldTypeExprs{{}, {}}})

	m.doOpInterfaceType()

	it := m.PopValue().V.(TypeValue).Type.(*InterfaceType)
	if len(it.Methods) != 3 {
		t.Fatalf("expected 3 flattened+deduped methods (A,B,C), got %d: %+v", len(it.Methods), it.Methods)
	}
	got := map[Name]int{}
	for _, ft := range it.Methods {
		if ft.Type.Kind() == InterfaceKind {
			t.Fatalf("embedded-interface entry survived flattening: %+v", ft)
		}
		got[ft.Name]++
	}
	for _, n := range []Name{"A", "B", "C"} {
		if got[n] != 1 {
			t.Fatalf("method %s appears %d times, want exactly 1: %+v", n, got[n], it.Methods)
		}
	}
}

// The runtime struct-construction path (doOpStructType) must name embedded
// fields from the source expr like the preprocess path (buildFieldTypesAST)
// does. Nothing in the Files suite reaches this op with an embedded field
// (verified by instrumentation), so drive the op directly: the popped
// FieldType carries only the resolved type (the alias is gone), and the name
// must come from the expr as written.
func TestDoOpStructType_EmbedNames(t *testing.T) {
	m := NewMachine("p", nil)
	defer m.Release()

	// struct{ Int; *MyInt; pkg.T; N int } — three embed spellings + one named
	// field (must pass through untouched).
	exprs := FieldTypeExprs{
		{Type: Nx("Int")},                     // alias embed: resolved type int, spelled Int
		{Type: &StarExpr{X: Nx("MyInt")}},     // pointer embed: name from elem
		{Type: Sel(Nx("pkg"), "T")},           // qualified embed: name from selector
		{NameExpr: *Nx("N"), Type: Nx("int")}, // named field: not an embed
	}
	fts := []FieldType{
		{Type: IntType},
		{Type: &PointerType{Elt: IntType}},
		{Type: IntType},
		{Name: "N", Type: IntType},
	}
	for _, ft := range fts {
		m.PushValue(TypedValue{T: gTypeType, V: toTypeValue(ft)})
	}
	m.PushExpr(&StructTypeExpr{Fields: exprs})

	m.doOpStructType()

	st := m.PopValue().V.(TypeValue).Type.(*StructType)
	want := []struct {
		name  Name
		embed bool
	}{{"Int", true}, {"MyInt", true}, {"T", true}, {"N", false}}
	for i, w := range want {
		f := st.Fields[i]
		if f.Name != w.name || f.Embedded != w.embed {
			t.Fatalf("field %d: got (name=%s, embedded=%v), want (name=%s, embedded=%v)",
				i, f.Name, f.Embedded, w.name, w.embed)
		}
	}
}
