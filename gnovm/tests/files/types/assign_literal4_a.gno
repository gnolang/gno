package main

func main() {
	[]int{1, 2} := []int{3, 4}
}

// Error:
// files/types/assign_literal4_a.gno:4:2: no new variables on left side of :=

// TypeCheckError:
// main/files/types/assign_literal4_a.gno:4:2: non-name []int{…} on left side of :=
