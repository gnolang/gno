// loader.gno initialises the govDAO v3 implementation and tier structure.
//
// It intentionally does NOT add any members or set AllowedDAOs.  When the
// allowedDAOs list in the DAO proxy is empty, InAllowedDAOs() returns true
// for any caller (see r/gov/dao/proxy.gno), which lets a subsequent MsgRun
// bootstrap the member set and then lock things down.
//
// Bootstrap flow (official network genesis or local dev):
//
//  1. All packages — including this loader — are deployed via MsgAddPackage.
//     The loader sets up tier entries and the DAO implementation.
//  2. A MsgRun executes a setup script (e.g. govdao_prop1.gno) which:
//     a. Adds a temporary deployer as T1 member (for supermajority).
//     b. Creates a governance proposal to register validators, votes YES,
//        and executes it.
//     c. Adds the real govDAO members directly via memberstore.Get().
//     d. Removes the temporary deployer.
//     e. Calls dao.UpdateImpl to set AllowedDAOs, locking down access.
//
// See misc/deployments/ for concrete genesis generation examples.
package loader

import (
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

func init() {
	// Create tier entries in the members tree (required before any SetMember).
	memberstore.Get().SetTier(memberstore.T1)
	memberstore.Get().SetTier(memberstore.T2)
	memberstore.Get().SetTier(memberstore.T3)

	// Set the DAO implementation.  AllowedDAOs is intentionally left empty
	// so that the genesis MsgRun can manipulate the memberstore directly.
	dao.UpdateImpl(cross, dao.UpdateRequest{
		DAO: impl.GetInstance(),
	})
}
