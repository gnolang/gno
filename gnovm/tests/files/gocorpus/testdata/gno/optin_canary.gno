// Verifies .gno opt-in: presence of // GoOutput: triggers the
// symmetric Gno-vs-Go comparison. Here Gno matches Go (fmt.Println
// produces the same line in both), so no // Divergence: is needed.

package main

import "fmt"

func main() {
	fmt.Println("hello from .gno")
}

// Output:
// hello from .gno

// GoOutput:
// hello from .gno
