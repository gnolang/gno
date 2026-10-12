> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `pausable` - Emergency-stop pattern

Composes with `gno.land/p/nt/ownable/v0` to give an owner a programmatic kill
switch: flip `Pause` to halt sensitive operations, `Unpause` to resume.

## Usage

```go
package myrealm

import (
    "chain/runtime"

    "gno.land/p/nt/ownable/v0"
    "gno.land/p/nt/pausable/v0"
)

var p *pausable.Pausable

func init() {
    caller := runtime.PreviousRealm()
    if !caller.IsUserCall() {
        panic("must be deployed by a user")
    }
    p = pausable.NewFromOwnable(ownable.NewWithAddress(caller.Address()))
}

// Transfer is blocked while paused.
func Transfer(cur realm, to address, amount int64) error {
    if p.IsPaused() {
        return pausable.ErrPaused
    }
    // ... do the transfer
    return nil
}

// Only the owner can flip the switch.
func Pause(cur realm) error   { return p.Pause(0, cur) }
func Unpause(cur realm) error { return p.Unpause(0, cur) }
```

`Pausable` does not enforce anything by itself — callers must check `IsPaused()`
(or compare against `ErrPaused`) inside the functions that should be gated.

## API

```go
type Pausable struct{ /* unexported */ }

var ErrPaused = errors.New("pausable: realm is currently paused")

// NewFromOwnable wraps an existing Ownable. The Pausable starts unpaused.
func NewFromOwnable(o *ownable.Ownable) *Pausable

// State.
func (p Pausable) IsPaused() bool

// Owner-only mutations (thread the caller's own cur; pass 0 as the first arg).
// Return ownable.ErrUnauthorized if the caller is not the owner.
func (p *Pausable) Pause(_ int, rlm realm) error   // emits "Paused"   event with `by` = owner
func (p *Pausable) Unpause(_ int, rlm realm) error // emits "Unpaused" event with `by` = owner

// Access the underlying ownable (e.g. to transfer ownership).
func (p *Pausable) Ownable() *ownable.Ownable
```

## Notes

- Ownership is delegated to the embedded `*ownable.Ownable`. `Pause`/`Unpause`
  assert `rlm.IsCurrent()` and require `rlm.Previous().Address()` to equal the
  owner, so pass your own `cur` as `rlm`.
- Pausing is advisory: a function only respects the pause flag if it explicitly
  checks `IsPaused()`.
