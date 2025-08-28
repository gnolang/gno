package main

import (
	"gno.land/r/x/manfred_upgrade_patterns/upgrade_f/home"
	"gno.land/r/x/manfred_upgrade_patterns/upgrade_f/v1impl"
	"gno.land/r/x/manfred_upgrade_patterns/upgrade_f/v2impl"
)

func main() {
	cross(home.SetImpl)(v1impl.Instance())
	println(home.Render(""))
	println(home.Foo())

	println("-------------")
	cross(home.SetImpl)(v2impl.Instance())
	println(home.Render(""))
	println(home.Foo())
}

// Output:
// hello from v1
// not implemented
// -------------
// hello from v2
// undefined
