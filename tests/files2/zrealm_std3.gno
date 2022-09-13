// PKGPATH: gno.land/r/std_test
package std_test

import (
	"std"
)

func displayOrigCaller() {
	println(std.GetOrigCaller())
}

func main() {
	displayOrigCaller()
	std.ExecAsPkg(func() {
		displayOrigCaller()
	})
	displayOrigCaller()
}

// Output:
// g157y5v3k529jyzhjjz4fn49tzzhf4gess6v39xg
// TODO: should be the address package
// g157y5v3k529jyzhjjz4fn49tzzhf4gess6v39xg
