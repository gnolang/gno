// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"io"
)

func main() {
	a := &io.EOF
	println(a)
}

// Output:
// &(&(ref(p:4021fadf187fd2c8277aad0fbc5e8fc4e7f1c75f:10) errors.errorString) *errors.errorString)
