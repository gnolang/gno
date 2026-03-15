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

func init() {
	// Create a dummy executor
	cb := func(realm) error { return nil }
	executor = dao.NewSimpleExecutor(cb, "")
}

func main() {
	// Try to execute a proposal that doesn't exist
	govdao.ExecuteProposal(404, executor)
}

// Error:
// proposal not found
