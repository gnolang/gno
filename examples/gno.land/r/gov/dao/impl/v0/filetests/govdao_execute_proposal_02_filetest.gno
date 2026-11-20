// PKGPATH: gno.land/r/test/govdao

package govdao

import (
	"testing"

	"gno.land/r/gov/dao/impl/v0"
)

var govdao *impl.GovDAO

func main(cur realm) {
	govdao = impl.NewGovDAO(cross(cur))
	// Try to execute a proposal using a nil executor
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao")) // stand in for the proxy
	govdao.ExecuteProposal(cross(cur), 0, nil)
}

// Error:
// an executor is required to execute the proposal
