// PKGPATH: gno.land/r/test
package main

var root interface{}

func main() {
	println(root)
	root = 1
	println(root)
}

// Error:
// expected package name [test] but got [main]

// Go_Output:
// (0x0,0x0)
// (0x1046e07c0,0x1046dab60)
