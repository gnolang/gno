// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/struct"
import "fmt"

type LocalBar crossrealm.Bar

var lb LocalBar

func main() {
	fmt.Println(lb)
}

// Error:
// cannot attach a value of a type defined by another realm
