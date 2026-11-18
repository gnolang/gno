package gnolang

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"
)

const realmValueSrcA = `package a

type T struct{ N int }
type A [2]T

func Get() T          { return T{} }
func Take(t T)        {}
func TakeAny(x any)   {}
func TakePtr(t *T)    {}
`

// TestRealmValue type-checks a file of realm gno.land/r/b importing realm
// gno.land/r/a and asserts where checkRealmValues reports ("line:col" in b).
func TestRealmValue(t *testing.T) {
	cases := []struct {
		name string
		body string // after "package b\n\nimport \"gno.land/r/a\"\n\n" (line 5)
		self string // realm the file runs in; default gno.land/r/b
		want []string
	}{
		// M2: a type switch binds v to a.T in that clause only
		{"case_clause", "func f(x any) {\n\tswitch v := x.(type) {\n\tcase a.T:\n\tcase *a.T:\n\t\t_ = v\n\t}\n}", "", []string{"7:2"}},
		{"case_clause_ptr", "func f(x any) {\n\tswitch v := x.(type) {\n\tcase *a.T:\n\t\t_ = v\n\t}\n}", "", nil},
		// M4: an unnamed struct argument becomes an a.T parameter
		{"implicit_arg", "func f() { a.Take(struct{ N int }{1}) }", "", []string{"5:19"}},
		{"implicit_arg_any", "func f() { a.TakeAny(struct{ N int }{1}) }", "", nil},
		// M5: an unnamed result or parameter has no identifier to judge
		{"signature_result", "func f() a.T { panic(0) }", "", []string{"5:10"}},
		{"signature_param", "var g = func(a.T) {}", "", []string{"5:14"}},
		{"signature_ptr", "func f(*a.T) *a.T { return nil }", "", nil},
		// M8: a filetest checked as b but running in realm a holds a's value
		{"own_realm_by_path", "var v a.T", "gno.land/r/a", nil},
		{"other_realm_by_path", "var v a.T", "", []string{"5:5"}},
		// M9..M15: types holding a.T or a.A by value
		{"named_array", "var v a.A", "", []string{"5:5"}},
		{"named_array_ptr", "var v *a.A", "", nil},
		{"array_of_struct", "var v [2]a.T", "", []string{"5:5"}},
		{"array_of_ptr", "var v [2]*a.T", "", nil},
		{"slice", "var v []a.T", "", []string{"5:5"}},
		{"slice_of_ptr", "var v []*a.T", "", nil},
		{"map_key", "var v map[a.T]int", "", []string{"5:5"}},
		{"map_key_ptr", "var v map[*a.T]int", "", nil},
		{"map_elem", "var v map[int]a.T", "", []string{"5:5"}},
		{"map_elem_ptr", "var v map[int]*a.T", "", nil},
		{"chan", "var v chan a.T", "", []string{"5:5"}},
		{"chan_ptr", "var v chan *a.T", "", nil},
		{"struct_field", "var v struct{ F a.T }", "", []string{"5:5", "5:15"}},
		{"struct_field_ptr", "var v struct{ F *a.T }", "", nil},
		{"interface", "var v any = &a.T{}", "", nil},
		{"field_read_of_temporary", "var n = a.Get().N", "", nil},
	}

	fset := token.NewFileSet()
	fa, err := parser.ParseFile(fset, "a.go", realmValueSrcA, 0)
	if err != nil {
		t.Fatal(err)
	}
	pkgA, err := (&types.Config{Importer: importer.Default()}).Check("gno.land/r/a", fset, []*ast.File{fa}, nil)
	if err != nil {
		t.Fatal(err)
	}
	imp := importerFunc(func(path string) (*types.Package, error) {
		if path == "gno.land/r/a" {
			return pkgA, nil
		}
		return nil, fmt.Errorf("unknown import %q", path)
	})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "package b\n\nimport \"gno.land/r/a\"\n\n" + tc.body + "\n"
			fb, err := parser.ParseFile(fset, "b.go", src, 0)
			if err != nil {
				t.Fatal(err)
			}
			info := newRealmValueInfo()
			pkgB, err := (&types.Config{Importer: imp}).Check("gno.land/r/b", fset, []*ast.File{fb}, info)
			if err != nil {
				t.Fatal(err)
			}
			self := tc.self
			if self == "" {
				self = "gno.land/r/b"
			}
			var got []string
			for _, e := range checkRealmValues(fset, []*ast.File{fb}, info, pkgB, self) {
				p := fset.Position(e.(types.Error).Pos)
				got = append(got, fmt.Sprintf("%d:%d", p.Line, p.Column))
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("reports at %v, want %v", got, tc.want)
			}
		})
	}
}

type importerFunc func(string) (*types.Package, error)

func (f importerFunc) Import(path string) (*types.Package, error) { return f(path) }

// A filetest with type errors leaves expressions untyped (nil type).
func TestRealmValueNilType(t *testing.T) {
	if got := derefType(nil); got != nil {
		t.Errorf("derefType(nil) = %v, want nil", got)
	}
	if got := foreignRealmValue(nil, realmSelf{}); got != nil {
		t.Errorf("foreignRealmValue(nil) = %v, want nil", got)
	}
}
