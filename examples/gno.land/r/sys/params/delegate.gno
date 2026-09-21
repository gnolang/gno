package params

// Delegation: letting a realm other than GovDAO manage one specific parameter.
//
// This realm holds a capability nothing else can: the `sys/params` stdlib
// refuses every caller except `gno.land/r/sys/params`, checked in the VM by
// package path, and the same gate covers the getters. So every parameter write
// on the chain physically originates here, and "delegating a parameter" can only
// mean adding a path in this file that authorizes someone other than a GovDAO
// vote.
//
// Two shapes are deliberately NOT used.
//
// Not a registry of key -> realm. This realm can never be redeployed
// (AddPackage refuses an occupied path), and the stdlib gate names this exact
// path, so making a NEW parameter delegatable already requires editing this
// file, which means a chain relaunch. A registry's runtime generality would
// therefore never be exercised: the only freedom that matters is "which realm,
// or none" for keys already blessed in source. A named slot says exactly that
// and nothing more. It also removes a whole class of mistake — there is no
// key string to mis-compose, no container to accidentally expose, and no way for
// a proposal to name a key the author never considered. A shape-based allowlist
// like "<module>:p:<name>" would have admitted bank:p:restricted_denoms and
// auth:p:unrestricted_addrs, which is not a delegation anyone asked for.
//
// Not a capability object. Returning something the delegate holds would make
// authority survive revocation, because the check would have happened at
// construction. The authority is re-checked on every crossing call instead, so
// clearing the slot takes effect immediately.

import (
	"chain"

	"gno.land/r/gov/dao"
)

// Event names. PascalCase with a named const is the house style across r/sys
// and r/gov; the bare lowercase "set" elsewhere in this realm predates it.
const (
	DelegateSetEvent     = "ParamDelegateSet"
	DelegateClearedEvent = "ParamDelegateCleared"
	DelegateWriteEvent   = "ParamDelegateWrite"
)

// assertDelegate authorizes a caller as the holder of a delegated capability.
//
// want is the authorized package path, or "" when nothing is delegated.
//
// The empty check comes FIRST and is load-bearing, not defensive. A direct call
// from a user account has an empty previous package path — that is exactly what
// IsUserCall tests — so comparing against an unset slot would compare "" to ""
// and admit every user on the chain. r/gov/dao's own allowlist has the
// mirror-image bug (it returns true for everyone when the list is empty), kept
// there deliberately for genesis bootstrap. Nothing here needs that, so empty
// denies.
//
// Matching is exact, never a prefix. A sub-realm identity minted by
// cur.Sub(subpath) presents the synthesized path "host#subpath", so a delegate
// may legitimately be a single DAO hosted by a multi-tenant realm — for
// instance "gno.land/r/nt/commondao/v0#dao/42". An anchored-prefix match on the
// host would hand the capability to every DAO that realm hosts, and a bare
// prefix match would additionally match sibling packages. "#" cannot occur in a
// real package path, so exact matching on the full string is unambiguous.
//
// Takes rlm in the non-crossing dispatch position, matching assertValsetCaller
// so the two gates stay one shape.
// subject names what is being written, so the refusal stays specific to the
// capability rather than generic to the mechanism. params_valset_auth.txtar
// pins valset's exact wording as a regression test for PR #5485.
func assertDelegate(_ int, rlm realm, want, subject string) {
	if want == "" {
		panic("unauthorized: no delegate is configured for " + subject)
	}
	// The helper trusts its rlm input, so a future caller threading a stashed
	// or sibling-frame realm value would otherwise bypass the path check below.
	if !rlm.IsCurrent() {
		panic("unauthorized: rlm is not the caller's live cur")
	}
	if rlm.Previous().PkgPath() != want {
		panic("unauthorized: only " + want + " may write " + subject)
	}
}

// assertDelegatePath rejects package paths that must never be stored in a
// delegation slot, at the point a proposal is built rather than when it
// executes.
//
// Rejecting the empty string is the same fail-open guard as in assertDelegate,
// enforced on the way in as well so an accidental clear-by-empty-set cannot be
// mistaken for a grant. Requiring the gno.land/r/ prefix excludes user accounts
// (empty path), pure packages, and ephemeral `maketx run` realms — the last of
// which matters because a run realm's path is non-empty, so a check based on
// "is this code" would let one through.
func assertDelegatePath(pkgpath string) {
	if pkgpath == "" {
		panic("invalid delegate: empty package path")
	}
	if !isRealmPath(pkgpath) {
		panic("invalid delegate: must be a gno.land/r/ realm path, got " + pkgpath)
	}
	// Reject anything no caller could ever present, so a delegation cannot be
	// silently dead: GovDAO would believe it granted the capability while the
	// delegate is refused on every call. The sibling UpdateImpl rejects
	// whitespace-padded entries for the same reason.
	for i := 0; i < len(pkgpath); i++ {
		c := pkgpath[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '/' || c == '.' || c == '-' || c == '_' || c == '#':
		default:
			panic("invalid delegate: " + pkgpath +
				" contains a character no package path or sub-identity can hold")
		}
	}
	if n := countByte(pkgpath, '#'); n > 1 {
		panic("invalid delegate: more than one '#' in " + pkgpath)
	}
}

func countByte(s string, b byte) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			n++
		}
	}
	return n
}

const realmPathPrefix = "gno.land/r/"

func isRealmPath(pkgpath string) bool {
	return len(pkgpath) > len(realmPathPrefix) &&
		pkgpath[:len(realmPathPrefix)] == realmPathPrefix
}

// newDelegateProposal builds the GovDAO proposal that applies a delegation
// change. apply mutates the slot and is run only when the vote passes.
//
// event is emitted from inside the executor, so an observer sees the delegation
// change exactly when it takes effect rather than when it was proposed. from is
// the previous holder ("" when none), included so a re-delegation is auditable
// from the single event without reading prior state.
// apply returns the holder it replaced, read at EXECUTION time. Capturing that
// at creation would misreport a re-delegation: two set-proposals created while
// the slot is empty, then executed in sequence, would both emit from="" even
// though the second replaced the first.
func newDelegateProposal(cur realm, event, key, to, title, desc string, apply func() string) dao.ProposalRequest {
	callback := func(cur realm) error {
		from := apply()
		chain.Emit(event, "key", key, "from", from, "to", to)
		return nil
	}
	return dao.NewProposalRequest(title, desc,
		dao.NewSimpleExecutor(0, cur, callback, ""))
}
