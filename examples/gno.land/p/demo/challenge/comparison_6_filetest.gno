package main

var a int = 0
var b int64 = 1

func main() {
	println(a + b)
}

// Error:
// main/gno.land/p/demo/challenge/comparison_6_filetest.gno:7: incompatible types in binary expression: a<VPBlock(3,1)> ADD b<VPBlock(3,2)>
