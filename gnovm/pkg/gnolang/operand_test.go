package gnolang

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// funcBody preprocesses src and returns the body of its function f.
func funcBody(t *testing.T, store Store, src string) (Store, *FuncDecl) {
	t.Helper()
	m := NewMachine("main", store)
	fn, err := m.ParseFile("main.gno", src)
	if err != nil {
		t.Fatal(err)
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	if errs := PreprocessFilesCollect(m.Store, pn, fset); len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	for _, d := range fn.Decls {
		if d, ok := d.(*FuncDecl); ok && d.Name == "f" {
			return m.Store, d
		}
	}
	t.Fatal("no func f")
	return nil, nil
}

// stmtExpr picks the expression a test statement classifies: the right
// side of `_ = x`, an expression statement, the value of `var p T = v`,
// or the first statement of a type switch's first clause (whose block
// is then the one to classify in).
func stmtExpr(last BlockNode, st Stmt) (BlockNode, Expr) {
	switch st := st.(type) {
	case *AssignStmt:
		return last, st.Rhs[0]
	case *ExprStmt:
		return last, st.X
	case *DeclStmt:
		return last, st.Body[0].(*ValueDecl).Values[0]
	case *SwitchStmt:
		cl := &st.Clauses[0]
		return stmtExpr(cl, cl.Body[0])
	}
	return last, nil
}

// Each statement of f is classified in order; `_ = <call>` entries
// marked callee classify the call's function instead.
func TestOperandOf(t *testing.T) {
	store, fd := funcBody(t, nil, `package main

type S struct{ a [2]int; p *S }

func (S) M()   {}
func (*S) PM() {}

type I interface{ M() }

func g() (int, int) { return 1, 2 }
func h()            {}
func fs() S         { return S{} }

func f(x int, s S, ps *S, m map[string]int, sl []int, i interface{}, str string, ma map[string][2]int, pa *[2]int, ii I) {
	_ = x
	_ = 1
	_ = "a"
	_ = f
	_ = s.a
	_ = s.a[0]
	_ = ps.a
	_ = m["k"]
	_ = sl[0]
	_ = S{}.a
	_ = x + 1
	_ = *ps
	a, b := g()
	_ = i.(int)
	h()
	_ = g
	_, _ = a, b
	var p *int = nil
	_ = int64(x)
	_ = len(sl)
	_ = str[0]
	_ = fs().a
	_ = ma["k"][0]
	_ = pa[0]
	_ = s.M
	_ = ps.PM
	_ = ii.M
	_ = (*ps).a
	switch v := i.(type) {
	case int:
		_ = v
	}
	_ = p
}

func main() {}
`)
	type want struct {
		mode   opMode
		callee bool
	}
	wants := []want{
		{opVariable, false}, // x
		{opConstant, false}, // 1
		{opConstant, false}, // "a"
		{opValue, false},    // f: a function is not addressable
		{opVariable, false}, // s.a
		{opVariable, false}, // s.a[0]
		{opVariable, false}, // ps.a
		{opMapIndex, false}, // m["k"]
		{opVariable, false}, // sl[0]
		{opValue, false},    // S{}.a
		{opValue, false},    // x + 1
		{opVariable, false}, // *ps
		{opTuple, false},    // a, b := g()
		{opValue, false},    // i.(int)
		{opNoValue, false},  // h()
		{opValue, false},    // g
		{opVariable, false}, // _, _ = a, b: a
		{opNil, false},      // var p *int = nil
		{opType, true},      // int64(x): int64
		{opBuiltin, true},   // len(sl): len
		{opValue, false},    // str[0]: a string element
		{opValue, false},    // fs().a: a call result's field
		{opValue, false},    // ma["k"][0]: an array element in a map
		{opVariable, false}, // pa[0]
		{opValue, false},    // s.M: a method value
		{opValue, false},    // ps.PM
		{opValue, false},    // ii.M
		{opVariable, false}, // (*ps).a
		{opVariable, false}, // v of a type switch
		{opVariable, false}, // p
	}
	if len(fd.Body) != len(wants) {
		t.Fatalf("want %d statements, got %d", len(wants), len(fd.Body))
	}
	for i, st := range fd.Body {
		last, x := stmtExpr(fd, st)
		if wants[i].callee {
			x = x.(*CallExpr).Func
		}
		o := operandOf(store, last, x)
		if o.mode != wants[i].mode {
			t.Errorf("statement %d (%s): want %s, got %s", i, x, wants[i].mode, o.mode)
		}
	}
}

// A package name, a package variable and a package function.
func TestOperandOfPackage(t *testing.T) {
	db := memdb.NewMemDB()
	var store Store = NewStore(nil, dbadapter.StoreConstructor(db, stypes.StoreOptions{}), iavl.StoreConstructor(db, stypes.StoreOptions{}))
	m := NewMachine("gno.land/p/demo/pp", store)
	m.RunMemPackageWithOverrides(&std.MemPackage{
		Type:  MPUserProd,
		Name:  "pp",
		Path:  "gno.land/p/demo/pp",
		Files: []*std.MemFile{{Name: "pp.gno", Body: "package pp\n\nvar V int\n\nfunc F() {}\n"}},
	}, true)
	store, fd := funcBody(t, store, `package main

import "gno.land/p/demo/pp"

func f() {
	_ = pp.V
	_ = pp.F
	pp.F()
}

func main() {}
`)
	_, x := stmtExpr(fd, fd.Body[0])
	if o := operandOf(store, fd, x.(*SelectorExpr).X); o.mode != opPackage {
		t.Errorf("pp: want %s, got %s", opPackage, o.mode)
	}
	if o := operandOf(store, fd, x); o.mode != opVariable {
		t.Errorf("pp.V: want %s, got %s", opVariable, o.mode)
	}
	_, x = stmtExpr(fd, fd.Body[1])
	if o := operandOf(store, fd, x); o.mode != opValue {
		t.Errorf("pp.F: want %s, got %s", opValue, o.mode)
	}
}
