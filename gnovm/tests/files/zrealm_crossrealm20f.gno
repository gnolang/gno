// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"fmt"
	crossrealm "gno.land/r/demo/tests/crossrealm/struct"
)

type Dt *crossrealm.Bar

var b Dt

func main() {
	b = &crossrealm.Bar{A: 22} // this should panic
	//println(b)
	fmt.Println(b)
}

// Error:
// cannot attach a value of a type defined by another realm: heapitem((struct{(22 int)} gno.land/r/demo/tests/crossrealm/struct.Bar))
