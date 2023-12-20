package main

import "fmt"

// TODO: no support for this expr for now, should fix in another context
func main() {
	x := 11
	y := uint64(-(1 + 1<<x))
	println(y)
	fmt.Printf("%T \n", y)
}

// Error:
// main/files/debug/10a27.gno:8: (const (1 <untyped> bigint)) + (const (1 <untyped> bigint)) << (const-type uint)(x<VPBlock(1,0)>) is not a type
