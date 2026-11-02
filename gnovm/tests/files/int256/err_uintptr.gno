package main

func main() {
	var x uintptr = 1
	println(x)
}

// Error:
// main/err_uintptr.gno:4:6-19: name uintptr not defined in fileset with files [err_uintptr.gno]
