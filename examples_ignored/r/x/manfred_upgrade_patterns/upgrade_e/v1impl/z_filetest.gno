package main

import (
	home "gno.land/r/x/manfred_upgrade_patterns/upgrade_e"
	_ "gno.land/r/x/manfred_upgrade_patterns/upgrade_e/v1impl"
)

func main() {
	println(home.Render(""))
	println(home.Foo())
}

// Output:
// hello from v1
// not implemented
