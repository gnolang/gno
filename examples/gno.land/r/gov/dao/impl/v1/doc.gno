// Package impl (v1) is the GovDAO implementation whose treasury proposals use
// gno.land/r/gov/dao/treasury/v1 (uint256 amounts, grc20 v1 tokens). It is
// otherwise impl/v0, with the same member store (memberstore/v0) and proposal
// builders. It does not share v0's law: it starts with its own default
// (Supermajority 66.66%), because v0 exposes no way to read its current law.
// NewInstallRequest says so in the proposal text; a DAO that changed its law
// under v0 restores it with NewChangeLawRequest after the install.
//
// Proposal filters fail closed: impl/v0's FilterByTier is enforced with the
// same rules as impl/v1's (v0 builders stay usable after the install), and
// any other filter is refused at proposal creation.
//
// Governance installs it by voting on NewInstallRequest, which asks
// dao.UpdateImpl for AllowedDAOs = [impl/v0, impl/v1]. Proposal statuses live
// on the implementation instance, so proposals still open under impl/v0 when
// the install executes cannot be voted on or executed afterwards.
package impl
