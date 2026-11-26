// PKGPATH: gno.land/r/zrealm_trust_direct_std_out
// t72stdout imports only io: a bytes.Reader method runs read-only there.
package zrealm_trust_direct_std_out

import (
	"bytes"

	"gno.land/r/tests/vm/t72stdout"
)

func main(cur realm) {
	t72stdout.Fill(cross(cur), bytes.NewReader([]byte("abcd")))
	println("state:", t72stdout.State())
}

// Error:
// cannot copy to readonly tainted slice
