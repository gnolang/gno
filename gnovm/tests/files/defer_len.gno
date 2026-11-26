package main

// As in Go, a deferred len is refused: its result has nowhere to go.
func main() {
	var a [3]int
	defer len(a)
}

// Error:
// main/defer_len.gno:6:8-14: defer discards result of len

// TypeCheckError:
// main/defer_len.gno:6:8: defer discards result of len(a) (constant 3 of type int)
