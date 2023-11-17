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

// 현재 메인 함수에서 println은 실행되지만, defer는 실행되지 않음.
// 아마 내 생각엔 메모리 오류가 발생하는게 이것과 관련이 있을 것 같음.
// 즉, 함수 호출이 끝나면서 스택에 쌓인 defer를 실행하려고 하면, 이미 스택이 비워져서 오류가 발생하는 듯?
// 그럼 어떻게 해결해야 하나?

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