// PKGPATH: gno.land/r/test/member_storage_canonical
package member_storage_canonical

import (
	"gno.land/p/nt/commondao/v0"
)

func main(cur realm) {
	s := commondao.NewMemberStorage()
	println(commondao.IsCanonicalMemberStorage(s))
}

// Output:
// true
