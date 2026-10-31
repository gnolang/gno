> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `ownable` - Ownership pattern for realms

Provides an `Ownable` object that gates privileged operations behind a single owner address. Embed it in a realm (or any struct) to restrict actions like configuration changes, withdrawals, or upgrades.

## Usage

```go
package myrealm

import (
    "chain/runtime"

    "gno.land/p/nt/ownable/v0"
)

// The owner address is chosen explicitly at construction. A common
// choice is the deployer, captured in init after confirming it is a
// real user call.
var owner *ownable.Ownable

func init() {
    caller := runtime.PreviousRealm()
    if !caller.IsUserCall() {
        panic("must be deployed by a user")
    }
    owner = ownable.NewWithAddress(caller.Address())
}

// SetFee is gated: only the current owner may call it.
func SetFee(cur realm, newFee int64) {
    if !cur.IsCurrent() {
        panic("spoofed realm")
    }
    owner.AssertOwnedBy(cur.Previous().Address())
    fee = newFee
}

// Hand the realm over. TransferOwnership itself verifies the caller is owner.
func TransferOwner(cur realm, to address) error {
    return owner.TransferOwnership(0, cur, to)
}
```

There is no auth-mode flag. The single `NewWithAddress` constructor replaced the
old `New` / `NewWithOrigin` / `NewWithAddressByPrevious` sugar: the realm now picks
the owner address explicitly rather than baking a runtime walk into the struct.

## API

```go
type Ownable struct{ /* unexported */ }

const OwnershipTransferEvent = "OwnershipTransfer"

var (
    ErrUnauthorized   = errors.New("ownable: caller is not owner")
    ErrInvalidAddress = errors.New("ownable: new owner address is invalid")
)

// NewWithAddress is the only constructor: the realm picks the owner
// address explicitly (e.g. cur.Previous().Address() after checking
// cur.Previous().IsUserCall() in init).
func NewWithAddress(addr address) *Ownable

// Queries (caller supplies the address to check).
func (o *Ownable) Owner() address             // "" if o is nil or ownership was dropped
func (o *Ownable) OwnedBy(addr address) bool  // true if addr is the current owner
func (o *Ownable) AssertOwnedBy(addr address) // panics with ErrUnauthorized if addr is not the owner

// Authority mutation (thread the caller's own cur; pass 0 as the first arg).
func (o *Ownable) TransferOwnership(_ int, rlm realm, newOwner address) error
func (o *Ownable) DropOwnership(_ int, rlm realm) error // sets owner to "" — irreversible
```

## Notes

- Authority-mutating methods assert `rlm.IsCurrent()` and identify the caller as `rlm.Previous().Address()`, which must equal the current owner. The principal is therefore unforgeable: an attacker cannot supply an arbitrary caller address. Pass `0` as the placeholder first arg and your own `cur` as `rlm`.
- Read helpers (`OwnedBy`, `AssertOwnedBy`) take a bare address; the caller extracts it, guarding with `cur.IsCurrent()` before reading `cur.Previous().Address()`.
- `TransferOwnership` rejects an invalid `newOwner` with `ErrInvalidAddress`. Both mutators emit `OwnershipTransferEvent` with `from` and `to` fields.
- `DropOwnership` is permanent: `owner` becomes `""`, so every owner-gated action becomes unreachable.
