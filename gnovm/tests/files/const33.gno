package main

var x = 1
var y = 1

const b = x == y

func main() {
	println("ok")
}

// Error:
// main/const33.gno:6:7-17: x<VPBlock(2,0)> (variable of type int) is not constant

// TypeCheckError:
// main/const33.gno:6:11: x == y (untyped bool value) is not constant
