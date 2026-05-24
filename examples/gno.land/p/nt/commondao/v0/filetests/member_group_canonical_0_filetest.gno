// PKGPATH: gno.land/r/test/member_group_canonical
package member_group_canonical

import (
	"gno.land/p/nt/commondao/v0"
)

func main() {
	g := commondao.NewMemberGroup("name", commondao.NewMemberStorage())
	println(commondao.IsCanonicalMemberGroup(g))
}

// Output:
// true
