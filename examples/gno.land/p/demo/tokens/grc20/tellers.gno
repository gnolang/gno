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
		host:  ledger.token.origRealm,
		Token: ledger.token,
	}
}

// UserTeller acts for direct EOA callers of hostPath. Use the token realm's
// path for local entry points or a trusted relay's path to delegate to it.
// The host is fixed at construction, includes its sub-realms, and cannot be
// empty. A leaked teller cannot write from any other host.
func (ledger *PrivateLedger) UserTeller(hostPath string) Teller {
	if ledger == nil {
		panic("Ledger cannot be nil")
	}
	host, _, _ := chain.SplitPkgSubPath(hostPath)
	if host == "" {
		panic("UserTeller host cannot be empty")
	}

	return &fnTeller{
		accountFn: func(_ int, rlm realm) address {
			return rlm.Previous().Address()
		},
		host:     host,
		userOnly: true,
		Token:    ledger.token,
	}
}

// IsCanonicalUserTeller accepts only a UserTeller for this exact token and host.
// Check before dispatching any methods on an externally supplied Teller.
func IsCanonicalUserTeller(teller Teller, token *Token, hostPath string) bool {
	ft, ok := teller.(*fnTeller)
	host, _, _ := chain.SplitPkgSubPath(hostPath)
	return ok && ft != nil && token != nil && ft.Token == token &&
		ft.accountFn != nil && ft.userOnly && host != "" && ft.host == host
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

// guardWrite confines frame-relative tellers to their designated host and its
// sub-realms. CallerTeller is bound to the token origin; UserTeller also rejects
// callers other than direct MsgCall EOAs. Checking the host prevents a leaked
// teller from debiting another host's callers or spending their allowances.
// Transfer, Approve and TransferFrom verify IsCurrent before calling this.
func (ft *fnTeller) guardWrite(_ int, rlm realm) error {
	if ft.userOnly && !rlm.Previous().IsUserCall() {
		return ErrNotUserCall
	}
	if ft.host == "" {
		return nil
	}
	host, _, _ := chain.SplitPkgSubPath(rlm.PkgPath())
	if host != ft.host {
		if ft.userOnly {
			return ErrForeignUserTeller
		}
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
