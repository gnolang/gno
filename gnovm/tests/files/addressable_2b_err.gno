package main

func main() {
	_ = &getSlice()
}

func getSlice() []int {
	return []int{1}
}

// Error:
// main/files/addressable_2b_err.gno:4:6: cannot take address of getSlice<VPBlock(3,1)>()
