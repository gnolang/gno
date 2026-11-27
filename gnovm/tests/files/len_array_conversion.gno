package main

// A conversion is not a call, so len and cap of a converted array stay
// constant and the operand is not evaluated, as in Go. A call inside the
// conversion's operand is still evaluated.

type T [3]int

var calls int

func f() [3]int {
	calls++
	return [3]int{}
}

func main() {
	var p *[3]int
	println(len(T(*p)), cap(T(*p)))

	n := len(T(f()))
	println(n, calls)
}

// Output:
// 3 3
// 3 1
