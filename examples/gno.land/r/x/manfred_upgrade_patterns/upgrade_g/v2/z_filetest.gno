package main

import (
	"fmt"

	"gno.land/r/x/manfred_upgrade_patterns/upgrade_d/v2"
)

func main() {
	println("a", v1.Get("a"))
	println("b", v1.Get("b"))
	println("c", v1.Get("c"))
	println("d", v1.Get("d"))
	println("e", v1.Get("e"))
}

// Output:
// a v2:a-imported-from-v1:1000:true
// b v2:b-imported-from-v1:2000:true
// c v2:c:3:true
// d v2:d:4:false
// e nil
