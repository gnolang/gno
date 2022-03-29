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
// g1w3jhxarpv3j8yh6lta047h6lta047h6l46ncpj
