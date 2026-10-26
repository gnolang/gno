package main

func main() {
	_ = getArr()[:]
}

func getArr() [1]int {
	return [1]int{1}
}

// TypeCheckError:
// main/addressable_1d_err.gno:4:6: cannot slice unaddressable value getArr() (value of type [1]int)

// Error:
// main/addressable_1d_err.gno:4:6-17: cannot slice unaddressable value getArr<VPBlock(3,1)>()
