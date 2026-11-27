// PKGPATH: gno.land/r/zrealm_import_nested
// An interface call made inside trusted stdlib code is still checked against
// the realm that made it: io.SectionReader over an untrusted bytes.Reader.
package zrealm_import_nested

import (
	"bytes"
	"io"

	"gno.land/r/tests/vm/impvictim"
)

func main(cur realm) {
	impvictim.Fill(cross(cur), io.NewSectionReader(bytes.NewReader([]byte("abcd")), 0, 4))
	println("victim:", impvictim.State())
}

// Error:
// cannot copy to readonly tainted slice
