package grc721

import (
	"chain"
)

// CallerTeller resolves the acting account at each write as rlm.Previous() —
// the realm that crossed into the realm holding the teller.
//
// SECURITY: this accessor hangs off *PrivateLedger, not *Token, and that is
// load-bearing. A frame-relative teller acts as whoever crossed into its
// holder, so it is only meaningful inside the token's own realm, where the
// wrappers act for a caller who knowingly invoked the collection. Anywhere
// else it is a confused deputy: a realm a user merely calls could move that
// user's tokens. The *Token pointer is published — exported vars, the
// collection facade, grc721reg — while the ledger is not, since NewToken hands
// it to the creating realm and nowhere else. So a foreign realm cannot mint
// one.
//
// Construction privacy alone is not enough: a realm may legally build a teller
// and then export the VALUE. The write methods therefore also verify that the
// invoking realm is the token's own (see guardHome), which leaves a leaked
// teller inert everywhere but home.
func (led *PrivateLedger) CallerTeller() Teller {
	if led == nil {
		panic("Ledger cannot be nil")
	}

	return &fnTeller{
		accountFn: func(_ int, rlm realm) address {
			return rlm.Previous().Address()
		},
		homeGuard: true,
		Token:     led.token,
	}
}

// Write methods return ErrReadonly.
func (tok *Token) ReadonlyTeller() Teller {
	if tok == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: nil,
		Token:     tok,
	}
}

// Permanently acts as the calling realm (verified via IsCurrent).
func (tok *Token) RealmTeller(_ int, rlm realm) Teller {
	if tok == nil {
		panic("Token cannot be nil")
	}

	if !rlm.IsCurrent() {
		panic(ErrSpoofedRealm)
	}

	caller := rlm.Address()

	return &fnTeller{
		accountFn: func(_ int, _ realm) address {
			return caller
		},
		Token: tok,
	}
}

// Like RealmTeller but acts as a sub-account derived from slug.
func (tok *Token) RealmSubTeller(_ int, rlm realm, slug string) Teller {
	if tok == nil {
		panic("Token cannot be nil")
	}

	if !rlm.IsCurrent() {
		panic(ErrSpoofedRealm)
	}

	account := accountSlugAddr(rlm.Address(), slug)

	return &fnTeller{
		accountFn: func(_ int, _ realm) address {
			return account
		},
		Token: tok,
	}
}

// Admin-grade: issuer-only (holds the PrivateLedger). Panics on invalid addr to close the empty-sentinel hole.
func (led *PrivateLedger) ImpersonateTeller(addr address) Teller {
	if led == nil {
		panic("Ledger cannot be nil")
	}

	if !addr.IsValid() {
		panic(ErrInvalidAddress)
	}

	return &fnTeller{
		accountFn: func(_ int, _ realm) address {
			return addr
		},
		Token: led.token,
	}
}

// guardHome confines a frame-relative teller to the token's own realm.
// Construction privacy stops a foreign realm from minting one; this stops a
// minted one from travelling, which is what happens when a realm legally builds
// a teller and then exports the value.
//
// The check is on the invoking realm's path alone — deliberately NOT on whether
// the resolved actor is an end user. Keying on the actor only blocks the case
// where the debited party is the signing user, and leaves two doors open: a
// realm can be charged by a realm it calls, and TransferFrom resolves the
// *spender* from the frame, so a realm reached from an honest hub spends that
// hub's approvals against any owner who granted one. Both are the same defect
// as the original one level up — frame-relative resolution means whoever you
// call can act as you — and neither is reachable once the teller only works at
// home.
//
// The host is compared after stripping any ":subpath" synthesized by realm.Sub,
// so the token's own sub-realms are not falsely rejected.
//
// A foreign realm that needs to move a user's tokens uses the ordinary route:
// the owner Approves it, and it spends as itself through RealmTeller, which is
// eagerly bound to its own address and approval-gated.
//
// The leading int keeps this a plain method.
func (ft *fnTeller) guardHome(_ int, rlm realm) error {
	if !ft.homeGuard {
		return nil
	}

	host, _, _ := chain.SplitPkgSubPath(rlm.PkgPath())
	if host != ft.Token.origRealm {
		return ErrForeignCallerTeller
	}

	return nil
}

func (ft *fnTeller) Approve(_ int, rlm realm, to address, tid TokenID) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}

	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}

	if err := ft.guardHome(0, rlm); err != nil {
		return err
	}

	caller := ft.accountFn(0, rlm)

	return ft.Token.ledger.Approve(caller, to, tid)
}

func (ft *fnTeller) SetApprovalForAll(_ int, rlm realm, operator address, approved bool) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}

	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}

	if err := ft.guardHome(0, rlm); err != nil {
		return err
	}

	caller := ft.accountFn(0, rlm)

	return ft.Token.ledger.SetApprovalForAll(caller, operator, approved)
}

func (ft *fnTeller) TransferFrom(_ int, rlm realm, from, to address, tid TokenID) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}

	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}

	if err := ft.guardHome(0, rlm); err != nil {
		return err
	}

	spender := ft.accountFn(0, rlm)

	return ft.Token.ledger.TransferFrom(spender, from, to, tid)
}

func accountSlugAddr(addr address, slug string) address {
	if slug == "" {
		return addr
	}

	key := addr.String() + "/" + slug

	return chain.PackageAddress(key)
}
