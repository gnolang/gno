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
			desc:    "direct methods: exported keeps empty pkg, unexported stamped with enclosing pkg",
			in:      []FieldType{mkMethod("M", fn), mkMethod("m", fn)},
			pkgPath: "q",
			want:    []want{{"M", ""}, {"m", "q"}},
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

// A same-name/same-package method with conflicting signatures is a
// should-not-happen (go/types rejects it upstream); flatten guards with a panic.
func TestFlattenInterfaceMethods_ConflictPanics(t *testing.T) {
	t.Parallel()
	fn := &FuncType{}
	fn2 := &FuncType{Results: []FieldType{{Type: BoolType}}}
	in := []FieldType{mkMethod("m", fn), mkMethod("m", fn2)}

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on conflicting same-package method, got none")
		}
	}()
	flattenInterfaceMethods(in, "q")
}
