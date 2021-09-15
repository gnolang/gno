package gno

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	//"github.com/davecgh/go-spew/spew"
	"github.com/jaekwon/testify/assert"
)

// run empty main().
func TestRunEmptyMain(t *testing.T) {
	m := NewMachine("test", nil)
	main := FuncD("main", nil, nil, nil)
	m.RunDeclaration(main)
	m.RunMain()
}

// run main() with a for loop.
func TestRunLoopyMain(t *testing.T) {
	m := NewMachine("test", nil)
	c := `package test
func main() {
	for i:=0; i<1000; i++ {
		if i == -1 {
			return
		}
	}
}`
	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()
}

func TestEval(t *testing.T) {
	m := NewMachine("test", nil)
	c := `package test
func next(i int) int {
	return i+1
}`
	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	res := m.Eval(Call("next", "1"))
	fmt.Println(res)
}

func assertOutput(t *testing.T, input string, output string) {
	buf := new(bytes.Buffer)
	pn := NewPackageNode("test", ".test", &FileSet{})
	pkg := pn.NewPackage(nil)
	m := NewMachineWithOptions(MachineOptions{
		Package: pkg,
		Output:  buf,
	})
	n := MustParseFile("main.go", input)
	m.RunFiles(n)
	m.RunMain()
	assert.Equal(t, string(buf.Bytes()), output)
	err := m.CheckEmpty()
	assert.Nil(t, err)
}

func TestRunMakeStruct(t *testing.T) {
	assertOutput(t, `package test
type Outfit struct {
	Scarf string
	Shirt string
	Belts string
	Strap string
	Pants string
	Socks string
	Shoes string
}
func main() {
	s := Outfit {
		// some fields are out of order.
		// some fields are left unset.
		Scarf:"scarf",
		Shirt:"shirt",
		Shoes:"shoes",
		Socks:"socks",
	}
	// some fields out of order are used.
	// some fields left unset are used.
	print(s.Shoes+","+s.Shirt+","+s.Pants+","+s.Scarf)
}`, `shoes,shirt,,scarf`)
}

func TestRunReturnStruct(t *testing.T) {
	assertOutput(t, `package test
type MyStruct struct {
	FieldA string
	FieldB string
}
func myStruct(a, b string) MyStruct {
	return MyStruct{
		FieldA: a,
		FieldB: b,
	}
}
func main() {
	s := myStruct("aaa", "bbb")
	print(s.FieldA+","+s.FieldB)
}`, `aaa,bbb`)
}

//----------------------------------------
// Benchmarks

func BenchmarkPreprocess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// stop timer
		b.StopTimer()
		pkg := &PackageNode{
			PkgName: "main",
			PkgPath: ".main",
			FileSet: nil,
		}
		pkg.InitStaticBlock(pkg, nil)
		main := FuncD("main", nil, nil, Ss(
			A("mx", ":=", "1000000"),
			For(
				A("i", ":=", "0"),
				X("i < mx"),
				Inc("i"),
			),
		))
		b.StartTimer()
		// timer started
		main = Preprocess(nil, pkg, main).(*FuncDecl)
	}
}

func BenchmarkLoopyMain(b *testing.B) {
	m := NewMachine("test", nil)
	main := FuncD("main", nil, nil, Ss(
		A("mx", ":=", "10000000"),
		For(
			A("i", ":=", "0"),
			// X("i < 10000000"),
			X("i < mx"),
			Inc("i"),
		),
	))
	m.RunDeclaration(main)
	for i := 0; i < b.N; i++ {
		m.RunMain()
	}
}

//----------------------------------------
// Unsorted

type Struct1 struct {
	A int
	B int
}

func TestModifyTypeAsserted(t *testing.T) {
	x := Struct1{1, 1}
	var v interface{} = x
	x2 := v.(Struct1)
	x2.A = 2

	// only x2 is changed.
	assert.Equal(t, x.A, 1)
	assert.Equal(t, v.(Struct1).A, 1)
	assert.Equal(t, x2.A, 2)
}

type Interface1 interface {
	Foo()
}

func TestTypeConversion(t *testing.T) {
	x := 1
	var v interface{} = x
	if _, ok := v.(Interface1); ok {
		panic("should not happen")
	}
	v = nil
	if _, ok := v.(Interface1); ok {
		panic("should not happen")
	}
	assert.Panics(t, func() {
		// this would panic.
		z := v.(Interface1)
		fmt.Println(z)
	})

}

func TestSomething(t *testing.T) {
	type Foo struct {
		X interface{}
	}

	type Bar struct {
		X interface{}
		Y bool
	}

	fmt.Println(unsafe.Sizeof(Foo{}))                        // 16
	fmt.Println(unsafe.Sizeof(Foo{X: reflect.ValueOf(0)}.X)) // still 16? weird.
	fmt.Println(unsafe.Sizeof(reflect.ValueOf(0)))           // 24
	fmt.Println(unsafe.Sizeof(Bar{}))

}

// XXX is there a way to test in Go as well as Gno?
func TestDeferOrder(t *testing.T) {
	a := func() func(int, int) int {
		fmt.Println("a constructed")
		return func(x int, y int) int {
			fmt.Println("a called")
			return x + y
		}
	}
	b := func() int {
		fmt.Println("b constructed")
		return 1
	}
	c := func() int {
		fmt.Println("c constructed")
		return 2
	}
	defer a()(b(), c())
	fmt.Println("post defer")

	// should print
	// a constructed
	// b constructed
	// c constructed
	// post defer
	// a called
}

// XXX is there a way to test in Go as well as Gno?
func TestCallOrder(t *testing.T) {
	a := func() func(int, int) int {
		fmt.Println("a constructed")
		return func(x int, y int) int {
			fmt.Println("a called")
			return x + y
		}
	}
	b := func() int {
		fmt.Println("b constructed")
		return 1
	}
	c := func() int {
		fmt.Println("c constructed")
		return 2
	}
	a()(b(), c())

	// should print
	// a constructed
	// b constructed
	// c constructed
	// a called
}

// XXX is there a way to test in Go as well as Gno?
func TestBinaryShortCircuit(t *testing.T) {
	tr := func() bool {
		fmt.Println("t called")
		return true
	}
	fa := func() bool {
		fmt.Println("f called")
		return false
	}
	if fa() && tr() {
		fmt.Println("error")
	} else {
		fmt.Println("done")
	}

}

// XXX is there a way to test in Go as well as Gno?
func TestSwitchDefine(t *testing.T) {
	var x interface{} = 1
	switch y := x.(type) {
	case int:
		fmt.Println("int", y) // XXX
	default:
		fmt.Println("not int")
	}
}

// XXX is there a way to test in Go as well as Gno?
func TestBinaryCircuit(t *testing.T) {
	tr := func() bool {
		fmt.Println("tr() called")
		return true
	}
	fa := func() bool {
		fmt.Println("fa() called")
		return false
	}

	fmt.Println("case 1")
	fmt.Println(fa() && tr())
	fmt.Println("case 1")
	fmt.Println(tr() || fa())

	// should print
	// case 1
	// fa() called
	// false
	// case 1
	// tr() called
	// true
}

func TestMultiAssignment(t *testing.T) {
	buf := make([]int, 4)
	ref := func(i int) *int {
		fmt.Printf("ref(%v) called\n", i)
		return &buf[i]
	}
	val := func(i int) int {
		fmt.Printf("val(%v) called\n", i)
		return i
	}

	*ref(0), *ref(1), *ref(2), *ref(3) =
		val(11), val(22), val(33), val(44)

	/*
		ref(0) called
		ref(1) called
		ref(2) called
		ref(3) called
		val(11) called
		val(22) called
		val(33) called
		val(44) called
	*/
}

// XXX is there a way to test in Go as well as Gno?
func TestCallLHS(t *testing.T) {
	x := 1
	xptr := func() *int {
		return &x
	}
	*xptr() = 2
	assert.Equal(t, x, 2)
}

// XXX is there a way to test in Go as well as Gno?
func TestCallFieldLHS(t *testing.T) {
	type str struct {
		X int
	}
	x := str{}
	xptr := func() *str {
		return &x
	}
	y := 0
	xptr().X, y = 2, 3
	assert.Equal(t, x.X, 2)
	assert.Equal(t, y, 3)
}
