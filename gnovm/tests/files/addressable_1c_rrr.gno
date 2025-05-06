package main

func main() {
	_ = &getArr()[0]
}

func getArr() [1]int {
	arr := [1]int{1}
	return arr
}

// TypeCheckError:
// main/files/addressable_1c_rrr.gno:4:7: invalid operation: cannot take address of getArr()[0] (value of type int)
