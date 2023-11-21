package gnolang

import (
	"testing"
)

func TestPrintlnWithNilValue(t *testing.T) {
	m := NewMachine("test", nil)

	c := `package test
	func main() {
		println(nil)
	}`

	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()

	assertOutput(t, c, "undefined\n")
}

func TestPrintlnWithEmptySlice(t *testing.T) {
	m := NewMachine("test", nil)

	c := `package test
	func main() {
		var a []string
		println(a)
	}`
	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()
	assertOutput(t, c, "undefined\n")
}

func TestPrintlnWithNonNilSlice(t *testing.T) {
	m := NewMachine("test", nil)
	c := `package test
	func main() {
		a := []string{"a", "b"}
		println(a)
	}`

	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()

	assertOutput(t, c, "slice[(\"a\" string),(\"b\" string)]\n")
}

func TestPrintlnFunction(t *testing.T) {
	m := NewMachine("test", nil)

	c := `package test
	func foo(a, b int) int {
		return a + b
	}
	func main() {
		println(foo(1, 3))
	}`

	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()

	assertOutput(t, c, "4\n")
}

func TestCompositeSlice(t *testing.T) {
	m := NewMachine("test", nil)
	c := `package test
	func main() {
		a, b, c, d := 1, 2, 3, 4
		x := []int{
			a: b,
			c: d,
		}
		println(x)
	}`

	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()

	assertOutput(t, c, "slice[(0 int),(2 int),(0 int),(4 int)]\n")
}

func TestSimpleRecover(t *testing.T) {
	m := NewMachine("test", nil)
	c := `package test

    func main() {
        defer func() { println("recover", recover()) }()
        println("simple panic")
    }`

	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()

	assertOutput(t, c, "simple panic\nrecover undefined\n")
}

func TestRecoverWithPanic(t *testing.T) {
	m := NewMachine("test", nil)
	c := `package test

	func main() {
		f()
	}

	func f() {
		defer func() { println("f recover", recover()) }()
		defer g()
		panic("wtf")
	}

	func g() {
		println("g recover", recover())
	}`

	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()

	assertOutput(t, c, "g recover wtf\nf recover undefined\n")
}

func TestNestedRecover(t *testing.T) {
	m := NewMachine("test", nil)
	c := `package test

    func main() {
        defer func() { println("outer recover", recover()) }()
        defer func() { println("nested panic") }()
        println("simple panic")
    }`

	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()

	assertOutput(t, c, "simple panic\nnested panic\nouter recover undefined\n")
}

func TestFunction(t *testing.T) {
	m := NewMachine("test", nil)
	c := `package test
	func f() func() {
		return nil
	}
	
	func main() {
		g := f()
		println(g)
		if g == nil {
			println("nil func")
		}
	}`
	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()

	assertOutput(t, c, "func()()\nnil func\n")
}
