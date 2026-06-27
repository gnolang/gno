package gnolang

import "testing"

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
			// With an embed present, the slow path runs and a direct
			// unexported method IS stamped with the enclosing package.
			desc: "direct unexported stamped with enclosing pkg when an embed is present",
			in: []FieldType{
				mkEmbed("p", mkMethod("E", fn)),
				mkMethod("m", fn),
			},
			pkgPath: "q",
			want:    []want{{"E", ""}, {"m", "q"}},
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
			desc: "same name, different package: distinct methods coexist",
			in: []FieldType{
				mkEmbed("p", FieldType{Name: "sec", Type: fn, PkgPath: "p"}),
				mkMethod("sec", fn), // declared directly in enclosing q
			},
			pkgPath: "q",
			want:    []want{{"sec", "p"}, {"sec", "q"}},
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
// unstamped; pre-flattening persisted state also decodes to empty). Both must
// yield one TypeID: the sort key and the emitted id both qualify via the
// enclosing package, so the order can't flip between the two representations.
func TestInterfaceTypeID_PkgPathProvenance(t *testing.T) {
	t.Parallel()
	fn := &FuncType{}
	stamped := &InterfaceType{PkgPath: "p", Methods: []FieldType{
		{Name: "M", Type: fn},
		{Name: "z", Type: fn, PkgPath: "p"}, // hoisted from an embed
	}}
	unstamped := &InterfaceType{PkgPath: "p", Methods: []FieldType{
		{Name: "M", Type: fn},
		{Name: "z", Type: fn}, // declared directly, or legacy-decoded
	}}
	if stamped.TypeID() != unstamped.TypeID() {
		t.Fatalf("stamped vs empty PkgPath gave different TypeIDs:\n stamped:   %s\n unstamped: %s",
			stamped.TypeID(), unstamped.TypeID())
	}
}
