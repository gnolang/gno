package main

// Compute fibonacci numbers, no memoization
func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-2) + fib(n-1)
}

func main() {
	println(fib(35))
	//println(fib(10))
}

// Output:
// 9227465
