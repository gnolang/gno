// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import "gno.land/p/demo/tests/p_crossrealm"

var b0 *p_crossrealm.Container

// XXX, this works, attach reference with type defined in p
func init() {
	b0 = &p_crossrealm.Container{
		A: 1,
	}
}

func main() {
	print(".")
}

// Output:
// .

// Error:
