// MAXALLOC: 10000

package main

var s0 = make([]byte, 120)
var s = make([]byte, 100_000)

func main() {
	println("hello")
}

// Error:
// allocation limit exceeded
