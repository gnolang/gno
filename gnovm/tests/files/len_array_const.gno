package main

// len and cap of an array, or of a pointer to one, are constant when the
// operand holds no call, and the operand is then not evaluated, as in Go. An
// operand with a call is evaluated.

type A [2]int

var calls int

func f() [3]int {
	calls++
	return [3]int{}
}

func g() *A {
	calls++
	return nil
}

func main() {
	var a [4][5]int
	i := 10
	println(len(a[i]), cap(a[i]))

	var p *[2]int
	println(len(*p), len(p))

	var q *A
	println(len(q), cap(*q))

	n := len(f())
	println(n, calls)

	m := len([2]int{f()[0], 0})
	println(m, calls)

	k := len(g())
	println(k, calls)

	c := cap(g())
	println(c, calls)
}

// Output:
// 5 5
// 2 2
// 2 2
// 3 1
// 2 2
// 2 3
// 2 4
