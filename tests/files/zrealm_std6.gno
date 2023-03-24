// PKGPATH: gno.land/r/std_test
package std_test

import (
	"fmt"
	"std"
)

func printOrigCaller() {
	println(std.GetOrigCaller())
}

func main() {
	printOrigCaller()
	std.ExecAsPkg(printOrigCaller)
	printOrigCaller()
}

// Output:
// g1fahslvxakc058u0wa98ytfc9fd6trhe5qfh9ld
// g157y5v3k529jyzhjjz4fn49tzzhf4gess6v39xg
// g1fahslvxakc058u0wa98ytfc9fd6trhe5qfh9ld
