// PKGPATH: gno.land/r/test/govdao
package govdao

import (
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
)

var (
	executor dao.Executor
	govdao   = impl.NewGovDAO()
)

func init(cur realm) {
	// Create a dummy executor
	cb := func(realm) error { return nil }
	executor = dao.NewSimpleExecutor(0, cur, cb, "")
}

func main(cur realm) {
	// Try to execute a proposal that doesn't exist
	govdao.ExecuteProposal(0, cur, 404, executor)
}

// Error:
// proposal not found
