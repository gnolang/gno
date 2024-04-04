package main

type op func(int) int

func exec1(opFn func(int) int) {
	println(opFn)
	println(opFn(0))
}

func exec2(opFn op) {
	println(opFn)
	println(opFn(0))
}

func main() {
	var inc op
	inc = func(n int) int {
		n = n + 1
		return n
	}
	dec := func(n int) int {
		n = n - 1
		return n
	}
	exec1(inc)
	exec2(dec)
}

// Output:
// 1
// ( main.op)
// -1
