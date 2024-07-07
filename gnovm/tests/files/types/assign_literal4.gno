package main

func main() {
	[]int{1, 2} = []int{3, 4}
}

// Error:
// main/files/types/assign_literal4.gno:4:2: cannot assign to [](const-type int){(const (1 int)), (const (2 int))}
