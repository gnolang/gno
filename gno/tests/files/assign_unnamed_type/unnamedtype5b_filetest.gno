package main

type op func(int) int

func popFn1() op {
	return func(n int) int {
		n = n + 1
		return n
	}
}

func popFn2() func(int) int {
	var dec op

	dec = func(n int) int {
		n = n - 1
		return n
	}

	return dec
}

func main() {
	println(popFn1()(0))
	println(popFn2()(0))
}

// Output:
// 1
// -1
