// MAXALLOC: 500000000
package main

func grow() {
	var hugeString string
	for {
		hugeString += string(make([]byte, 100_000_000))
	}
}

func main() {
	grow()
}

// Error:
// allocation limit exceeded
