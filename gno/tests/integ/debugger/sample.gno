// This is a sample target gno program to test the gnovm debugger.
// See ../../cmd/gno/debug_test.go for the debugger test cases.

package main

func f(name string, i int) {
	println(name, i)
}

func g(s string, n int) {
	f(s, n)
}

var global = "test"

type T struct {
	A []int
}

func (t *T) get(i int) int {
	r := t.A[i]
	if i == 0 {
		b := "zero"
		println(b)
	} else {
		b := "!zero"
		println(b)
	}
	return r
}

func main() {
	num := 5
	println("in main")
	if num > 2 {
		b := 3
		g("hello", b)
	}
	t := T{A: []int{1, 2, 3} }
	println(t.get(1))
	x := 0
	for i := 0; i < 5; i++ {
		x = i
	}
	println("bye", x)
}
