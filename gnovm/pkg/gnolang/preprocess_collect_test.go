package gnolang

import (
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
	want := []string{"declared and not used: x", "declared and not used: y", "label L2 defined and not used", "declared and not used: t"}
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
}

func main() { f(0) }
`)
	if len(msgs) != 1 || !strings.Contains(msgs[0], "a repeated on left side of :=") {
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
	switch fn {
	case nil: // fine: the literal, converted to the tag type
	}
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
		"missing init expr for const declaration",
		"cannot use float32 as int",
		"too few values in struct literal of type main.S",
		"too many values in struct literal of type main.S",
		"index 2 out of bounds [0:2]",
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
		"invalid import path \"a b\": invalid character ' '",
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
		"duplicate key (3 int) in map literal",
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

// collectCodes returns the error codes of a collect run, in order.
func collectCodes(t *testing.T, src string) []ErrCode {
	t.Helper()
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", src)
	if err != nil {
		t.Fatal(err)
	}
	fset := &FileSet{Files: []*FileNode{fn}}
	pn := NewPackageNode("main", "main", fset)
	var codes []ErrCode
	for _, e := range PreprocessFilesCollect(m.Store, pn, fset) {
		codes = append(codes, e.Code())
	}
	return codes
}

func TestPreprocessCollectCodes(t *testing.T) {
	codes := collectCodes(t, `package main

import "a b"

var _ map[func()]int

func f(x int) {
	y := 1
	switch x {
	case 0, 0:
	}
	len("a")
}

func main() {}
`)
	want := []ErrCode{CodeImportPath, CodeMapKey, CodeDuplicateCase, CodeNotUsed, CodeUnusedVar}
	if len(codes) != len(want) {
		t.Fatalf("want %v, got %v", want, codes)
	}
	for i := range want {
		if codes[i] != want[i] {
			t.Errorf("error %d: want %s, got %s", i, want[i], codes[i])
		}
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
