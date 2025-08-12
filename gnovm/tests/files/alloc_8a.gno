// MAXALLOC: 1000
package main

var counter int

func recurse(n int) {
	counter++
	if counter == 100 {
		return
	}
	recurse(n + 1)
}
func main() {
	recurse(0)
}

// Error:
// allocation limit exceeded
