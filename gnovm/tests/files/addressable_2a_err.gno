package main

func main() {
	_ = &[]int{1}[:]
}

// Error:
// illegal assignment X expression type *gnolang.SliceExpr

// TypeCheckError:
// main/addressable_2a_err.gno:4:7: invalid operation: cannot take address of []int{…}[:] (value of type []int)
