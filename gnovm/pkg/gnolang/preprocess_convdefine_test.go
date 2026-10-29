package gnolang

import (
	"fmt"
	"testing"
)

// staticTypesOf preprocesses src and returns, per block declaring name in
// source order, the static type the block records for it.
func staticTypesOf(t *testing.T, src string, name Name) (out []string) {
	t.Helper()
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", src)
	if err != nil {
		t.Fatal(err)
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	PredefineFileSet(m.Store, pn, fset)
	Preprocess(m.Store, pn, fn)
	Transcribe(fn, func(_ []Node, _ TransField, _ int, n Node, stage TransStage) (Node, TransCtrl) {
		if bn, ok := n.(BlockNode); ok && stage == TRANS_ENTER {
			sb := bn.GetStaticBlock()
			if idx, ok := sb.GetLocalIndex(name); ok {
				out = append(out, fmt.Sprintf("%v", sb.Types[idx]))
			}
		}
		return n, TRANS_CONTINUE
	})
	return out
}

// A name defined from a conversion to an interface type has the interface
// as its static type in every block that declares it, including the
// if/switch clause blocks that copy the init statement's names.
func TestPreprocessDefineConvIface(t *testing.T) {
	cases := []struct {
		src  string
		want []string
	}{
		{"const a = 0\nfunc main() { i := (interface{})(a); _ = i }", []string{"interface {}"}},
		{"func main() { i := (interface{})(5); _ = i }", []string{"interface {}"}},
		{"func main() { a := 1; i := interface{}(a); _ = i }", []string{"interface {}"}},
		{"func main() { a := 1; i := any(a); _ = i }", []string{"interface {}"}},
		{"type I interface{ M() }\ntype T int\nfunc (T) M() {}\nfunc main() { var a T; i := I(a); _ = i }", []string{"main.I"}},
		{"func f() interface{} { return 1 }\nfunc main() { i := f(); _ = i }", []string{"interface {}"}},
		{"type N int\nfunc main() { i := N(5); _ = i }", []string{"main.N"}},
		{"func main() { var i interface{} = 5; _ = i }", []string{"interface {}"}},
		// The if/switch clause blocks copy the init's names: the if or
		// switch block, then one block per clause (an if always has two).
		{"const a = 0\nfunc main() { switch i := (interface{})(a); i.(type) { case [0]string: default: } }",
			[]string{"interface {}", "interface {}", "interface {}"}},
		{"func main() { if i := any(0); true { _ = i } else { _ = i } }",
			[]string{"interface {}", "interface {}", "interface {}"}},
		{"type N int\nfunc main() { if i := N(0); true { _ = i } }", []string{"main.N", "main.N", "main.N"}},
		{"func main() { if i := 0; true { _ = i } }", []string{"int", "int", "int"}},
	}
	for _, c := range cases {
		got := staticTypesOf(t, "package main\n"+c.src, "i")
		if fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("%q: static types of i = %v, want %v", c.src, got, c.want)
		}
	}
}
