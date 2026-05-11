package main

func main() {
	_ = &getArr()[0]
}

func getArr() [1]int {
	arr := [1]int{1}
	return arr
}

// TypeCheckError:
// main/addressable_1c_rrr.gno:4:7: invalid operation: cannot take address of getArr()[0] (value of type int)

// Error:
// main/addressable_1c_rrr.gno:4:6-18: invalid operation: cannot take address of getArr<VPBlock(3,1)>()[(const (0 int))] (value of type int)
