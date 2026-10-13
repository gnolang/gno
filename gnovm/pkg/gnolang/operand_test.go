package gnolang

import (
	"testing"
)

// Each statement of f is `_ = <expr>` (or an expression statement); the
// test classifies the expression of each in order.
func TestOperandOf(t *testing.T) {
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", `package main

type S struct{ a [2]int; p *S }

func g() (int, int) { return 1, 2 }
func h()            {}

func f(x int, s S, ps *S, m map[string]int, sl []int, i interface{}) {
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
}

func main() {}
`)
	if err != nil {
		t.Fatal(err)
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	if errs := PreprocessFilesCollect(m.Store, pn, fset); len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	var fd *FuncDecl
	for _, d := range fn.Decls {
		if d, ok := d.(*FuncDecl); ok && d.Name == "f" {
			fd = d
		}
	}
	want := []opMode{
		opVariable, // x
		opConstant, // 1
		opConstant, // "a"
		opValue,    // f: a function is not addressable
		opVariable, // s.a
		opVariable, // s.a[0]
		opVariable, // ps.a
		opMapIndex, // m["k"]
		opVariable, // sl[0]
		opValue,    // S{}.a
		opValue,    // x + 1
		opVariable, // *ps
		opTuple,    // a, b := g()
		opValue,    // i.(int)
		opNoValue,  // h()
		opValue,    // g
		opVariable, // _, _ = a, b: a
	}
	if len(fd.Body) != len(want) {
		t.Fatalf("want %d statements, got %d", len(want), len(fd.Body))
	}
	for i, st := range fd.Body {
		var x Expr
		switch st := st.(type) {
		case *AssignStmt:
			x = st.Rhs[0]
		case *ExprStmt:
			x = st.X
		}
		o := operandOf(m.Store, fd, x)
		if o.mode != want[i] {
			t.Errorf("statement %d (%s): want %s, got %s", i, x, want[i], o.mode)
		}
	}
}
