package main

func main() {
	1 := 6
}

// Error:
// assign_literal_a.gno:4:2: no new variables on left side of :=

// TypeCheckError:
// main/assign_literal_a.gno:4:2: non-name 1 on left side of :=
