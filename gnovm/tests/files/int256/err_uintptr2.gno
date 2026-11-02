package main

func main() {
	var x uint256 = 1
	println(uintptr(x))
}

// Error:
// main/err_uintptr2.gno:5:10-17: name uintptr not declared
