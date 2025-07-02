package main

import (
	"gno.land/r/x/manfred_upgrade_patterns/upgrade_f/home"
	"gno.land/r/x/manfred_upgrade_patterns/upgrade_f/v1impl"
)

func main() {
	cross(home.SetImpl)(v1impl.Instance())
	println(home.Render(""))
	println(home.Foo())
}

// Output:
// hello from v1
// not implemented
