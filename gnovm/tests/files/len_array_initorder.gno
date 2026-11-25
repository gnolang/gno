package main

// len and cap of an array fold to a constant, but the operand still orders
// package initialization, as in Go, also inside a larger constant: every var
// below waits for a.

var x = show("x", len(a))

var y = f()

var z = show("z", len(a)-1)

var w = show("w", int(cap(b[len(a)-1])))

var a = mk()

var b [3][2]int

func show(s string, n int) int {
	println(s, n)
	return n
}

func f() int {
	return show("y", cap(a))
}

func mk() [3]int {
	println("a")
	return [3]int{}
}

func main() {}

// Output:
// a
// x 3
// y 3
// z 2
// w 2
