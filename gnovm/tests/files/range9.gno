package main

func main() {
	for range 1000 {
	}
}

// Error:
// main/range9.gno:4:2-5:3: range iteration requires map, string, array, slice, or pointer to array; got BigintKind
