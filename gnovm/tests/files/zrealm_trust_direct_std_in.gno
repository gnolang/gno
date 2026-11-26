// PKGPATH: gno.land/r/zrealm_trust_direct_std_in
// t72stdin reaches bytes only through t72a: the standard library is trusted
// anywhere in the import tree, so a bytes.Reader method writes.
package zrealm_trust_direct_std_in

import (
	"bytes"

	"gno.land/r/tests/vm/t72stdin"
)

func main(cur realm) {
	t72stdin.Fill(cross(cur), bytes.NewReader([]byte("abcd")))
	println("state:", t72stdin.State())
}

// Output:
// state: abcd
