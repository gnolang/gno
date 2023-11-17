package gnolang

import (
	"testing"
)

func TestPrintlnPrintNil(t *testing.T) {
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

func TestComposite(t *testing.T) {
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

	assertOutput(t, c, "simple panic\nrecover\n")
}

// TODO: Resolve runtime error
// current output: g recover <nil> wtf
func TestRecover(t *testing.T) {
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

	assertOutput(t, c, "g recover wtf\nf recover wtf\n")
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

	assertOutput(t, c, "simple panic\nnested panic\nouter recover\n")
}
