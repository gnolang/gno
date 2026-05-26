package grc721

import (
	"chain"
)

// CallerTeller returns a GRC721 compatible teller that, at each write call,
// resolves the caller as rlm.Previous() — the realm that crossed into the
// caller. rlm must be the caller's own captured cur (asserted via
// rlm.IsCurrent() inside the Teller methods).
func (tok *Token) CallerTeller() Teller {
	if tok == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: func(_ int, rlm realm) address {
			return rlm.Previous().Address()
		},
		Token: tok,
	}
}

// ReadonlyTeller is a GRC721 compatible teller that returns ErrReadonly for
// any write operation.
func (tok *Token) ReadonlyTeller() Teller {
	if tok == nil {
		panic("Token cannot be nil")
	}

	return &fnTeller{
		accountFn: nil,
		Token:     tok,
	}
}

// RealmTeller returns a GRC721 compatible teller that will store the
// caller realm permanently. Calling anything through this teller will
// result in approval or ownership changes for the realm that initialized
// the teller. The initializer of this teller should usually never share
// the resulting Teller from this method except maybe for advanced
// delegation flows such as a DAO treasury management.
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

// ImpersonateTeller returns a GRC721 compatible teller that impersonates as a
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

// generic tellers methods.
//

// TransferFrom moves token tid from `from` to `to`. The caller must be the
// owner of tid, an approved operator for the owner, or the address approved
// for tid.
//
// EIP-721's safeTransferFrom is intentionally not exposed: it would require
// a registry-based ERC721TokenReceiver check (see
// https://github.com/gnolang/gno/issues/5546). Implementing it as a
// TransferFrom alias would leak a false safety guarantee.
func (ft *fnTeller) TransferFrom(_ int, rlm realm, from, to address, tid TokenID) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}

	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}

	caller := ft.accountFn(0, rlm)
	if !ft.isApprovedOrOwner(caller, tid) {
		return ErrCallerIsNotOwnerOrApproved
	}

	return ft.Token.ledger.Transfer(from, to, tid)
}

func (ft *fnTeller) Approve(_ int, rlm realm, to address, tid TokenID) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}

	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}

	caller := ft.accountFn(0, rlm)

	owner, err := ft.Token.OwnerOf(tid)
	if err != nil {
		return err
	}

	if caller != owner && !ft.Token.IsApprovedForAll(owner, caller) {
		return ErrCallerIsNotOwnerOrApproved
	}

	return ft.Token.ledger.Approve(to, tid)
}

func (ft *fnTeller) SetApprovalForAll(_ int, rlm realm, op address, approved bool) error {
	if ft.accountFn == nil {
		return ErrReadonly
	}

	if !rlm.IsCurrent() {
		return ErrSpoofedRealm
	}

	caller := ft.accountFn(0, rlm)

	return ft.Token.ledger.SetApprovalForAll(caller, op, approved)
}

// isApprovedOrOwner reports whether addr is allowed to act on tid:
// either it owns tid, the owner has approved-for-all addr, or addr is
// the per-token approved address for tid.
func (ft *fnTeller) isApprovedOrOwner(addr address, tid TokenID) bool {
	owner, err := ft.Token.OwnerOf(tid)
	if err != nil {
		return false
	}

	if addr == owner || ft.Token.IsApprovedForAll(owner, addr) {
		return true
	}

	approved, err := ft.Token.GetApproved(tid)
	if err != nil {
		return false
	}

	return approved == addr
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
