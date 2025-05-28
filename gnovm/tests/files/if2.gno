package main

import "fmt"

func main() {
	var i int
	if i % 1000000 {
		fmt.Println("oops")
	}
}

// Error:
// main/if2.gno:7:2-9:3: expected typed bool kind, but got IntKind

// TypeCheckError:
// main/if2.gno:7:5: non-boolean condition in if statement
