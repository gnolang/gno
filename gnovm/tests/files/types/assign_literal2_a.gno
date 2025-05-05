package main

func main() {
	var a int
	a, 2 := 6, 6
}

// Error:
// files/types/assign_literal2_a.gno:5:2: no new variables on left side of :=

// TypeCheckError:
// main/files/types/assign_literal2_a.gno:5:5: non-name 2 on left side of :=; main/files/types/assign_literal2_a.gno:4:6: declared and not used: a
