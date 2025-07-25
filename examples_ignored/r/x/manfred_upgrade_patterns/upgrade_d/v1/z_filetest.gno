package main

import (
	v1 "gno.land/r/x/manfred_upgrade_patterns/upgrade_d/v1"
)

func main() {
	println("a", v1.Get("a"))
	println("b", v1.Get("b"))
	println("c", v1.Get("c"))
}

// Output:
// a v1:a:1
// b v1:b:2
// c nil
