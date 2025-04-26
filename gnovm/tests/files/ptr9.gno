package main

func main() {
	v := 1
	println(*v)
}

// Error:
// main/files/ptr9.gno:5:10: invalid operation: cannot indirect v<VPBlock(1,0)> (variable of type int)

// TypeCheckError:
// main/files/ptr9.gno:5:11: invalid operation: cannot indirect v (variable of type int)
