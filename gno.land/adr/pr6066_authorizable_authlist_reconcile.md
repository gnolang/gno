# ADR: `authorizable` — keep the auth list in sync with ownership transfer/renounce

## Context

`p/nt/ownable/v0/exts/authorizable` adds a second authorization tier over
`p/nt/ownable/v0`: one superuser (the `Ownable` owner) plus an `authorized`
`bptree` of member addresses. `New` seeds the auth list with the owner, and the
README documents the invariant "The owner is automatically added to the auth
list."

`Authorizable` embeds `*ownable.Ownable`, so it inherits the promoted
`TransferOwnership`/`DropOwnership` mutators. Those methods only mutate
`Ownable.owner`; they never touch the `authorized` tree. After an ownership
transfer the list therefore desyncs:

- the **new owner is not on the auth list**, so a consumer realm gating an
  action on `AssertPreviousOnAuthList` rejects the legitimate new owner; and
- the **deposed owner stays on the auth list** and keeps whatever
  "authorized" (moderator) privileges the list grants until manually removed.

This inverts the documented invariant and leaves a stale grant after a
security-relevant ownership rotation. It is not remotely exploitable by a third
party (`TransferOwnership` requires `rlm.Previous().Address() == owner`), but it
is a real authorization-consistency defect for any consumer that exposes
ownership transfer alongside an `AssertPreviousOnAuthList` gate (the package's
own `Moderate` example in the README).

## Decision

Override the two ownership mutators on `*Authorizable` so the auth list is
reconciled with ownership, delegating the superuser guard to the embedded
`Ownable`:

- `TransferOwnership(_ int, rlm realm, newOwner address) error` — calls the
  embedded `Ownable.TransferOwnership` (which enforces `rlm.IsCurrent()` +
  `rlm.Previous().Address() == owner` and validates `newOwner`), then removes
  the old owner from `authorized` and adds the new owner.
- `DropOwnership(_ int, rlm realm) error` — calls the embedded
  `Ownable.DropOwnership`, then removes the old owner from `authorized`.

Removal uses `bptree.Remove`, which is a safe no-op when the key is absent (the
owner may have previously removed themselves from the list while staying
owner), so no extra existence check is needed. Member addresses added via
`AddToAuthList` (moderators) are untouched.

## Alternatives considered

- **Add-only (add the new owner, keep the old owner)**: fixes the locked-out
  new owner but deliberately preserves the stale grant — the core
  authorization concern. Rejected: a deposed owner can be re-added explicitly
  via `AddToAuthList` by the new owner if continued membership is intended.
- **Hook/callback in `ownable`**: a more invasive change to the base package
  with a larger blast radius and no current in-repo need. Rejected for this fix.
- **Document-only**: state that consumers must reconcile manually. Rejected —
  it leaves the invariant broken and the footgun in place; reconciliation is
  the behavior the "owner is automatically added" invariant implies.

## Consequences

- After `TransferOwnership`, the auth list contains the new owner and no longer
  contains the deposed owner; after `DropOwnership`, the deposed owner is
  removed. The "owner is always on the auth list" invariant now holds across
  the ownership lifecycle, not just at construction.
- Explicitly-added member addresses are unaffected by ownership changes.
- The exported `*Authorizable` handle gains no new write-authority exposure:
  both overrides delegate the same `rlm.IsCurrent()` + `OwnedBy(caller)` guard
  as the embedded `Ownable`, matching the package's existing caller-guarded
  mutator design.
- Regression tests: `TestTransferOwnershipReconcilesAuthList` and
  `TestDropOwnershipRemovesOwnerFromAuthList` pin the reconcile behavior and
  the `PreviousOnAuthList` user-facing gate; the `p/nt/ownable/v0` module and
  the `r/gnops/valopers` consumer tests pass.
