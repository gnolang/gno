// PKGPATH: gno.land/r/std_test
package std_test

import (
	"std"
)

func main() {
	caller := std.GetOrigCaller()
	println(caller)
}

// Output:
// g157y5v3k529jyzhjjz4fn49tzzhf4gess6v39xg
