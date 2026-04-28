package main

func main() {
	_ = &getSlice()
}

func getSlice() []int {
	return []int{1}
}

// Error:
// main/addressable_2b_err.gno:4:6-17: invalid operation: cannot take address of getSlice<VPBlock(3,1)>() (value of type []int)

// TypeCheckError:
// main/addressable_2b_err.gno:4:7: invalid operation: cannot take address of getSlice() (value of type []int)
