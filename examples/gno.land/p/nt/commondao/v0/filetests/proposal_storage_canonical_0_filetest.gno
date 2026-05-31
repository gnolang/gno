// PKGPATH: gno.land/r/test/proposal_storage_canonical
package proposal_storage_canonical

import (
	"gno.land/p/nt/commondao/v0"
)

func main(cur realm) {
	s := commondao.NewProposalStorage()
	println(commondao.IsCanonicalProposalStorage(s))
}

// Output:
// true
