package main

func main() {
	_ = getArr()[:]
}

func getArr() [1]int {
	return [1]int{1}
}

// TypeCheckError:
// main/addressable_1d_err.gno:4:6: invalid operation: cannot slice getArr() (value of type [1]int) (value not addressable)