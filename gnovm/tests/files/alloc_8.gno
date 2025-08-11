// MAXALLOC: 50000000
package main

func recurse(n int) {
	recurse(n + 1)
}
func main() {
	recurse(0)
}

// Error:
// allocation limit exceeded
