package gnolang

import (
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
		"main.gno:8:7", "0 is not a type",
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
			"main.gno:11:2", "name foo not declared",
			"main.gno:12:2", "name bar not declared",
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

// deployAfterCollect runs the ordinary (non-accumulate) path over src,
// optionally running a collect pass over the very same parsed nodes first,
// and reports what the ordinary path did. A fresh PackageNode each time, so
// only the AST is shared.
func deployAfterCollect(t *testing.T, src string, collectFirst bool) (res string) {
	t.Helper()
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", src)
	if err != nil {
		t.Fatal(err)
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	if collectFirst {
		PreprocessFilesCollect(m.Store, NewPackageNode("main", "main", fset), fset)
	}
	pn := NewPackageNode("main", "main", fset)
	defer func() {
		if r := recover(); r != nil {
			res = fmt.Sprint(r)
		}
	}()
	PredefineFileSet(m.Store, pn, fset)
	for _, f := range fset.Files {
		Preprocess(m.Store, pn, f)
	}
	return ""
}

// TestPreprocessCollectIsNotReusable pins that nodes accumulate mode has
// seen cannot be handed to the ordinary path. It drops a failing
// declaration and leaves it ATTR_PREPROCESSED, so without the guard the
// ordinary path skips the very code it must reject and ACCEPTS an invalid
// package — a checking pass would decide what the deploy path admits.
func TestPreprocessCollectIsNotReusable(t *testing.T) {
	for _, tc := range []struct{ name, src, wantAlone string }{
		{
			"file-level declaration",
			"package main\n\nvar x int = \"nope\"\n\nfunc main() { println(x) }\n",
			"cannot use untyped string as IntKind",
		},
		{
			"body statement",
			"package main\n\nfunc main() {\n\tvar s string = \"a\"\n\tvar i int = s\n\tprintln(i)\n}\n",
			"cannot use string as int",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			alone := deployAfterCollect(t, tc.src, false)
			if !strings.Contains(alone, tc.wantAlone) {
				t.Fatalf("the ordinary path must reject this on its own; got %q", alone)
			}
			after := deployAfterCollect(t, tc.src, true)
			if after == "" {
				t.Fatalf("a collect pass made the ordinary path ACCEPT an invalid package")
			}
			if !strings.Contains(after, "must not be preprocessed again") {
				t.Errorf("want the reuse guard, got %q", after)
			}
		})
	}
}

func TestPreprocessCollectConditionFollowOns(t *testing.T) {
	msgs, panicked := collectWith(t, `package main

func main() {
	did := 1
	var b bool
	if !did && b {
		_ = undefined1
	}
	for !did {
		_ = undefined2
	}
	switch !did {
	case true:
		_ = undefined3
	}
	var x interface{}
	switch t := x.(type) {
	case 0:
		t.x = 1
		x.x = 1
	}
}
`, nil)
	if panicked != nil {
		t.Fatalf("escaped: %v", panicked)
	}
	// A failed condition, tag or case leaves its body checked (did counts
	// as used); the clause's variable is invalid, so t.x is silent.
	wantErrors(t, msgs,
		"main.gno:6:5", "operator ! not defined on: IntKind",
		"main.gno:7:7", "undefined1",
		"main.gno:9:6", "operator ! not defined on: IntKind",
		"main.gno:10:7", "undefined2",
		"main.gno:12:9", "operator ! not defined on: IntKind",
		"main.gno:14:7", "undefined3",
		"main.gno:18:7", "0 is not a type",
		"main.gno:20:3", "missing field x in interface {}",
	)
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

func TestPreprocessCollectConstCycles(t *testing.T) {
	msgs, panicked := collectWith(t, `package main

const A = 1 + B
const B = C - 1
const C = A + B + 1

func main() {}
`, nil)
	if panicked != nil {
		t.Fatalf("escaped: %v", panicked)
	}
	// Every cycle, at its first member in source order, as gc.
	wantErrors(t, msgs,
		"main.gno:3:", "initialization cycle: A refers to B refers to C refers to A",
		"main.gno:4:", "initialization cycle: B refers to C refers to B",
	)
}
