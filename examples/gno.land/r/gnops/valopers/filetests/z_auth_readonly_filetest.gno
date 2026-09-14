// PKGPATH: gno.land/r/authreadonly
//
// Auth() exposes the realm's governance authority for rendering and
// inspection only. It returns a read-only string, never the live
// *authz.Authorizer.
//
// SCOPE. This asserts two things, one of which is now security-relevant:
//
//  1. Auth()'s RETURN TYPE — that the mutable handle is gone from the
//     type system. This is what makes an
//     Auth().Transfer(...) PoC stop compiling.
//  2. The PROPOSER, since ContractAuthority.String() now renders it.
//     Previously the string was byte-identical whether the authority was
//     the gated default or a wide-open AutoAccept proposer, so reverting
//     the fix left this file green. It no longer does.
//
// Reachability is still pinned elsewhere, not here:
//   - p/moul/authz TestDefaultProposerIsContractOnly,
//     TestExternalTransferRejected, TestExternalDoByPreviousRejected,
//     TestContractPathIsolation,
//     TestForeignFrameCannotDriveContractAuthority and
//     TestContractAuthorityUnauthorizedCaller — the contract-identity
//     gate. Reverting it fails those 6 plus TestAuthorityString,
//     TestExplicitProposerReplacesIdentityGate,
//     TestSpoofedProposerCannotImpersonateContractIdentity and
//     filetests/z_contract_authority_shape_filetest.gno, since the
//     rendered proposer is now part of what is pinned.
//   - ./z_foreign_realm_capability_filetest.gno — the consumer-level
//     one: a foreign realm cannot reach the privileged capability.
//   - ./z_govdao_only_principal_filetest.gno — that the repointed
//     authority names GovDAO and that the principal cannot be MINTED:
//     it drives the forged-executor route and MustCreateProposal, and
//     asserts both are refused.
package authreadonly

import (
	"gno.land/r/gnops/valopers"
)

func main(cur realm) {
	// An external realm observes a description; there is no mutator to
	// reach from here.
	println("authority:", valopers.Auth())
}

// Output:
// authority: contract_authority[contract=gno.land/r/gov/dao,proposer=contract-identity]
