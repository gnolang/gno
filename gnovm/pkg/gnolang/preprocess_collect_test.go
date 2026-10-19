package gnolang

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
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
		// The copied value of a split part has no span of its own: the
		// error sits at the part, i.e. at the name a.
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

// collectFiles is collectErrors over several files, in order; messages
// keep their position.
func collectFiles(t *testing.T, files ...string) []string {
	t.Helper()
	m := NewMachine("main", nil)
	defer m.Release()
	fset := &FileSet{}
	for i := 0; i < len(files); i += 2 {
		fn, err := m.ParseFile(files[i], files[i+1])
		if err != nil {
			t.Fatal(err)
		}
		fset.Files = append(fset.Files, fn)
	}
	pn := NewPackageNode("main", "main", fset)
	var msgs []string
	for _, e := range PreprocessFilesCollect(m.Store, pn, fset) {
		msgs = append(msgs, e.Unwrap().Error())
	}
	return msgs
}

// wantErrors checks that msgs are exactly the (position prefix, substring)
// pairs given, in order.
func wantErrors(t *testing.T, msgs []string, want ...string) {
	t.Helper()
	if len(msgs) != len(want)/2 {
		t.Fatalf("want %d errors, got %d: %v", len(want)/2, len(msgs), msgs)
	}
	for i, msg := range msgs {
		pos, sub := want[2*i], want[2*i+1]
		if !strings.HasPrefix(msg, "main/"+pos) || !strings.Contains(msg, sub) {
			t.Errorf("error %d: want %q containing %q, got %q", i, pos, sub, msg)
		}
	}
}

// Finer than a dropped declaration: what cannot be typed becomes
// invalidType, already reported, and everything depending on it is
// still checked (go/types' Typ[Invalid]).
func TestPreprocessCollectInvalid(t *testing.T) {
	t.Run("an undefined signature type keeps the function, its body and its callers", func(t *testing.T) {
		msgs, panicked := collectWith(t, `package main

func mine(int b) int {
	return b + 2
}

func _() (int, undefined) { return 0 }

func main() {
	mine()
	r := mine(1) + "s"
	println(r)
}
`, nil)
		if panicked != nil {
			t.Fatalf("escaped: %v", panicked)
		}
		wantErrors(t, msgs,
			"main.gno:3:15", "name not defined: b",
			"main.gno:7:16", "name not defined: undefined",
			"main.gno:4:9", "name b not declared",
			"main.gno:7:29", "expected 2 return values; got 1",
			"main.gno:10:2", "wrong argument count in call to mine",
			"main.gno:11:7", "cannot use untyped",
		)
	})

	t.Run("a declaration that fails keeps its names, typed invalid", func(t *testing.T) {
		msgs, panicked := collectWith(t, `package main

func bar() (T, int) { return 0, 0 }

func G() int { return 1 }

var a, b = G()

func main() {
	var x, y, z int
	x, y, z = bar()
	a, b = G()
	println(x, y, z, a, b)
}
`, nil)
		if panicked != nil {
			t.Fatalf("escaped: %v", panicked)
		}
		wantErrors(t, msgs,
			"main.gno:3:13", "name not defined: T",
			"main.gno:7:5", "assignment mismatch: 2 variable(s) but G",
			"main.gno:11:2", "assignment mismatch: 3 variables but bar",
			"main.gno:12:2", "assignment mismatch: 2 variables but G",
		)
	})

	t.Run("a dropped statement types its names invalid in its own block only", func(t *testing.T) {
		msgs := collectErrors(t, `package main

func a() {
	x := "s" + 1
	println(x)
}

func b() { println(x) }
`)
		if len(msgs) != 2 || !strings.Contains(msgs[1], "name x not declared") {
			t.Fatalf("want the root error and b's genuine one, got %v", msgs)
		}
	})

	t.Run("a dropped type keeps its identity with an invalid base", func(t *testing.T) {
		msgs := collectErrors(t, `package main

type T struct{ X Undef }

func (t T) get() int { return t.X }

var v = T{X: 1}

func main() {
	var s string = 1
	println(v, s)
}
`)
		if len(msgs) != 2 || !strings.Contains(msgs[0], "Undef") || !strings.Contains(msgs[1], "cannot use") {
			t.Fatalf("want Undef and the string error only, got %v", msgs)
		}
	})

	t.Run("a dropped import silences its uses", func(t *testing.T) {
		msgs := collectErrors(t, `package main

import pkg "nowhere/pkg"

func main() {
	pkg.F(1)
	x := pkg.V + 1
	println(x)
}
`)
		if len(msgs) != 1 || !strings.Contains(msgs[0], "unknown import path") {
			t.Fatalf("want the import error only, got %v", msgs)
		}
	})
	t.Run("an undefined package qualifier, a literal key and an invalid field keep going", func(t *testing.T) {
		msgs, panicked := collectWith(t, `package main

type T x.T

type S struct{ f x.T }

type A struct{ a A }

var _ = []int{a: true, true}

func f(t T, s S, p *A) {
	s.f.M()
	s.g = 1
	p.bar()
	println(t)
}

func main() {}
`, nil)
		if panicked != nil {
			t.Fatalf("escaped: %v", panicked)
		}
		wantErrors(t, msgs,
			"main.gno:3:8", "name not defined: x",
			"main.gno:5:18", "name not defined: x",
			"main.gno:7:6", "invalid recursive type: A -> A",
			"main.gno:9:15", "name not defined: a",
			"main.gno:9:24", "cannot use untyped bool",
			"main.gno:13:2", "missing field g in main.S",
		)
	})
}

func TestPreprocessCollectClauses(t *testing.T) {
	msgs, panicked := collectWith(t, `package main

type Stringer interface{ String() string }

func main() {
	var e interface{}
	switch e := e.(type) {
	case 0:
		e.x = 1
	case T:
		e.M()
	case Stringer:
		e.M()
	}
	switch 1 {
	case "s":
	case 2:
		var s string = 1
		_ = s
	}
}
`, nil)
	if panicked != nil {
		t.Fatalf("escaped: %v", panicked)
	}
	// A failing clause drops alone: its body is silenced (go/types), the
	// other clauses are checked.
	wantErrors(t, msgs,
		"main.gno:8:2", "0 is not a type",
		"main.gno:10:7", "name T not declared",
		"main.gno:13:3", "missing field M in main.Stringer",
		"main.gno:16:2", "cannot use",
		"main.gno:18:7", "cannot use",
	)
}

func TestPreprocessCollectElements(t *testing.T) {
	t.Run("each failing element of a package-level literal", func(t *testing.T) {
		msgs, panicked := collectWith(t, `package main

const max32 = 3.40282346638528859811704183484516925440e+38

var x = []interface{}{
	float32(max32),
	float32(max32 * 2),
	float32(-max32 * 2),
	"ok",
	float64(1e400),
	float64(-1e400),
}

func main() { println(x) }
`, nil)
		if panicked != nil {
			t.Fatalf("escaped: %v", panicked)
		}
		wantErrors(t, msgs,
			"main.gno:7:2", "cannot convert",
			"main.gno:8:2", "cannot convert",
			"main.gno:10:2", "cannot convert",
			"main.gno:11:2", "cannot convert",
		)
	})

	t.Run("bad keys, also on the dependency path", func(t *testing.T) {
		msgs, panicked := collectWith(t, `package main

type T struct {
	toInt    map[string]int
	toString map[int]string
}

var u = t

var t = T{
	foo.toInt:    make(map[string]int),
	bar.toString: make(map[int]string),
}

func main() { println(u) }
`, nil)
		if panicked != nil {
			t.Fatalf("escaped: %v", panicked)
		}
		wantErrors(t, msgs,
			"main.gno:11:2", "invalid field name foo.toInt in struct literal",
			"main.gno:12:2", "invalid field name bar.toString in struct literal",
		)
	})

	t.Run("in a function body", func(t *testing.T) {
		msgs, panicked := collectWith(t, `package main

type P struct{ A, B int }

func main() {
	x := []int{1, "s", 2, 3.5}
	p := P{1, "s"}
	m := map[string]int{"a": 1, 2: 2, "c": "s"}
	println(x, p, m)
}
`, nil)
		if panicked != nil {
			t.Fatalf("escaped: %v", panicked)
		}
		wantErrors(t, msgs,
			"main.gno:6:16", "cannot use",
			"main.gno:6:24", "cannot convert",
			"main.gno:7:12", "cannot use",
			"main.gno:8:30", "cannot use",
			"main.gno:8:41", "cannot use",
		)
	})
}

func TestPreprocessCollectDuplicates(t *testing.T) {
	t.Run("methods, which the parser does not see", func(t *testing.T) {
		msgs := collectFiles(t, "main.gno", `package main

type T int

func (T) m()  {}
func (T) m()  {}
func (*T) p() {}
func (*T) p() {}

func main() {}
`)
		wantErrors(t, msgs,
			"main.gno:6:1", "T.m redeclared in this block; other declaration at main/main.gno:5:1",
			"main.gno:8:1", "T.p redeclared in this block; other declaration at main/main.gno:7:1",
		)
	})

	t.Run("across files, the resolving declaration survives", func(t *testing.T) {
		msgs := collectFiles(t,
			"a.gno", `package main

var v int

func f() int { return v }
`,
			"b.gno", `package main

var v string

func f() string { return v }

func main() { println(f()) }
`)
		wantErrors(t, msgs,
			"a.gno:3:5", "v redeclared in this block; other declaration at main/b.gno:3:5",
			"a.gno:5:1", "f redeclared in this block; other declaration at main/b.gno:5:1",
		)
	})
}

func TestPreprocessCollectDeterministic(t *testing.T) {
	src := `package main

type T struct{ X Undef }

func (t T) get() int { return t.X }

func mine(int b) int { return b + 2 }

var a, b = mine()

func main() {
	var e interface{}
	switch e := e.(type) {
	case 0:
		e.x = 1
	case T:
		e.get()
	}
	x := []int{1, "s", undefinedV}
	y := x + mine(1, 2)
	println(a, b, y, Undef2)
}
`
	first := collectErrors(t, src)
	if len(first) < 6 {
		t.Fatalf("want a multi-error file, got %v", first)
	}
	for i := 1; i < 50; i++ {
		if got := collectErrors(t, src); !slices.Equal(got, first) {
			t.Fatalf("run %d differs:\n%v\nvs\n%v", i, got, first)
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

func TestPreprocessCollectLenCapEverySite(t *testing.T) {
	// A bad len or cap operand does not drop its statement: the call is
	// still an int, so every site in one expression is reported, as gc does.
	msgs := collectErrors(t, `package main

func main() {
	var m *map[int]int
	var s *[]int
	x := len(m) +
		cap(s)
	_ = x
}
`)
	want := []string{"invalid argument: *map[int]int for built-in len", "invalid argument: *[]int for built-in cap"}
	if len(msgs) != 2 || msgs[0] != want[0] || msgs[1] != want[1] {
		t.Fatalf("want %v, got %v", want, msgs)
	}
}

func TestPreprocessCollectIndexBase(t *testing.T) {
	// Only a pointer to an array is dereferenced by an index expression.
	msgs := collectErrors(t, `package main

func main() {
	var ps *string
	var pa *[3]int
	var pm *map[int]int
	_ = ps[0]
	_ = pa[0]
	_ = pm[0]
	_ = (*ps)[0]
}
`)
	want := []string{"invalid operation: cannot index ps (variable of type *string)", "invalid operation: cannot index pm (variable of type *map[int]int)"}
	if len(msgs) != 2 || msgs[0] != want[0] || msgs[1] != want[1] {
		t.Fatalf("want %v, got %v", want, msgs)
	}
}

func TestPreprocessCollectConstLenCap(t *testing.T) {
	// len or cap of an array is constant only when its operand holds no
	// call; a conversion or another constant len is not a call.
	msgs := collectErrors(t, `package main

var s [][10]int
var a [3][4]int

func f() [5]int { return [5]int{} }

func main() {
	const m = len(s[len(s)-1])
	const n = cap(f())
	const p = len(a[len(a)-1])
	const q = len([4]int(a[0]))
	const r = cap(a)
	_, _, _, _, _ = m, n, p, q, r
}
`)
	// The operand prints in its preprocessed form, as the neighbouring
	// "is not constant" messages do.
	if len(msgs) != 2 || !strings.Contains(msgs[0], "(s<") || !strings.Contains(msgs[1], "(f<") ||
		!strings.HasSuffix(msgs[0], ") (value of type int) is not constant") || !strings.HasSuffix(msgs[1], ") (value of type int) is not constant") {
		t.Fatalf("got %v", msgs)
	}
}

func TestPreprocessCollectShiftInInterface(t *testing.T) {
	// An untyped expression passed to an interface takes its default type
	// as a whole, so the shift's operand in `1.0 + 1<<s` is a float64.
	msgs := collectErrors(t, `package main

func main() {
	var s uint
	var i any = 1 + 1<<s
	println(1 + 1<<s, i)
	println(1.0 + 1<<s)
}
`)
	if len(msgs) != 1 || !strings.Contains(msgs[0], "operator << not defined on") {
		t.Fatalf("got %v", msgs)
	}
}

func TestPreprocessCollectEmbeddedIdentity(t *testing.T) {
	// struct{T} and struct{T T} share a TypeID but are not identical.
	msgs := collectErrors(t, `package main

type T int

func main() {
	var x struct{ T }
	var y struct{ T T }
	var z struct{ T }
	x = y
	x = z
	var xs []struct{ T }
	var ys []struct{ T T }
	xs = ys
	_, _ = x, xs
}
`)
	want := []string{
		"cannot use struct{T main.T} as struct{T main.T} (field T is embedded in only one)",
		"cannot use []struct{T main.T} as []struct{T main.T}",
	}
	if len(msgs) != 2 || msgs[0] != want[0] || msgs[1] != want[1] {
		t.Fatalf("want %v, got %v", want, msgs)
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

// depStore is a bare store that resolves "dep", "dep2" and "dep3" to
// one-file packages.
func depStore() Store {
	store := NewStore(nil, nil, nil)
	store.SetPackageGetter(func(pkgPath string, st Store) (*PackageNode, *PackageValue) {
		switch pkgPath {
		case "dep", "dep2", "dep3":
		default:
			return nil, nil
		}
		m := NewMachineWithOptions(MachineOptions{PkgPath: pkgPath, Store: st, SkipPackage: true})
		defer m.Release()
		return m.RunMemPackage(&std.MemPackage{Type: MPStdlibProd, Name: pkgPath, Path: pkgPath, Files: []*std.MemFile{
			{Name: "dep.gno", Body: "package " + pkgPath + "\n\nconst X = 1\n\ntype T struct{}\n\nfunc F() int { return X }\n"},
		}}, false)
	})
	return store
}

// collectPkgErrors preprocesses srcs (one file each: a.gno, b.gno, ...) as
// one package in accumulate mode over depStore and returns the messages
// with their position kept.
func collectPkgErrors(t *testing.T, srcs ...string) []string {
	t.Helper()
	m := NewMachine("main", depStore())
	defer m.Release()
	var files []*FileNode
	for i, src := range srcs {
		fn, err := m.ParseFile(fmt.Sprintf("%c.gno", 'a'+i), src)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, fn)
	}
	fset := &FileSet{Files: files}
	pn := NewPackageNode("main", "main", fset)
	var msgs []string
	for _, e := range PreprocessFilesCollect(m.Store, pn, fset) {
		msgs = append(msgs, e.Unwrap().Error())
	}
	return msgs
}

func TestPreprocessCollectUnusedImport(t *testing.T) {
	t.Run("unused import, at its spec", func(t *testing.T) {
		msgs := collectPkgErrors(t, `package main

import (
	"dep"
	"dep2"
)

func main() { println(dep.X) }
`)
		if len(msgs) != 1 || msgs[0] != `main/a.gno:5:2-8: "dep2" imported and not used` {
			t.Fatalf("got %v", msgs)
		}
	})

	t.Run("imports are per file; a use in another file does not count", func(t *testing.T) {
		msgs := collectPkgErrors(t, `package main

import d "dep"

func main() { println(g()) }
`, `package main

import "dep"

func g() int { return dep.F() }
`)
		if len(msgs) != 1 || msgs[0] != `main/a.gno:3:8-15: "dep" imported as d and not used` {
			t.Fatalf("got %v", msgs)
		}
	})

	t.Run("a type, a constant and a value reference are uses", func(t *testing.T) {
		msgs := collectPkgErrors(t, `package main

import (
	"dep"
	d2 "dep2"
	"dep3"
)

type S struct{ f dep.T }

const c = d2.X

var _ = dep3.F

func main() { _ = S{} }
`)
		if len(msgs) != 0 {
			t.Fatalf("false positive: %v", msgs)
		}
	})

	t.Run("blank imports are exempt", func(t *testing.T) {
		if msgs := collectPkgErrors(t, "package main\n\nimport _ \"dep\"\n\nfunc main() {}\n"); len(msgs) != 0 {
			t.Fatalf("false positive: %v", msgs)
		}
	})

	t.Run("a name a dropped statement mentioned is not reported", func(t *testing.T) {
		msgs := collectPkgErrors(t, "package main\n\nimport \"dep\"\n\nfunc main() { dep.Nope() }\n")
		for _, msg := range msgs {
			if strings.Contains(msg, "imported and not used") {
				t.Fatalf("reported a dropped statement's import: %v", msgs)
			}
		}
	})

	t.Run("source order, files then specs", func(t *testing.T) {
		msgs := collectPkgErrors(t, "package main\n\nimport (\n\t\"dep2\"\n\t\"dep\"\n)\n\nfunc main() {}\n", "package main\n\nimport \"dep3\"\n")
		want := []string{
			`main/a.gno:4:2-8: "dep2" imported and not used`,
			`main/a.gno:5:2-7: "dep" imported and not used`,
			`main/b.gno:3:8-14: "dep3" imported and not used`,
		}
		if fmt.Sprint(msgs) != fmt.Sprint(want) {
			t.Fatalf("got %v\nwant %v", msgs, want)
		}
	})
}

func TestUnusedImportOnChain(t *testing.T) {
	// The on-chain path (RunFiles) rejects the package once its files are
	// preprocessed, with the first unused import's positioned message.
	m := NewMachine("test", depStore())
	defer m.Release()
	fn := m.MustParseFile("a.gno", "package test\n\nimport \"dep\"\n\nfunc f() int { return 1 }\n")
	defer func() {
		r := recover()
		pe, ok := r.(*PreprocessError)
		if !ok || !strings.Contains(pe.Error(), `test/a.gno:3:8-13: "dep" imported and not used`) {
			t.Fatalf("want an unused-import PreprocessError, got %v", r)
		}
	}()
	m.RunFiles(fn)
	t.Fatal("RunFiles accepted an unused import")
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
	if len(msgs) != 1 || msgs[0] != "duplicate case MyInt(1) in expression switch" {
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

func TestPreprocessCollectTypeSwitchGuardShadow(t *testing.T) {
	// The guard t shadows the init t in each clause; a use of t in a
	// clause is the guard's, so the init t is declared and not used.
	msgs := collectErrorsRaw(t, `package main
func f(x interface{}) {
	switch t := 0; t := x.(type) {
	case int:
		_ = t
	}
}
`)
	want := []string{
		"main/main.gno:3:9-10: declared and not used: t",
	}
	if strings.Join(msgs, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected:\n%s", strings.Join(msgs, "\n"))
	}
}

func TestPreprocessCollectUndefinedDependency(t *testing.T) {
	// An undefined name in a declaration's type is reported where it is
	// spelled, not at the declaration.
	msgs := collectErrorsRaw(t, `package main
type S struct {
	err foo.Bar
	Num int
}
type T struct {
	x T1
}
`)
	want := []string{
		"main/main.gno:3:6-13: name not defined: foo",
		"main/main.gno:7:4-6: name not defined: T1",
	}
	if strings.Join(msgs, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected:\n%s", strings.Join(msgs, "\n"))
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

func TestPreprocessCollectStructLiteralKeys(t *testing.T) {
	// Every key that is not a field name is reported, before it is
	// resolved as an expression; keys of other literals are untouched.
	msgs := collectErrors(t, `package main

type T struct{ toInt, toString int }

var foo, bar struct{ toInt, toString int }

func main() {
	t := T{
		foo.toInt:    1,
		bar.toString: 2,
	}
	m := map[string]int{"a": 1, "b": 2}
	s := []int{0: 1, 3: 2}
	_, _, _ = t, m, s
}
`)
	want := []string{
		"invalid field name foo.toInt in struct literal",
		"invalid field name bar.toString in struct literal",
	}
	if len(msgs) != 2 || msgs[0] != want[0] || msgs[1] != want[1] {
		t.Fatalf("got %v, want %v", msgs, want)
	}
}

func TestPreprocessCollectUnknownField(t *testing.T) {
	msgs := collectErrors(t, `package main

type T struct{ x int }

func main() {
	_ = T{y: 1}
	_ = struct{}{foo: 0}
}
`)
	want := []string{
		"unknown field y in struct literal of type main.T",
		"unknown field foo in struct literal of type struct{}",
	}
	if len(msgs) != 2 || msgs[0] != want[0] || msgs[1] != want[1] {
		t.Fatalf("got %v, want %v", msgs, want)
	}
}

func TestPreprocessCollectBlankSelector(t *testing.T) {
	msgs := collectErrors(t, `package main

var t struct{ _ int }

func main() {
	_ = t._
}
`)
	want := "t._ undefined (type struct{_ int} has no field or method _)"
	if len(msgs) != 1 || msgs[0] != want {
		t.Fatalf("got %v, want [%s]", msgs, want)
	}
}

func TestPreprocessCollectReceiverType(t *testing.T) {
	msgs := collectErrors(t, `package main

type A = int

func (x int) m()      {}
func (p *int) n()     {}
func (a A) o()        {}
func (s struct{}) p() {}

func main() {}
`)
	want := []string{
		"cannot define new methods on non-local type int",
		"cannot define new methods on non-local type int",
		"cannot define new methods on non-local type int",
		"invalid receiver type struct{}",
	}
	if len(msgs) != len(want) {
		t.Fatalf("got %v, want %v", msgs, want)
	}
	for i := range want {
		if msgs[i] != want[i] {
			t.Fatalf("got %v, want %v", msgs, want)
		}
	}
}

func TestPreprocessCollectTypeDeclNotAType(t *testing.T) {
	// A type declaration naming a constant, variable or function fails
	// at the name, also when that value's declaration needs the type.
	msgs := collectErrors(t, `package main

const C = 0

type T1 C

var V = 1

type T2 V

func F() {}

type T3 F

type T4 K

const K = T4(0)

func main() {
	const c = 0
	type L c
}
`)
	want := []string{
		"C is not a type",
		"V is not a type",
		"F is not a type",
		"K is not a type",
		"c is not a type",
	}
	if len(msgs) != len(want) {
		t.Fatalf("got %v, want %v", msgs, want)
	}
	for i := range want {
		if msgs[i] != want[i] {
			t.Fatalf("got %v, want %v", msgs, want)
		}
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

func TestPreprocessCollectConstFloatOverflow(t *testing.T) {
	// A float constant past big.Float's exponent range is an infinity:
	// rejected where it arises, so no later fold meets it (Inf - Inf).
	msgs := collectErrors(t, `package main

const a = 1e+500000000
const b = a * a
const c = b - b
const d = a / 1e-500000000
const e = 1e+700000000
const f = -a

func main() {}
`)
	want := []string{
		"constant result is not representable",
		"constant result is not representable",
		"malformed constant: 1e+700000000",
	}
	if len(msgs) != len(want) {
		t.Fatalf("got %v, want %v", msgs, want)
	}
	for i := range want {
		if msgs[i] != want[i] {
			t.Fatalf("got %v, want %v", msgs, want)
		}
	}
}

func TestPreprocessCollectExprStmts(t *testing.T) {
	msgs := collectErrors(t, `package main

var notmain func()

var lit = func() {
	a := 1 // unused in a package-level literal
}

func g(n int) int { return n }

func main() {
	len("foo")
	1 + 2
	int(3)
	g(1) // a call is a statement
	var x = &main
	_ = x
	main = notmain
	println("a", "b")
	var s []string
	println(s...)
	_ = append(nil, 1)
	_ = []int{g(1): 2}
	const c = "abc"
	_ = map[int]int{len(c) + 0: 0, 3: 1} // len of a constant is constant
}
`)
	want := []string{
		"declared and not used: a",
		"len(\"foo\") (value of type int) is not used",
		"1 + 2 (value of type",
		"int(3) (value of type int) is not used",
		"cannot take address of main (value of type func())",
		"cannot assign to main (neither addressable nor a map index expression)",
		"invalid use of ... with built-in println",
		"first argument to append must be a typed slice; have untyped nil",
		"index g(1) must be integer constant",
		"duplicate key 3 in map literal",
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

func TestPreprocessCollectExprStrings(t *testing.T) {
	// One rule, one message: a non-constant literal index is reported
	// the same way whatever its expression; messages print expressions
	// as go/types does.
	msgs := collectErrors(t, `package main

func main() {
	i := 1
	_ = []int{i: 2}
	_ = [2]int{i + 1: 2}
	a := []int{1}
	append(a, 2)
	make([]int, 1)
	make(map[string][]int)
	func() {}
}
`)
	want := []string{
		"index i must be integer constant",
		"index i + 1 must be integer constant",
		"append(a, 2) (value of type []int) is not used",
		"make([]int, 1) (value of type []int) is not used",
		"make(map[string][]int) (value of type map[string][]int) is not used",
		"(func() literal) (value of type func()) is not used",
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

func TestPreprocessCollectFuncDeclName(t *testing.T) {
	// A parameter or local that shadows a declared function is
	// addressable and assignable, also through a faux block.
	msgs := collectErrors(t, `package main

func g() {}

func h(g func()) {
	_ = &g
	if x := 1; x > 0 {
		g = nil
	}
	_ = &h
}

func main() { h(g) }
`)
	if len(msgs) != 1 || !strings.Contains(msgs[0], "cannot take address of h") {
		t.Fatalf("want one error on &h, got %v", msgs)
	}
}

func TestLenConstStringFolds(t *testing.T) {
	// len of a constant string is a constant: folded at preprocess, so
	// it costs nothing at run time (a gas change pinned by the ADR).
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", "package main\n\nconst c = \"abc\"\n\nvar n = len(c)\n\nfunc main() {}\n")
	if err != nil {
		t.Fatal(err)
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	if errs := PreprocessFilesCollect(m.Store, pn, fset); len(errs) > 0 {
		t.Fatal(errs)
	}
	for _, d := range fn.Decls {
		vd, ok := d.(*ValueDecl)
		if !ok || vd.NameExprs[0].Name != "n" {
			continue
		}
		cx, ok := vd.Values[0].(*ConstExpr)
		if !ok {
			t.Fatalf("len(c) not folded: %T", vd.Values[0])
		}
		if cx.T != IntType || cx.GetInt() != 3 {
			t.Fatalf("len(c) folded to %v", cx.TypedValue)
		}
		return
	}
	t.Fatal("declaration n not found")
}

func TestPreprocessCollectAppendArgs(t *testing.T) {
	msgs := collectErrorsRaw(t, `package main

func main() {
	s := make([]int, 8)
	_ = append()
	_ = append(s...)
	_ = append(s, 2, s...)
	_ = append(s, make([]int, 0))
	var q int8 = 1
	_ = append(s, q)
	var b []byte
	b = append(b, "x"...)
	var e []error
	e = append(e, nil)
	_ = append(s, 1, 2)
	_ = append(s, s...)
}
`)
	want := []string{
		"main.gno:5:6-14: invalid operation: not enough arguments for append() (expected 1, found 0)",
		"main.gno:6:6-18: not enough arguments in call to append\n\thave ([]int...)\n\twant ([]int, ...int)",
		"main.gno:7:19-20: too many arguments in call to append\n\thave ([]int, number, []int...)\n\twant ([]int, ...int)",
		"main.gno:8:16-30: cannot use make([]int, 0) (value of type []int) as int value in argument to append",
		"main.gno:10:16-17: cannot use q (variable of type int8) as int value in argument to append",
	}
	if len(msgs) != len(want) {
		t.Fatalf("want %d errors, got %d: %q", len(want), len(msgs), msgs)
	}
	for i, w := range want {
		if !strings.HasSuffix(msgs[i], w) {
			t.Errorf("error %d: want suffix %q, got %q", i, w, msgs[i])
		}
	}
}

func TestPreprocessCollectMakeNegative(t *testing.T) {
	msgs := collectErrorsRaw(t, `package main

func main() {
	_ = make([]int, -1)
	const i = -1
	_ = make([]int, 0, i)
	_ = make(map[int]int, int8(-1))
	_ = make([]int, -1.0)
	_ = make([]int, -1.2)
	_ = make([]int, 2)
}
`)
	want := []string{
		"main.gno:4:18-20: invalid argument: index -1 (constant of type int) must not be negative",
		"main.gno:6:21-22: invalid argument: index i (constant -1 of type int) must not be negative",
		"main.gno:7:24-32: invalid argument: index int8(-1) (constant -1 of type int8) must not be negative",
		"main.gno:8:18-22: invalid argument: index -1.0 (constant -1 of type int) must not be negative",
		"main.gno:9:18-22: invalid argument: index -1.2 (untyped float constant) must not be negative",
	}
	if len(msgs) != len(want) {
		t.Fatalf("want %d errors, got %d: %q", len(want), len(msgs), msgs)
	}
	for i, w := range want {
		if !strings.HasSuffix(msgs[i], w) {
			t.Errorf("error %d: want suffix %q, got %q", i, w, msgs[i])
		}
	}
}

func TestPreprocessCollectSelectorUndefined(t *testing.T) {
	msgs := collectErrors(t, `package main

const c = 3

var v int

func main() {
	_ = c.String()
	_ = v.String()
	_ = []int{}.Len
	_ = int.String
	var f func()
	_ = f.X
}
`)
	want := []string{
		"c.String undefined (type untyped int has no field or method String)",
		"v.String undefined (type int has no field or method String)",
		"[]int{}.Len undefined (type []int has no field or method Len)",
		"int.String undefined (type int has no method String)",
		"f.X undefined (type func() has no field or method X)",
	}
	if len(msgs) != len(want) {
		t.Fatalf("want %d errors, got %d: %q", len(want), len(msgs), msgs)
	}
	for i, w := range want {
		if msgs[i] != w {
			t.Errorf("error %d: want %q, got %q", i, w, msgs[i])
		}
	}
}

// collectErrorsFiles is collectErrors over several files of one package,
// with an empty package "foo" in the store for them to import.
func collectErrorsFiles(t *testing.T, srcs ...string) []string {
	t.Helper()
	m := NewMachine("main", nil)
	defer m.Release()
	fpn := NewPackageNode("foo", "foo", &FileSet{})
	m.Store.SetBlockNode(fpn)
	m.Store.SetCachePackage(fpn.NewPackage(m.Alloc))
	fset := &FileSet{}
	for i, src := range srcs {
		fn, err := m.ParseFile(fmt.Sprintf("f%d.gno", i), src)
		if err != nil {
			t.Fatal(err)
		}
		fset.Files = append(fset.Files, fn)
	}
	pn := NewPackageNode("main", "main", fset)
	var msgs []string
	for _, e := range PreprocessFilesCollect(m.Store, pn, fset) {
		msgs = append(msgs, e.Unwrap().Error())
	}
	return msgs
}

func TestPreprocessCollectImportName(t *testing.T) {
	// Spans are the declarations' name spans as Go2Gno sets them; the
	// line is what matters.
	check := func(t *testing.T, msgs []string, want map[string]string) {
		t.Helper()
		// The clashing imports are unused too (gc reports that as well);
		// this test is about the clash.
		msgs = slices.DeleteFunc(slices.Clone(msgs), func(m string) bool { return strings.Contains(m, "imported") })
		if len(msgs) != len(want) {
			t.Fatalf("want %d errors, got %d: %q", len(want), len(msgs), msgs)
		}
		for _, m := range msgs {
			loc, msg, _ := strings.Cut(m, ": ")
			line := strings.Split(loc, ":")[1]
			if want[line] != msg {
				t.Errorf("line %s: want %q, got %q", line, want[line], msg)
			}
		}
	}
	t.Run("declaration after import", func(t *testing.T) {
		msgs := collectErrorsFiles(t, `package main

import "foo"
import bar "foo"
import baz "foo"
import qux "foo"

var foo int

type bar struct{}

func baz() {}

const qux = 1

func main() {}
`)
		check(t, msgs, map[string]string{
			"8":  `foo already declared through import of package foo ("foo")`,
			"10": `bar already declared through import of package foo ("foo")`,
			"12": `baz already declared through import of package foo ("foo")`,
			"14": `qux already declared through import of package foo ("foo")`,
		})
	})
	t.Run("import after a declaration in another file", func(t *testing.T) {
		msgs := collectErrorsFiles(t, "package main\n\nvar foo int\n\nfunc main() {}\n",
			"package main\n\nimport \"foo\"\n")
		check(t, msgs, map[string]string{"3": `foo already declared through import of package foo ("foo")`})
	})
	t.Run("a name declared in another file is not a clash", func(t *testing.T) {
		msgs := collectErrorsFiles(t, "package main\n\nimport \"foo\"\n\nfunc main() {}\n",
			"package main\n\nvar bar int\n")
		if len(msgs) != 1 || !strings.Contains(msgs[0], `"foo" imported and not used`) {
			t.Fatalf("want only the unused import, got %q", msgs)
		}
	})
}
