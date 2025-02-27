// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"errors"
	"io"
)

func main() {
	a := &io.EOF
	*a = errors.New("modified")
	println(a)
}

// Error:
// cannot modify external-realm or non-realm object
