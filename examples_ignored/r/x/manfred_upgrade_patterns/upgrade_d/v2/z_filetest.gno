package main

import v2 "gno.land/r/x/manfred_upgrade_patterns/upgrade_d/v2"

func main() {
	println("a", v2.Get("a"))
	println("b", v2.Get("b"))
	println("c", v2.Get("c"))
	println("d", v2.Get("d"))
	println("e", v2.Get("e"))
}

// Output:
// a v2:a-imported-from-v1:1000:true
// b v2:b-imported-from-v1:2000:true
// c v2:c:3:true
// d v2:d:4:false
// e nil
