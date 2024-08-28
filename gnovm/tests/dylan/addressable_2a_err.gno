package main

func main() {
	_ = &[]int{1}[:]
}

// Error:
// illegal assignment X expression type *gnolang.SliceExpr
