// PKGPATH: gno.land/r/test
package test

import (
	"std"
)

var (
	realmAddr std.Address = std.GetOrigPkgAddr()
)

func main() {
	println(realmAddr)
}

// Output:
// g148esskg7wzxxw6axwv4j5nkgjzdp6v4zmy3pwc
