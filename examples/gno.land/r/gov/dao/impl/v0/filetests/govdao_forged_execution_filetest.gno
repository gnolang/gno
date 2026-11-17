// PKGPATH: gno.land/r/test/govforge

// A realm holding its own GovDAO cannot drive its crossing methods: executing
// would cross into the executor as impl, a gov/dao subpackage executors trust,
// and run the callback with governance's identity. Only the proxy may call in.
package govforge

import (
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/impl/v0"
)

func main(cur realm) {
	g := impl.NewGovDAO()
	cb := func(cur realm) error {
		println("callback ran as:", cur.Previous().PkgPath())
		return nil
	}
	e := dao.NewSimpleExecutor(0, cur, cb, "forged")
	println("post:", revive(func() {
		g.PostCreateProposal(cross(cur), dao.NewProposalRequest("t", "d", e), dao.ProposalID(1))
	}))
	g.ExecuteProposal(cross(cur), dao.ProposalID(1), e)
}

// Output:
// post: only the gno.land/r/gov/dao proxy may call this

// Error:
// only the gno.land/r/gov/dao proxy may call this
