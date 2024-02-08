// PKGPATH: gno.land/r/test
package main

var root interface{}

func main() {
	println(root)
	root = 1
	println(root)
}

// Go_Error:

// Gno_Error:
// expected package name [test] but got [main]

// Go_Output:
// (0x0,0x0)
// (0x1043c03e0,0x1043b8890)

// The below tests that the realm's block (of 1 variable) changed.  The first
// element image in the package (block) is for the "main" function, which
// appears first because function declarations are defined in a file before
// vars.
