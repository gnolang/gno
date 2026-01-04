package main

func main() {
	a, b := 1, 2
	a, b := 3, 4
}

// Error:
// redefine6.gno:5:2: no new variables on left side of :=

// TypeCheckError:
// main/redefine6.gno:5:7: no new variables on left side of :=; main/redefine6.gno:4:2: declared and not used: a; main/redefine6.gno:4:5: declared and not used: b
