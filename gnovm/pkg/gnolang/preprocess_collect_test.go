package gnolang

import (
	"bytes"
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

func TestPreprocessCollectAssign(t *testing.T) {
	msgs := collectErrors(t, `package main

type S struct {
	x int "foo"
}
type P *S
type Q struct{ a [2]int }

func (S) m() {}

func f(p P, m map[int]S, mq map[int]Q, mp map[int]*S) {
	var s S
	var u struct {
		x int "bar"
	}
	var v struct {
		x int "foo"
	}
	s = u // tags differ
	s = v // tags agree
	s = S(u)
	m[0].x = 1 // in a map
	mp[0].x = 1 // through a pointer is fine
	mq[0].a[1] = 2 // array element in a map
	_ = p.x // field through a named pointer is fine
	p.m() // method through a named pointer is not
	_ = s
}

func main() {}
`)
	want := []string{
		"cannot use struct{x int} as struct{x int}: struct tags differ",
		"cannot assign to struct field m[0].x in map",
		"cannot assign to mq[0].a[1] (neither addressable nor a map index expression)",
		"missing field m in main.P",
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

func TestPreprocessCollectNamedPointerShadow(t *testing.T) {
	// Through a declared pointer type a method is still the shallowest
	// provider, so it shadows the promoted field and the selector fails.
	msgs := collectErrors(t, `package main

type E struct{ f int }
type T struct{ E }

func (*T) f() int { return 0 }

type P *T

func main() {
	var x P
	_ = x.f
}
`)
	if len(msgs) != 1 || !strings.Contains(msgs[0], "missing field f") {
		t.Fatalf("unexpected: %v", msgs)
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

func TestPreprocessCollectAssignTags(t *testing.T) {
	// Tags take part in identity for method types and copy element types.
	msgs := collectErrors(t, `package main

type I interface{ M(struct{ x int "b" }) }
type S struct{}

func (S) M(struct{ x int "a" }) {}

func f() {
	var i I = S{}
	var a []struct{ x int "a" }
	var b []struct{ x int "b" }
	var c []struct{ x int "a" }
	copy(a, b)
	copy(a, c)
	_ = i
}

func main() {}
`)
	want := []string{
		"main.S does not implement main.I (wrong type for method M)",
		"arguments to copy have different element types struct{x int} and struct{x int}: struct tags differ",
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

func TestPreprocessCollectDecls(t *testing.T) {
	msgs := collectErrors(t, `package main

type S struct{ a, b int }
type V func(a ...int, b int) // variadic not last

const (
	A int = 1
	B byte // no value
)

func f3() (float32, int, string) { return 1, 2, "3" }

func f(x interface{}) {
	i, g, s := f3()
	g, h, t := f3() // g changes type
	_, _, _, _, _ = i, g, s, h, t
	var fl float64
	fl, j := 1, 2 // the constant converts to float64
	_, _ = fl, j
	_ = S{7}
	_ = S{7, 8, 9}
	_ = [2]int{1, 2, 3}
	_ = [2]int{5: 0}
	for ; ; k := 1 {
		_ = k
	}
	switch u := x.(type) {
	}
	switch w := x.(type) {
	case int:
		_ = w
	}
}

func main() {}
`)
	want := []string{
		"can only use ... with final parameter in list",
		"missing init expr for B",
		"cannot use float32 as int",
		"too few values in struct literal of type main.S",
		"too many values in struct literal of type main.S",
		"index 2 is out of bounds (>= 2)",
		"index 5 out of bounds [0:2]",
		"cannot declare in post statement of for loop",
		"declared and not used: u",
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

func TestPreprocessCollectRedefine(t *testing.T) {
	// A := target the block already declares is assigned: an untyped
	// constant must be assignable (kind and representability), a typed
	// value assignable; the target keeps its declared type.
	msgs := collectErrors(t, `package main

const c = 5

func f2() (int, bool) { return 1, true }

func f() {
	var s string
	s, j := 1, 2 // not assignable
	var b bool
	b, k := 1, 2 // not assignable
	var u uint8
	u, l := 300, 2 // overflows
	var fl float64
	fl, m := 1+2, 2 // untyped constant expressions convert
	fl, n := -1, 2
	fl, o := 1<<2, 2
	fl, p := c, 2
	var e interface{}
	e, ok := f2() // assignable, e keeps its type
	_, _, _, _, _, _, _, _ = j, k, l, m, n, o, p, ok
	_, _, _, _, _ = s, b, u, fl, e
}

func main() {}
`)
	want := []string{
		"cannot use 1 (untyped int constant) as string value in assignment",
		"cannot use 1 (untyped int constant) as bool value in assignment",
		"bigint overflows target kind",
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

func TestDefineRedeclaredKeepsType(t *testing.T) {
	// On the run path the redeclared targets take the converted values
	// and keep their declared types.
	var buf bytes.Buffer
	m := NewMachineWithOptions(MachineOptions{PkgPath: "main", Output: &buf})
	defer m.Release()
	fn := m.MustParseFile("main.gno", `package main

const c = 5

func f2() (int, bool) { return 1, true }

func main() {
	var fl float64
	fl, j := 1+2, 2
	println(fl, j)
	fl, k := c, 3
	println(fl, k)
	var e interface{}
	e, ok := f2()
	println(e, ok)
}
`)
	m.RunFiles(fn)
	m.RunMain()
	if got, want := buf.String(), "3 2\n5 3\n1 true\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestConstSpecTypeWithoutValues(t *testing.T) {
	// A typed spec without values repeats nothing and does not become the
	// type the next untyped spec inherits.
	m := NewMachine("main", nil)
	defer m.Release()
	fn := m.MustParseFile("main.gno", "package main\n\nconst (\n\tA int = 1\n\tB byte\n\tC\n)\n")
	c := fn.Decls[2].(*ValueDecl)
	if tx, ok := c.Type.(*NameExpr); !ok || tx.Name != "int" || len(c.Values) != 1 {
		t.Fatalf("C: type %v values %v", c.Type, c.Values)
	}
	msgs := collectErrors(t, "package main\n\nconst (\n\tA int = 1\n\tB byte\n\tC\n)\n\nvar _ int = C\n")
	if len(msgs) != 1 || msgs[0] != "missing init expr for B" {
		t.Fatalf("unexpected: %v", msgs)
	}
}

func TestPreprocessCollectStructFieldName(t *testing.T) {
	msgs := collectErrors(t, `package main

type T struct{ x int }

func main() {
	_ = T{1: 2}
	_ = T{"x": 2}
	_ = T{x: 1}
}
`)
	want := []string{"invalid field name 1 in struct literal", "invalid field name \"x\" in struct literal"}
	if len(msgs) != 2 || msgs[0] != want[0] || msgs[1] != want[1] {
		t.Fatalf("got %v, want %v", msgs, want)
	}
}

func TestPreprocessCollectConsts(t *testing.T) {
	msgs := collectErrors(t, `package main

import ""
import "a b"
import "/foo"

const big = 1 << 600 // 601 bits
const sum = 1<<511 + 1<<511 // 513 bits
const prod = (1 << 300) * (1 << 300)
const fine = 1<<511 - 1

func f() {
L:
	{
		break L // a block
		for {
			continue L
		}
	}
M:
	for {
		break M
	}
}

func main() {}
`)
	want := []string{
		"invalid import path (empty string)",
		"invalid import path (invalid character U+0020 ' ')",
		"import path cannot be absolute path: \"/foo\"",
		"constant shift overflow",
		"constant addition overflow",
		"constant multiplication overflow",
		"invalid break label L",
		"invalid continue label L",
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

// locatedErrors returns the located messages ("file:L:C-C2: msg") of a
// collect run.
func locatedErrors(t *testing.T, src string) []string {
	t.Helper()
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", src)
	if err != nil {
		t.Fatal(err)
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	var msgs []string
	for _, e := range PreprocessFilesCollect(m.Store, pn, fset) {
		msgs = append(msgs, e.Unwrap().Error())
	}
	return msgs
}

func TestPreprocessCollectConstShiftPositions(t *testing.T) {
	// A shift with an untyped count folds through a synthesized node;
	// the error must still point at the expression, not at the
	// enclosing file or function.
	msgs := locatedErrors(t, `package main

const a = 1 << 1025

func f() {
	v := 1 << 1025
	_ = v
}

func main() {}
`)
	want := []string{
		"main/main.gno:3:11-20: constant shift overflow",
		"main/main.gno:6:7-16: constant shift overflow",
	}
	if len(msgs) != len(want) {
		t.Fatalf("want %v, got %v", want, msgs)
	}
	for i := range want {
		if msgs[i] != want[i] {
			t.Errorf("error %d: want %q, got %q", i, want[i], msgs[i])
		}
	}
}

func TestPreprocessCollectConstShiftCount(t *testing.T) {
	msgs := collectErrors(t, `package main

const c uint = 2000
const _ = 0 << 2000
const _ = 1 << c
const _ = 1 << 1074 // within the count bound: overflow
const _ = 1e1000 >> 1 // a float is converted first: no operation named

func main() {}
`)
	want := []string{
		"invalid operation: invalid shift count 2000 (untyped int constant)",
		"invalid operation: invalid shift count c (constant 2000 of type uint)",
		"constant shift overflow",
		"constant overflow",
	}
	if len(msgs) != len(want) {
		t.Fatalf("want %v, got %v", want, msgs)
	}
	for i := range want {
		if msgs[i] != want[i] {
			t.Errorf("error %d: want %q, got %q", i, want[i], msgs[i])
		}
	}
}

func TestPreprocessCollectImportPoison(t *testing.T) {
	// A dropped import poisons its default name: a use of it is not a
	// second error.
	msgs := collectErrors(t, `package main

import "/foo"

func main() { foo.Bar() }
`)
	want := []string{"import path cannot be absolute path: \"/foo\""}
	if len(msgs) != 1 || msgs[0] != want[0] {
		t.Fatalf("want %v, got %v", want, msgs)
	}
}
