package grc20

import (
	"chain"
)

// CallerTeller returns a GRC20 compatible teller that, at each write call,
// resolves the caller as rlm.Previous() — the realm that crossed into the
// caller. rlm must be the caller's own captured cur (asserted via
// rlm.IsCurrent() inside the Teller methods).
//
// SECURITY: this accessor hangs off *PrivateLedger, not *Token, and that is
// load-bearing. A frame-relative teller debits whoever crossed into the realm
// holding it, so it is only ever meaningful inside the token's own realm,
// whose wrappers act for a caller who knowingly invoked the token. Anywhere
// else it is a confused deputy. The *Token pointer is published (exported
// vars, grc20factory, grc20reg) but the ledger is not — NewToken hands it to
// the creating realm and nowhere else — so a foreign realm cannot mint one.
//
// Construction privacy alone is not enough: a realm may legally build a teller
// and then export the VALUE. The write methods therefore also verify that the
// invoking realm is the token's own (see guardWrite), which makes a leaked
// teller inert everywhere but home.
func (ledger *PrivateLedger) CallerTeller() Teller {
	if ledger == nil {
		panic("Ledger cannot be nil")
	}

	return &fnTeller{
		accountFn: func(_ int, rlm realm) address {
			return rlm.Previous().Address()
		},
		homeGuard: true,
		Token:     ledger.token,
	}
}

// UserTeller returns a Teller that acts on behalf of the immediate EOA caller.
//
// It is home-guarded and rejects non-user calls, limiting writes to MsgCall
// interactions originating from a signing user.
func (ledger *PrivateLedger) UserTeller() Teller {
	if ledger == nil {
		panic("Ledger cannot be nil")
	}

	return &fnTeller{
		accountFn: func(_ int, rlm realm) address {
			return rlm.Previous().Address()
		},
		homeGuard: true,
		userOnly:  true,
		Token:     ledger.token,
	}
}

// ReadonlyTeller is a GRC20 compatible teller that panics for any write operation.
func (tok *Token) ReadonlyTeller() Teller {
	if tok == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: nil,
		Token:     tok,
	}
}

// RealmTeller returns a GRC20 compatible teller that will store the
// caller realm permanently. Calling anything through this teller will
// result in allowance or balance changes for the realm that initialized the teller.
// The initializer of this teller should usually never share the resulting Teller from
// this method except maybe for advanced delegation flows such as a DAO treasury
// management.
//
// rlm must be the caller's own captured cur (asserted via rlm.IsCurrent()).
// The address is frozen eagerly at construction.
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

// RealmSubTeller is like RealmTeller but uses the provided slug to derive a
// subaccount.
//
// rlm must be the caller's own captured cur (asserted via rlm.IsCurrent()).
// The subaccount address is frozen eagerly at construction.
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

// ImpersonateTeller returns a GRC20 compatible teller that impersonates as a
// specified address. This allows operations to be performed as if they were
// executed by the given address, enabling the caller to manipulate tokens on
// behalf of that address.
//
// It is particularly useful in scenarios where a contract needs to perform
// actions on behalf of a user or another account, without exposing the
// underlying logic or requiring direct access to the user's account. The
// returned teller will use the provided address for all operations, effectively
// masking the original caller.
//
// This method should be used with caution, as it allows for potentially
// sensitive operations to be performed under the guise of another address.
func (ledger *PrivateLedger) ImpersonateTeller(addr address) Teller {
	if ledger == nil {
		panic("Ledger cannot be nil")
	}

	return &fnTeller{
		accountFn: func(_ int, _ realm) address {
			return addr
		},
		Token: ledger.token,
	}
}

// generic tellers methods.
//

// guardWrite confines a frame-relative (homeGuard) teller to the token's own
// realm. Construction privacy stops a foreign realm from minting one;
// this stops a minted one from travelling, which is what happens when a realm
// legally builds a teller and then exports the value.
//
// The check is on the invoking realm's path alone — deliberately NOT on
// whether the resolved actor is an end user. Keying on the actor only blocks
// the case where the debited party is the signing user, which leaves two doors
// open: a realm can be charged by a realm it calls, and TransferFrom resolves
// the *spender* from the frame, so a realm reached from an honest hub spends
// that hub's allowance against any owner who granted one. Both are the same
// defect as the original one level up — frame-relative resolution means
// whoever you call can act as you — and neither is reachable once the teller
// only works at home.
//
// The host is compared after stripping any ":subpath" synthesized by
// realm.Sub, so the token's own sub-realms are not falsely rejected.
//
// A foreign realm that needs to move a user's funds uses the ordinary route:
// Approve, then RealmTeller().TransferFrom, which is eagerly bound to that
// realm's own address and allowance-gated.
//
// A userOnly teller adds one further condition: the immediate caller must be an
// EOA reached through MsgCall, so a realm that cross-calls the surface function
// is refused rather than served as itself.
//
// The leading int keeps this a plain method.
func (ft *fnTeller) guardWrite(_ int, rlm realm) error {
	if ft.userOnly && !rlm.Previous().IsUserCall() {
		return ErrNotUserCall
	}
	if !ft.homeGuard {
		return nil
	}
	host, _, _ := chain.SplitPkgSubPath(rlm.PkgPath())
	if host != ft.Token.origRealm {
		return ErrForeignCallerTeller
	}
	return nil
}

func (ft *fnTeller) Transfer(_ int, rlm realm, to address, amount int64) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}
	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}
	if err := ft.guardWrite(0, rlm); err != nil {
		return err
	}
	caller := ft.accountFn(0, rlm)
	return ft.Token.ledger.Transfer(caller, to, amount)
}

func (ft *fnTeller) Approve(_ int, rlm realm, spender address, amount int64) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}
	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}
	if err := ft.guardWrite(0, rlm); err != nil {
		return err
	}
	caller := ft.accountFn(0, rlm)
	return ft.Token.ledger.Approve(caller, spender, amount)
}

func (ft *fnTeller) TransferFrom(_ int, rlm realm, owner, to address, amount int64) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}
	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}
	if err := ft.guardWrite(0, rlm); err != nil {
		return err
	}
	spender := ft.accountFn(0, rlm)
	return ft.Token.ledger.TransferFrom(owner, spender, to, amount)
}

// helpers
//

// accountSlugAddr returns the address derived from the specified address and slug.
func accountSlugAddr(addr address, slug string) address {
	// XXX: use a new `std.XXX` call for this.
	if slug == "" {
		return addr
	}
	key := addr.String() + "/" + slug
	return chain.PackageAddress(key) // temporarily using this helper
}
