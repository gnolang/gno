package main

func main() {
	_ = &getSlice().([]int)
}

func getSlice() any {
	return []int{1}
}

// Error:
// main/addressable_6d_err.gno:4:6-25: invalid operation: cannot take address of getSlice<VPBlock(3,1)>().([](const-type int)) (value of type []int)

// TypeCheckError:
// main/addressable_6d_err.gno:4:7: invalid operation: cannot take address of getSlice().([]int) (comma, ok expression of type []int)
