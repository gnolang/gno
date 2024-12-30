// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"fmt"
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type Dt *crossrealm.Bar

var b Dt

func init() {
}

func main() {
	b = &crossrealm.Bar{A: 22} // this should panic
	//println(b)
	fmt.Println(b)
}

// Error:
// cannot attach a reference to an unreal object from an external realm
