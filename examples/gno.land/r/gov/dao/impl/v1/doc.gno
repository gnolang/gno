// Package impl (v1) is the GovDAO implementation whose treasury proposals use
// gno.land/r/gov/dao/treasury/v1 (uint256 amounts, grc20 v1 tokens). It is
// otherwise impl/v0, with the same member store (memberstore/v0), law and
// proposal builders.
//
// Governance installs it by voting on NewInstallRequest, which asks
// dao.UpdateImpl for AllowedDAOs = [impl/v0, impl/v1]. Proposal statuses live
// on the implementation instance, so proposals still open under impl/v0 when
// the install executes cannot be voted on or executed afterwards.
package impl
