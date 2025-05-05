package main

import (
	"gno.land/r/x/manfred_upgrade_patterns/upgrade_c/root"
	v1 "gno.land/r/x/manfred_upgrade_patterns/upgrade_c/v1"
	v2 "gno.land/r/x/manfred_upgrade_patterns/upgrade_c/v2"
)

func main() {
	println("# v1 impl")
	println("root.Get()", root.Get())
	println("v1.Get()", v1.Get())
	println("v1.Inc()", v1.Inc())
	println("v1.Inc()", v1.Inc())
	println("v1.Inc()", v1.Inc())
	println("v1.Get()", v1.Get())
	println()

	println("# v2 impl")
	root.SetCurrentImpl("gno.land/r/x/manfred_upgrade_patterns/upgrade_c/v2")
	println("v2.Get()", v2.Get())
	println("v2.Inc()", v2.Inc())
	println("v2.Inc()", v2.Inc())
	println("v2.Inc()", v2.Inc())
	println("v2.Get()", v2.Get())
	println()

	println("# getters")
	println("root.Get()", root.Get())
	println("v1.Get()", v1.Get())
	println("v2.Get()", v2.Get())
}

// Output:
// # v1 impl
// root.Get() 0
// v1.Get() 0
// v1.Inc() 1
// v1.Inc() 2
// v1.Inc() 3
// v1.Get() 3
//
// # v2 impl
// v2.Get() 6
// v2.Inc() 1003
// v2.Inc() 2003
// v2.Inc() 3003
// v2.Get() 6006
//
// # getters
// root.Get() 3003
// v1.Get() 3003
// v2.Get() 6006
