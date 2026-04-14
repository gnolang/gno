package main

func main() {
	var i = "h"
	b := make([]byte, i)

	println(b)
}

// Error:
// main/varg_3.gno:5:7-22: invalid argument: index i<VPBlock(1,0)> (variable of type string) must be integer

// TypeCheckError:
// main/varg_3.gno:5:20: invalid argument: index i (variable of type string) must be integer
