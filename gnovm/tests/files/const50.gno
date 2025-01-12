package main

var x = "a"

const v = len(x)

func main() {
	println("ok")
}

// Error:
// main/files/const50.gno:8:7: x<VPBlock(2,0)> (variable of type string) is not constant
