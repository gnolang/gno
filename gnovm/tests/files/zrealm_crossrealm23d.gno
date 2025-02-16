// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"errors"
	"io"
)

func main() {
	io.EOF = errors.New("modified")
}

// Error:
// cannot modify external-realm or non-realm object
