package main

func main() {
	_ = &getSlice().([]int)
}

func getSlice() any {
	return []int{1}
}

// Error:
// main/files/addressable_6d_err.gno:4:6: cannot take address of getSlice<VPBlock(3,1)>().([](const-type int))
