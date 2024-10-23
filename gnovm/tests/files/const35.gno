package main

var x = 1

const ff = +(1 << x)

func main() {
	println("ok")
}

// Error:
// main/files/const35.gno:5:7: +(const (1 <untyped> bigint)) << (const-type uint)(x<VPBlock(2,0)>) (variable of type <untyped> bigint) is not constant
