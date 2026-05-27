package main

// Tests that method bodies are analysed for package-level variable dependencies.
// A = T{}.GetB() has no direct syntactic dependency on B, but GetB's body
// reads B. Compared to initorder9, this test ensures that a dependency on GetB
// does not recurse infinitely.
// Init order: T (type, no deps), B (no deps), A (deps T and B via GetB).

type T struct{}

func (T) GetB(i int) int {
	if i < 3 {
		return T{}.GetB(i + 1)
	}
	return B
}

var (
	A = T{}.GetB(0)
	B = 42
)

func main() { println(A) }

// Output:
// 42
