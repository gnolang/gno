package gnolang

import (
	"fmt"
	"strings"
	"testing"
)

// collectErrors preprocesses one file in accumulate mode and returns the
// error messages, stripped of their location prefix.
func collectErrors(t *testing.T, src string) []string {
	t.Helper()
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", src)
	if err != nil {
		t.Fatal(err)
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	errs := PreprocessFilesCollect(m.Store, pn, fset)
	var msgs []string
	for _, e := range errs {
		msg := e.Unwrap().Error()
		if i := strings.Index(msg, ": "); i >= 0 {
			msg = msg[i+2:]
		}
		msgs = append(msgs, msg)
	}
	return msgs
}

func TestPreprocessCollect(t *testing.T) {
	t.Run("every failing statement is reported", func(t *testing.T) {
		msgs := collectErrors(t, `package main

func main() {
	var a int = "s"
	b := 1
	b = "x"
	c := 2
	println(a, b, c)
}
`)
		if len(msgs) != 2 {
			t.Fatalf("want 2 errors, got %d: %v", len(msgs), msgs)
		}
		if !strings.Contains(msgs[0], "cannot use") || !strings.Contains(msgs[1], "cannot use") {
			t.Fatalf("unexpected messages: %v", msgs)
		}
	})

	t.Run("names declared by a dropped statement are poisoned", func(t *testing.T) {
		msgs := collectErrors(t, `package main

func main() {
	x := "s" + 1
	println(x)
	println(y)
}
`)
		// x's declaration fails; println(x) must not add "name x not
		// declared", while println(y) is a real error.
		if len(msgs) != 2 {
			t.Fatalf("want 2 errors, got %d: %v", len(msgs), msgs)
		}
		if !strings.Contains(msgs[1], "name y not declared") {
			t.Fatalf("unexpected messages: %v", msgs)
		}
	})

	t.Run("top-level declarations and nested blocks", func(t *testing.T) {
		msgs := collectErrors(t, `package main

var G int = "g"

func f() int { return "s" }

func main() {
	if true {
		var n string = 1
		println(n)
	}
	println(G, f())
}
`)
		if len(msgs) != 3 {
			t.Fatalf("want 3 errors, got %d: %v", len(msgs), msgs)
		}
	})

	t.Run("Preprocess still stops at the first error", func(t *testing.T) {
		m := NewMachine("main", nil)
		defer m.Release()
		fn, err := m.ParseFile("main.gno", `package main

func main() {
	var a int = "s"
	var b int = "t"
	println(a, b)
}
`)
		if err != nil {
			t.Fatal(err)
		}
		fset := &FileSet{Files: []*FileNode{fn}}
		pn := NewPackageNode("main", "main", fset)
		PredefineFileSet(m.Store, pn, fset)
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected a panic")
			}
			pe, ok := r.(*PreprocessError)
			if !ok || !strings.Contains(pe.Unwrap().Error(), "main.gno:4") {
				t.Fatalf("unexpected panic: %v", r)
			}
		}()
		Preprocess(m.Store, pn, fn)
	})
}

func TestPreprocessCollectInitCycle(t *testing.T) {
	msgs := collectErrors(t, `package main

var a = f()

func f() int { return a + g() }

func g() int { return g() } // recursion among functions is not a cycle

var b = 1

func main() { println(a, b) }
`)
	if len(msgs) != 1 || !strings.Contains(msgs[0], "initialization cycle: a refers to f refers to a") {
		t.Fatalf("unexpected: %v", msgs)
	}
	if msgs := collectErrors(t, "package main\n\nvar a = f()\n\nfunc f() int { return 1 }\n\nfunc main() { println(a) }\n"); len(msgs) != 0 {
		t.Fatalf("false positive: %v", msgs)
	}
}

func TestInitCycleOnChain(t *testing.T) {
	// The on-chain path (RunFiles) rejects the cycle at preprocess, before
	// any initializer runs, with the same positioned message.
	m := NewMachine("test", nil)
	defer m.Release()
	fn := m.MustParseFile("a.gno", "package test\n\nvar a = f()\n\nfunc f() int { return a }\n")
	defer func() {
		r := recover()
		pe, ok := r.(*PreprocessError)
		if !ok || !strings.Contains(pe.Error(), "initialization cycle: a refers to f refers to a") {
			t.Fatalf("want an initialization cycle PreprocessError, got %v", r)
		}
	}()
	m.RunFiles(fn)
	t.Fatal("RunFiles accepted a cycle")
}

// collectWith is collectErrors with the accumulator exposed: full messages
// (position kept), an optional in-walk hook, and the panic if one escaped.
func collectWith(t *testing.T, src string, hook func(Node, TransStage)) (msgs []string, panicked any) {
	t.Helper()
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", src)
	if err != nil {
		t.Fatal(err)
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	acc := &preprocessAccumulator{poisoned: map[Name]bool{}, referenced: map[Name]bool{}, testHook: hook}
	func() {
		defer func() { panicked = recover() }()
		predefineFileSet(m.Store, pn, fset, acc)
		for _, f := range fset.Files {
			preprocess(m.Store, pn, f, acc)
		}
	}()
	for _, e := range acc.errs {
		msgs = append(msgs, e.Unwrap().Error())
	}
	return msgs, panicked
}

func TestPreprocessCollectDrops(t *testing.T) {
	t.Run("a Go runtime error is a VM bug, not a diagnostic", func(t *testing.T) {
		hook := func(n Node, stage TransStage) {
			if nx, ok := n.(*NameExpr); ok && nx.Name == "crashme" && stage == TRANS_LEAVE {
				var s []int
				_ = s[1] // runtime.Error, raised inside doRecover like a VM bug would be
			}
		}
		msgs, panicked := collectWith(t, `package main

func main() {
	crashme := 1
	println(crashme)
}
`, hook)
		if panicked == nil || !isRuntimePanic(panicked) {
			t.Fatalf("want the runtime error to escape, got panic %v, errors %v", panicked, msgs)
		}
		if len(msgs) != 0 {
			t.Fatalf("want no diagnostic, got %v", msgs)
		}
	})

	t.Run("a split package-level var is guarded and positioned at its name", func(t *testing.T) {
		msgs, panicked := collectWith(t, `package main

var a, b = undefinedF, 2

func main() { println(a, b) }
`, nil)
		if panicked != nil {
			t.Fatalf("escaped: %v", panicked)
		}
		if len(msgs) != 1 || !strings.Contains(msgs[0], "main.gno:3:5") || !strings.Contains(msgs[0], "undefinedF") {
			t.Fatalf("want one error at the name a (3:5), got %v", msgs)
		}
	})

	t.Run("a suppressed cascade still poisons what it declares", func(t *testing.T) {
		msgs := collectErrors(t, `package main

func main() {
	x := undefinedA
	y := x + 1
	z := y
	println(z)
}
`)
		if len(msgs) != 1 || !strings.Contains(msgs[0], "undefinedA") {
			t.Fatalf("want only the root error, got %v", msgs)
		}
	})

	t.Run("a dropped dependency reports once, at its own declaration", func(t *testing.T) {
		for name, src := range map[string]string{
			"dependency first": `package main

const C = undefinedH + 1
var V = C * 2
var W = V + 1

func main() { println(W) }
`,
			"dependent first": `package main

var W = V + 1
var V = C * 2
const C = undefinedH + 1

func main() { println(W) }
`,
		} {
			msgs, panicked := collectWith(t, src, nil)
			if panicked != nil {
				t.Fatalf("%s: escaped: %v", name, panicked)
			}
			cLine := "main.gno:3:"
			if name == "dependent first" {
				cLine = "main.gno:5:"
			}
			if len(msgs) != 1 || !strings.Contains(msgs[0], cLine) || !strings.Contains(msgs[0], "undefinedH") {
				t.Fatalf("%s: want one error at C (%s), got %v", name, cLine, msgs)
			}
		}
	})

	t.Run("PreprocessCollect on one file node", func(t *testing.T) {
		m := NewMachine("main", nil)
		defer m.Release()
		fn, err := m.ParseFile("main.gno", `package main

var G int = "g"

func main() {
	var n string = 1
	println(G, n)
}
`)
		if err != nil {
			t.Fatal(err)
		}
		fset := &FileSet{Files: []*FileNode{fn}}
		pn := NewPackageNode("main", "main", fset)
		nn, errs := PreprocessCollect(m.Store, pn, fn)
		if nn == nil || len(errs) != 2 {
			t.Fatalf("want the node and 2 errors, got %v, %v", nn != nil, errs)
		}
		if !strings.Contains(errs[0].Unwrap().Error(), "main.gno:3:") || !strings.Contains(errs[1].Unwrap().Error(), "main.gno:6:") {
			t.Fatalf("unexpected positions: %v", errs)
		}
	})
}

func TestParseFileCollect(t *testing.T) {
	m := NewMachine("main", nil)
	defer m.Release()
	fn, perrs := m.ParseFileCollect("main.gno", `package main

func f() int {
	x := 1 +
	return x
}

func g() {
	if {
	}
	var s string = 1
	_ = s
}

func main() {}
`)
	// AllErrors: both errors of line 5 are kept, plus line 9.
	if fn == nil || len(perrs) != 3 {
		t.Fatalf("want a file node and 3 parser errors, got %v, %v", fn != nil, perrs)
	}
	for i, want := range []string{"main.gno:5:2:", "main.gno:5:2:", "main.gno:9:5:"} {
		if !strings.HasPrefix(perrs[i].Error(), want) {
			t.Errorf("parser error %d: got %v, want prefix %s", i, perrs[i], want)
		}
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	var msgs []string
	for _, e := range PreprocessFilesCollect(m.Store, pn, fset) {
		msgs = append(msgs, e.Unwrap().Error())
	}
	// The broken statements drop silently; the checker still reaches g's
	// type error and nothing mentions the placeholder.
	if len(msgs) != 1 || !strings.Contains(msgs[0], "main.gno:11:") || !strings.Contains(msgs[0], "cannot use") {
		t.Fatalf("want one error at line 11, got %v", msgs)
	}
	if fn, perrs := m.ParseFileCollect("main.gno", "func f() {}\n"); fn != nil || len(perrs) == 0 {
		t.Fatalf("no package clause: want nil node and an error, got %v, %v", fn != nil, perrs)
	}
}

func TestParseFileCollectDeclarations(t *testing.T) {
	// A declaration Go2Gno refuses drops with its own error; a blank
	// package name is reported and parsed past. (A missing import path
	// is left to the import rule, exercised by the corpus.)
	m := NewMachine("main", nil)
	defer m.Release()
	fn, perrs := m.ParseFileCollect("main.gno", `package _

type T int

func (x, y T) m() {}

func g() {
	var s string = 1
	_ = s
}
`)
	if fn == nil {
		t.Fatalf("want a file node, got errors %v", perrs)
	}
	var got []string
	for _, e := range perrs {
		got = append(got, e.Error())
	}
	for _, want := range []string{"main.gno:1:10: invalid package name _", "main.gno:5:7: "} {
		found := false
		for _, g := range got {
			if strings.HasPrefix(g, want) {
				found = true
			}
		}
		if !found {
			t.Errorf("missing an error starting %q in %v", want, got)
		}
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("_", "main", fset)
	var msgs []string
	for _, e := range PreprocessFilesCollect(m.Store, pn, fset) {
		msgs = append(msgs, e.Unwrap().Error())
	}
	// g is still checked.
	if len(msgs) < 1 || !strings.Contains(msgs[len(msgs)-1], "main.gno:8:") {
		t.Fatalf("want g's error last, got %v", msgs)
	}
	for _, m := range msgs {
		if strings.Contains(m, "!syntax") {
			t.Errorf("placeholder leaked: %s", m)
		}
	}
}

func TestPreprocessCollectUnused(t *testing.T) {
	msgs := collectErrors(t, `package main

func f() {
	x := 1 // unused
	y := 2
	y = 3 // assignment is not a use
	z := 4
	_ = &z
	var w int
	w++ // a use
	s := struct{ f int }{}
	s.f = 1 // a use of s
	c := 0
	func() { c = 1; _ = c }() // used in a closure
	var k int
	k += 1 // a use
L1:
	for {
		break L1
	}
L2:
	for {
		break
	}
	var i interface{} = 1
	switch t := i.(type) {
	case int:
	}
}

func main() { f() }
`)
	// Source order, as gc reports them; the guard variable in gc's spelling.
	want := []string{"declared and not used: x", "declared and not used: y", "label L2 defined and not used", "t declared and not used"}
	if len(msgs) != len(want) {
		t.Fatalf("want %d errors, got %d: %v", len(want), len(msgs), msgs)
	}
	for i, w := range want {
		if !strings.HasSuffix(msgs[i], w) {
			t.Errorf("error %d: want %q, got %q (all: %v)", i, w, msgs[i], msgs)
		}
	}
}

func TestPreprocessCollectUnusedOrder(t *testing.T) {
	// Reported in source order whatever the block nesting: the inner x
	// before the outer y, the labels where they stand. Checked many times
	// in one process so no map order can reach the text or the first error.
	src := `package main

func main() {
L1:
	for {
		break
	}
	if true {
		x := 1
	}
L2:
	for {
		break
	}
	y := 2
L3:
	for {
		break
	}
}
`
	want := "label L1 defined and not used|declared and not used: x|label L2 defined and not used|declared and not used: y|label L3 defined and not used"
	for i := 0; i < 100; i++ {
		if got := strings.Join(collectErrors(t, src), "|"); got != want {
			t.Fatalf("iteration %d:\n got %s\nwant %s", i, got, want)
		}
	}
	// The on-chain path panics with the first one.
	for i := 0; i < 100; i++ {
		m := NewMachine("main", nil)
		fn, err := m.ParseFile("main.gno", src)
		if err != nil {
			t.Fatal(err)
		}
		fset := &FileSet{Files: []*FileNode{fn}}
		pn := NewPackageNode("main", "main", fset)
		got := func() (s string) {
			defer func() { s = fmt.Sprint(recover()) }()
			PredefineFileSet(m.Store, pn, fset)
			Preprocess(m.Store, pn, fn)
			return "no error"
		}()
		m.Release()
		if !strings.Contains(got, "main.gno:5:2-7:3: label L1 defined and not used") {
			t.Fatalf("iteration %d: first error %q", i, got)
		}
	}
}

func TestPreprocessCollectUnusedFuncLit(t *testing.T) {
	// A package-level function literal has no FuncDecl; its body and its
	// nested literals are checked once.
	msgs := collectErrors(t, `package main

var f = func() {
	q := 1
	g := func() {
		r := 2
	}
	g()
}

var _ = func() int {
L:
	s := 3
	return 0
}()

func main() { f() }
`)
	want := []string{"declared and not used: q", "declared and not used: r", "label L defined and not used", "declared and not used: s"}
	if len(msgs) != len(want) {
		t.Fatalf("want %d errors, got %d: %v", len(want), len(msgs), msgs)
	}
	for i, w := range want {
		if !strings.HasSuffix(msgs[i], w) {
			t.Errorf("error %d: want %q, got %q", i, w, msgs[i])
		}
	}
}

func TestPreprocessCollectInitSignature(t *testing.T) {
	for _, tc := range []struct {
		pkgPath, src, want string
	}{
		{"gno.land/r/test", "func init(cur realm) {}", ""},
		{"gno.land/r/test", "func init(cur int) {}", "func init must have no arguments and no return values"},
		{"gno.land/r/test", "func init() int { return 0 }", "func init must have no arguments and no return values"},
	} {
		m := NewMachine("test", nil)
		fn, err := m.ParseFile("test.gno", "package test\n\n"+tc.src+"\n")
		if err != nil {
			t.Fatal(err)
		}
		fset := &FileSet{Files: []*FileNode{fn}}
		pn := NewPackageNode("test", tc.pkgPath, fset)
		errs := PreprocessFilesCollect(m.Store, pn, fset)
		m.Release()
		var msgs []string
		for _, e := range errs {
			msgs = append(msgs, e.Unwrap().Error())
		}
		got := strings.Join(msgs, "|")
		if tc.want == "" && got != "" || tc.want != "" && !strings.Contains(got, tc.want) {
			t.Errorf("%s: want %q, got %q", tc.src, tc.want, got)
		}
	}
}

func TestPreprocessCollectShortVarDecl(t *testing.T) {
	// The parser's resolver already rejects `a := 1; a := 2` and `b, b :=`;
	// the preprocessor rule covers what resolution leaves: range variables.
	msgs := collectErrors(t, `package main

func f(p int) {
	for i, v := range []int{1} {
		i := i // shadows the range variable: new
		v := v
		_, _ = i, v
	}
	for j := 0; j < 1; j++ {
		j := j // shadows the init variable: new
		_ = j
	}
	if x := 1; x > 0 {
		x := x
		_ = x
	}
	for a, a := range []int{1} { // ERROR repeated
		_ = a
	}
	var e any = 1
	switch t := e.(type) { // the guard lives in each clause's block
	case int:
		t := 3 // ERROR no new variables
		_ = t
	}
}

func main() { f(0) }
`)
	if len(msgs) != 2 || !strings.Contains(msgs[0], "a repeated on left side of :=") ||
		!strings.Contains(msgs[1], "no new variables on left side of :=") {
		t.Fatalf("unexpected: %v", msgs)
	}
}

func TestPreprocessCollectExprs(t *testing.T) {
	msgs := collectErrors(t, `package main

type S []int

func f() {
	var s []int
	_ = s[2:1] // inverted
	var a [1]int
	_ = a[2:] // out of bounds
	var b [10]int
	_ = b[0:10] // high may equal the length
	const c = "0123456789"
	_ = c[10] // out of bounds
	_ = c[9]
	var m1 *map[int]int
	_ = len(m1) // pointer to map
	var pa *[3]int
	_ = len(pa) // pointer to array is fine
	_ = []byte{1: 1, 2: 2, 1: 3} // duplicate index
	_ = struct{ X, Y int }{X: 1, X: 2}  // duplicate field
	_ = map[string]int{"a": 1, "b": 2}
	var sl S
	switch sl {
	case nil: // fine
	case S(nil): // a typed nil is not the nil literal
	}
	var e interface{}
	var fn func()
	_ = e == fn
	_ = fn == nil
}

func main() { f() }
`)
	want := []string{"invalid slice indices: 1 < 2", "index 2 out of bounds [0:2]", "index 10 out of bounds [0:10]", "*map[int]int for built-in len", "duplicate index 1 in array or slice literal", "duplicate field name X in struct literal", "main.S can only be compared to nil", "func() can only be compared to nil"}
	for _, w := range want {
		found := false
		for _, m := range msgs {
			if strings.Contains(m, w) {
				found = true
			}
		}
		if !found {
			t.Errorf("missing %q in %v", w, msgs)
		}
	}
	if len(msgs) != len(want) {
		t.Errorf("want %d errors, got %d: %v", len(want), len(msgs), msgs)
	}
}

func TestPreprocessCollectUnusedExempt(t *testing.T) {
	// A name only a folded constant uses, and a blank label, are not unused.
	src := `package main

func main() {
	var a [10]int
	const ca = len(a)
	_ = ca
_:
}
`
	if msgs := collectErrors(t, src); len(msgs) != 0 {
		t.Fatalf("false positive: %v", msgs)
	}
}

func TestPreprocessCollectMainSignature(t *testing.T) {
	msgs := collectErrors(t, `package main

func main(int) {}
`)
	if len(msgs) != 1 || msgs[0] != "func main must have no arguments and no return values" {
		t.Fatalf("got %v", msgs)
	}
}

func TestPreprocessCollectTypes(t *testing.T) {
	msgs := collectErrors(t, `package main

type I interface{ M() }
type P struct{}

func (*P) M() {}

type Field struct{ M func() }
type Bad struct{ m map[Bad]int } // incomplete key, decided at the decl
type Rec struct{ next *Rec }     // pointer cycle is fine
type Ptr *struct{ Ptr }          // embedded named pointer
type Bare interface{ int }

var _ map[func()]int
var _ map[*Rec]int

func (Bad) A() {}

func f(x int, e interface{}, i I, p *P, v P) {
	switch x {
	case 0, 1:
	case 0: // duplicate
	}
	switch e {
	case 0, int64(0), float32(10), float32(10): // duplicate only for the last
	}
	switch e {
	case interface{}(0), interface{}(0): // conversions are not constants
	}
	switch e.(type) {
	default:
	default:
	}
	switch {
	case true, true: // bools are exempt
	}
	switch e.(type) {
	case 0: // not a type
	}
	switch i.(type) {
	case P: // pointer receiver
	case *P, int:
	}
	switch x.(type) { // not an interface
	}
	_ = i.(P)
	_ = i.(*P)
	_ = e.(P)
	var fi Field
	i = fi // a func field is not a method
	var si []int
	var sf []float64
	copy(si, sf)
	copy(si, "hi")
	copy(x, si)
	var a [-1]int
	_ = a
	var b [...]int
	_ = b
	_ = p
	_ = v
}

func main() {}
`)
	want := []string{
		"invalid map key type main.Bad",
		"embedded field type cannot be a pointer",
		"interface contains embedded non-interface int",
		"invalid map key type func()",
		"duplicate case 0 in expression switch",
		"duplicate case float32(10) in expression switch",
		"multiple defaults in switch",
		"0 is not a type",
		"cannot have dynamic type main.P (method M has pointer receiver)",
		"cannot type switch on non-interface value x",
		"impossible type assertion: main.P does not implement main.I (method M has pointer receiver)",
		"main.Field does not implement main.I (main.Field.M is a field, not a method)",
		"arguments to copy have different element types int and float64",
		"arguments to copy have different element types int and byte",
		"copy expects slice arguments; found int and []int",
		"invalid array length -1",
		"invalid use of [...] array (outside a composite literal)",
	}
	for _, w := range want {
		found := false
		for _, m := range msgs {
			if strings.Contains(m, w) {
				found = true
			}
		}
		if !found {
			t.Errorf("missing %q in %v", w, msgs)
		}
	}
	if len(msgs) != len(want) {
		t.Errorf("want %d errors, got %d: %v", len(want), len(msgs), msgs)
	}
}

// collectErrorsRaw is collectErrors with the position kept.
func collectErrorsRaw(t *testing.T, src string) []string {
	t.Helper()
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", src)
	if err != nil {
		t.Fatal(err)
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	errs := PreprocessFilesCollect(m.Store, pn, fset)
	var msgs []string
	for _, e := range errs {
		msgs = append(msgs, e.Unwrap().Error())
	}
	return msgs
}

func TestPreprocessCollectTypePositions(t *testing.T) {
	// B's key check is skipped while B is a shell and decided when B
	// completes; the error still points at the key spelling in A.
	msgs := collectErrorsRaw(t, `package main
type B struct{ s []int; a A }
type A struct{ m map[B]int }
`)
	if len(msgs) != 1 || !strings.Contains(msgs[0], "main.gno:3:22-23: invalid map key type main.B") {
		t.Fatalf("unexpected: %v", msgs)
	}
	// A typed declared constant prints as T(value).
	msgs = collectErrors(t, `package main
type MyInt int
func f(mi MyInt) {
	switch mi {
	case 1, MyInt(1):
	}
}
`)
	if len(msgs) != 1 || msgs[0] != "duplicate case main.MyInt(1) in expression switch" {
		t.Fatalf("unexpected: %v", msgs)
	}
}

func TestPreprocessCollectTypeSwitchCases(t *testing.T) {
	// Every repeated case is reported, at the case, by type identity:
	// method order and parameter names do not distinguish types.
	msgs := collectErrorsRaw(t, `package main
func f(e interface{}) {
	switch e.(type) {
	case int:
	case int:
	case error:
	case error:
	case interface{ r(); w() }:
	case interface{ w(); r() }:
	case func(int):
	case func(f int):
	case nil, nil:
	}
}
`)
	want := []string{
		"main/main.gno:5:7-10: duplicate case int in type switch",
		"main/main.gno:7:7-12: duplicate case .uverse.error in type switch",
		"main/main.gno:9:7-28: duplicate case interface {r func(); w func()} in type switch",
		"main/main.gno:11:7-18: duplicate case func(int) in type switch",
		"main/main.gno:12:12-15: duplicate case nil in type switch",
	}
	if strings.Join(msgs, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected:\n%s", strings.Join(msgs, "\n"))
	}
}

func TestPreprocessCollectTypeGuardOutsideSwitch(t *testing.T) {
	// x.(type) parses anywhere; outside a type switch header it is an
	// error, not a VM failure, also when parenthesised as a switch tag.
	msgs := collectErrorsRaw(t, `package main
func f(i interface{}) {
	if x, ok := i.(type); ok {
		_ = x
	}
	switch (i.(type)) {
	}
}
`)
	want := []string{
		"main/main.gno:3:14-22: use of .(type) outside type switch",
		"main/main.gno:6:10-18: use of .(type) outside type switch",
	}
	if strings.Join(msgs, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected:\n%s", strings.Join(msgs, "\n"))
	}
}

func TestPreprocessCollectSwitchTag(t *testing.T) {
	// An incomparable tag is reported at the tag, before any clause
	// compares a case to it.
	msgs := collectErrorsRaw(t, `package main
func f() {
	var ar, ar1 [4]func()
	switch ar {
	case ar1:
	}
}
`)
	want := "main/main.gno:4:9-11: cannot switch on ar (variable of type [4]func()) ([4]func() is not comparable)"
	if len(msgs) != 1 || msgs[0] != want {
		t.Fatalf("unexpected:\n%s", strings.Join(msgs, "\n"))
	}
}

func TestPreprocessCollectInterfaceElems(t *testing.T) {
	// A blank method and an embedded non-interface are reported at the
	// element, not at the interface.
	msgs := collectErrorsRaw(t, `package main
type B1 interface {
	M()
	_()
}
type B2 interface {
	int
}
func f() {
	var _ interface{ _() }
}
`)
	want := []string{
		"main/main.gno:4:2-5: methods must have a unique non-blank name",
		"main/main.gno:7:2-5: interface contains embedded non-interface int",
		"main/main.gno:10:19-22: methods must have a unique non-blank name",
	}
	if strings.Join(msgs, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected:\n%s", strings.Join(msgs, "\n"))
	}
}

func TestPreprocessCollectMalformedFloatConstant(t *testing.T) {
	// A float literal past big.Float's exponent range parses as an
	// infinity; gc rejects it as malformed. An underflow is 0, as in gc.
	msgs := collectErrorsRaw(t, `package main
const _ = 1e+1000000000
const _ = 1e-1000000000
const _ = 1e+646456992
`)
	want := "main/main.gno:2:11-24: malformed constant: 1e+1000000000"
	if len(msgs) != 1 || msgs[0] != want {
		t.Fatalf("unexpected:\n%s", strings.Join(msgs, "\n"))
	}
}

func TestPreprocessCollectLiteralElements(t *testing.T) {
	// Every failing element of a composite literal is reported at the
	// element, in a function body and at package level alike; then the
	// literal drops. Index and key errors point at the element too.
	msgs := collectErrorsRaw(t, `package main
type T struct{ f map[string]string }
var m = map[string]int{
	"a": 1,
	1:   1,
	2:   2,
}
func f() {
	_ = map[string]int{
		"a": 1,
		"b", 2,
		"a": 3,
	}
	_ = []string{"x", 20}
	_ = []string{1: "dup", 1: "dup"}
	_ = T{f: {"a": "b"}}
}
`)
	want := []string{
		"main/main.gno:5:2-3: cannot use untyped Bigint as StringKind",
		"main/main.gno:6:2-3: cannot use untyped Bigint as StringKind",
		"main/main.gno:11:3-6: missing key in map literal",
		"main/main.gno:11:8-9: missing key in map literal",
		"main/main.gno:12:3-6: duplicate key \"a\" in map literal",
		"main/main.gno:14:20-22: cannot use untyped Bigint as StringKind",
		"main/main.gno:15:25-26: duplicate index 1 in array or slice literal",
		"main/main.gno:16:11-21: missing type in composite literal",
	}
	if strings.Join(msgs, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected:\n%s", strings.Join(msgs, "\n"))
	}
}

func TestPreprocessCollectInitExprCount(t *testing.T) {
	// An extra initializer is reported at the value, as gc does; a
	// missing one names the variable without its unresolved path.
	msgs := collectErrorsRaw(t, `package main
func main() {
	var c, d = 1, 2, 3
	var e, f, g = 1, 2
	_, _, _, _, _ = c, d, e, f, g
}
`)
	want := []string{
		"main/main.gno:3:19-20: extra init expr 3",
		"main/main.gno:4:6-20: missing init expr for g",
	}
	if strings.Join(msgs, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected:\n%s", strings.Join(msgs, "\n"))
	}
}

func TestPreprocessCollectPointerToInterface(t *testing.T) {
	// A pointer to an interface has no methods: it implements nothing
	// but the empty interface, in an assignment and in an assertion.
	msgs := collectErrorsRaw(t, `package main
type I interface{ M() }
type S struct{}
func (*S) M() {}
func f(p *I, i I) {
	var _ I = p
	var _ interface{} = p
	_ = i.(*I)
}
`)
	want := []string{
		"main/main.gno:6:6-13: *main.I does not implement main.I (type *main.I is pointer to interface, not interface)",
		"main/main.gno:8:6-12: impossible type assertion: *main.I does not implement main.I (type *main.I is pointer to interface, not interface)",
	}
	if strings.Join(msgs, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected:\n%s", strings.Join(msgs, "\n"))
	}
}
