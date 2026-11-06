// Verifies .gno opt-in: presence of // GoOutput: triggers the
// symmetric Gno-vs-Go comparison. Gno's side stays the native
// // Output: golden; like a .go corpus file, the opted-in file ends
// with exactly one // Verdict: line.

package main

import "fmt"

func main() {
	fmt.Println("hello from .gno")
}

// Output:
// hello from .gno

// GoOutput:
// hello from .gno

// Verdict: Clean run
