package main

func main() {
	[]int{1, 2} = []int{3, 4}
}

// Error:
// main/assign_literal4.gno:4:2-27: cannot assign to []int{…}

// TypeCheckError:
// main/assign_literal4.gno:4:2: cannot assign to []int{…} (neither addressable nor a map index expression)
