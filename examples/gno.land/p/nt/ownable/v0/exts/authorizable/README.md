> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `authorizable` - Second authorization tier over ownable

Extension of [`gno.land/p/nt/ownable/v0`](../..) that adds a second permission level on top of single-owner ownership: one **superuser** (the `ownable` owner) plus a list of **authorized** addresses. Use it for a moderator tier, an allowlist, or any "owner, plus a set of trusted others" pattern.

## Usage

```go
package myrealm

import (
    "chain/runtime"

    "gno.land/p/nt/ownable/v0"
    "gno.land/p/nt/ownable/v0/exts/authorizable"
)

// The superuser (and first entry on the auth list) is chosen explicitly.
// Here: the deployer, captured in init.
var auth *authorizable.Authorizable

func init() {
    caller := runtime.PreviousRealm()
    if !caller.IsUserCall() {
        panic("must be deployed by a user")
    }
    auth = authorizable.New(ownable.NewWithAddress(caller.Address()))
}

// Superuser-only: add a moderator.
func AddModerator(cur realm, addr address) error {
    return auth.AddToAuthList(0, cur, addr)
}

// Gate an action to anyone on the auth list.
func Moderate(cur realm) {
    auth.AssertPreviousOnAuthList(0, cur)
    // ... privileged work ...
}
```

## API

```go
type Authorizable struct {
    *ownable.Ownable // the owner is the superuser; all Ownable methods are inherited
    // unexported auth list
}

// New builds an Authorizable from an existing *ownable.Ownable.
// The owner is automatically added to the auth list.
func New(o *ownable.Ownable) *Authorizable

// Superuser-only (previous caller must be the owner).
func (a *Authorizable) AddToAuthList(_ int, rlm realm, addr address) error
func (a *Authorizable) DeleteFromAuthList(_ int, rlm realm, addr address) error

// Membership checks (return an error; nil means on the list).
func (a *Authorizable) OnAuthList(_ int, rlm realm) error         // is the caller realm itself on the list
func (a *Authorizable) PreviousOnAuthList(_ int, rlm realm) error // is the realm/user that crossed in on the list

// Assert variants panic instead of returning an error.
func (a Authorizable) AssertOnAuthList(_ int, rlm realm)
func (a Authorizable) AssertPreviousOnAuthList(_ int, rlm realm)

// Errors: ErrNotSuperuser, ErrNotInAuthList, ErrAlreadyInList
```

## Notes

- Every method takes the caller's own captured `cur` as `rlm` and asserts `rlm.IsCurrent()`, blocking the designation-forgery read where a non-crossing wrapper makes the realm walk return the wrong address. The first `_ int` argument is an unused placeholder: pass `0`.
- The superuser is authenticated by `rlm.Previous().Address()` matching the underlying `Ownable` owner, so `AddToAuthList` / `DeleteFromAuthList` succeed only when the owner is the crossing caller. Ownership transfer, renouncing, etc. come from the embedded [`Ownable`](../..).
- `PreviousOnAuthList` / `AssertPreviousOnAuthList` are the user-facing gate: they check the address that crossed into your realm. `OnAuthList` checks the calling realm itself; use it only when a realm-to-realm caller should be listed directly.
- The auth list is backed by a [`bptree`](../../../../bptree/v0), keyed by address string.
