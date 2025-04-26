package main

var x = 1

const ff = +(1 << x)

func main() {
	println("ok")
}

// Error:
// main/files/const35.gno:5:7: x<VPBlock(2,0)> (variable of type int) is not constant

// TypeCheckError:
// main/files/const35.gno:5:12: +(1 << x) (untyped int value) is not constant
