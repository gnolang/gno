// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/struct"
import "fmt"

type LocalBar *crossrealm.Bar // no attach

var lb LocalBar

func main() {
	lb = &crossrealm.Bar{A: 1}
}

// Error:
// cannot attach a reference to an unreal object from an external realm
