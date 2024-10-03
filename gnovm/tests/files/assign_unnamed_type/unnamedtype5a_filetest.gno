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
// func(n int)( int){...}
// 1
// (func(n int)( int){...} main.op)
// -1
