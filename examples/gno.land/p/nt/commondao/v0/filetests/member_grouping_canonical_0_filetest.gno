// PKGPATH: gno.land/r/test/member_grouping_canonical
package member_grouping_canonical

import (
	"gno.land/p/nt/commondao/v0"
)

func main() {
	g := commondao.NewMemberGrouping()
	println(commondao.IsCanonicalMemberGrouping(g))
}

// Output:
// true
