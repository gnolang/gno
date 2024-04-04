package main

type op func(int) int

func inc(n int) int {
	n = n + 1
	return n
}

func dec(n int) int {
	n = n - 1
	return n
}

func main() {
	var a op
	a = inc

	b := func(n int) int {
		return n
	}
	c := dec

	println(a(0), a)
	println(b(0), b)
	println(c(0), c)

	a, b, c = b, c, a

	println(a(0), a)
	println(b(0), b)
	println(c(0), c)
}

// Output:
// 1 (inc main.op)
// 0
// -1 dec
// 0 ( main.op)
// -1 dec
// 1 inc
