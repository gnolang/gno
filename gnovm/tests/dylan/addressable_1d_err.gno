package main

func main() {
	_ = getArr()[:]
}

func getArr() [1]int {
	return [1]int{1}
}

// Error:
// main/dylan/addressable_1d_err.gno:4:6: cannot take address of getArr<VPBlock(3,1)>()
