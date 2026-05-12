package main

func main() {
	_ = map[[2][]int]int{[2][]int{{1}, {2}}: 10}
}

// Error:
// main/maplit12.gno:4:6-46: invalid map key type [2][]int

// TypeCheckError:
// main/maplit12.gno:4:10: invalid map key type [2][]int
