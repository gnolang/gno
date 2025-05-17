package main

func main() {
	_ = &getSlice().([]int)
}

func getSlice() any {
	return []int{1}
}

// Error:
// illegal assignment X expression type *gnolang.TypeAssertExpr

// TypeCheckError:
// main/addressable_6d_err.gno:4:7: invalid operation: cannot take address of getSlice().([]int) (comma, ok expression of type []int)
