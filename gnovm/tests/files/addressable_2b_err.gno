package main

func main() {
	_ = &getSlice()
}

func getSlice() []int {
	return []int{1}
}

// Error:
// illegal assignment X expression type *gnolang.CallExpr

// TypeCheckError:
// main/addressable_2b_err.gno:4:7: invalid operation: cannot take address of getSlice() (value of type []int)
