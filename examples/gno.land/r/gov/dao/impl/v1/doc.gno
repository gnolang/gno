// Package impl (v1) is the GovDAO implementation whose treasury proposals use
// gno.land/r/gov/dao/treasury/v1 (uint256 amounts, grc20 v1 tokens). It is
// otherwise impl/v0, with the same member store (memberstore/v0) and proposal
// builders. Its law cannot be read from impl/v0 (v0 exposes no accessor), so
// the install proposal carries the law impl/v1 will use, voted on under v0's
// law and in force from impl/v1's first decision (see NewInstallRequest).
//
// Proposal filters fail closed: impl/v0's FilterByTier is enforced with the
// same rules as impl/v1's (v0 builders stay usable after the install), and
// any other filter is refused at proposal creation.
//
// Governance installs it by voting on NewInstallRequest(cur, reason, law), which asks
// dao.UpdateImpl for AllowedDAOs = [impl/v0, impl/v1]. Proposal statuses live
// on the implementation instance, so proposals still open under impl/v0 when
// the install executes cannot be voted on or executed afterwards.
package impl
