// PKGPATH: gno.land/r/std_test
package std_test

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
// g157y5v3k529jyzhjjz4fn49tzzhf4gess6v39xg
