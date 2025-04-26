package main

var x = "a"

const v = len(x)

func main() {
	println("ok")
}

// Error:
// main/files/const50.gno:5:7: x<VPBlock(2,0)> (variable of type string) is not constant

// TypeCheckError:
// main/files/const50.gno:5:11: len(x) (value of type int) is not constant
