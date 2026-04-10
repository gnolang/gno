# ADR: Account Sessions Support — Exploration 1 (PR #3970)

## Status

Draft — Superseded by #5307 (subaccount approach)

## Context

Gno accounts use a single key pair (`PubKey` + `Sequence` on
`BaseAccount`), creating several problems:

1. **Security**: A compromised key means total account loss. No way to
   delegate limited permissions without exposing the master key.
2. **Usability**: Every on-chain action requires the master key to sign.
   Interactive dApps (chat, games, social feeds) need dozens of
   confirmations per session — unusable with hardware wallets or
   keyring popups.
3. **Automation**: Can't grant a service, validator node, or agent
   limited access without sharing the master private key.
4. **Key rotation**: Rotating a compromised key requires moving all
   assets to a new address — there is no in-place rotation.

The goal is to let a master key authorize short-lived or limited-scope
*session keys* that can act on behalf of the account.

Related: [gnolang/gno#1499](https://github.com/gnolang/gno/issues/1499)

## Decision

### Core idea: restructure BaseAccount

Replace the single `PubKey`/`Sequence` on `BaseAccount` with a
`MasterKey` + `Sessions []AccountKey` model. Introduce `AccountKey` as
the unit of authentication identity:

```go
// tm2/pkg/std/account.go

type AccountKey interface {
    GetPubKey() crypto.PubKey
    SetPubKey(crypto.PubKey) error
    GetSequence() uint64
    SetSequence(uint64) error
}

type BaseAccount struct {
    Address       crypto.Address
    MasterKey     AccountKey     // root key (full control)
    RootSequence  uint64         // master key sequence
    Sessions      []AccountKey   // delegated keys
    Coins         Coins
    AccountNumber uint64
    SequenceSum   uint64         // total operations across all keys
}
```

The `Account` interface gains session management:

```go
type Account interface {
    // ... existing methods (Address, Coins, AccountNumber) ...
    GetSequenceSum() uint64
    SetSequenceSum(uint64) error
    SetMasterKey(crypto.PubKey) (AccountKey, error)
    GetMasterKey() AccountKey
    AddSession(pubKey crypto.PubKey) (AccountKey, error)
    DelSession(pubKey crypto.PubKey) error
    DelAllSessions() error
    GetKey(pubKey crypto.PubKey) (AccountKey, error)
    GetAllKeys() []AccountKey
}
```

### gno.land layer: GnoSession

`GnoAccount` embeds `BaseAccount` and adds `Sessions []GnoSession`:

```go
// gno.land/pkg/gnoland/types.go

type GnoSession struct {
    std.BaseAccountKey
    Flags                 BitSet    // permission flags
    ExpirationTime        time.Time
    CoinsTransferCapacity std.Coins // decreasing capacity
    RealmsWhitelist       []string  // glob patterns via filepath.Match
}
```

Permission flags (BitSet):
- `flagSessionUnlimitedTransferCapacity` — bypass coin transfer limits
- `flagSessionCanManageSessions` — create/revoke other sessions
- `flagSessionCanManagePackages` — deploy packages
- `flagSessionValidationOnly` — validator-only operations

### Key-Account relationship

One public key can control multiple accounts, and one account can have
multiple keys. This is safe because transactions always specify the
account address, so signatures are bound to a specific account context.
Each key has its own per-account sequence number.

### Sequence numbers

- Each `AccountKey` has an independent `Sequence` for replay protection.
- The account maintains `SequenceSum` — the total operations across all
  keys. New sessions start at the current `SequenceSum` to prevent
  replay attacks from previously revoked sessions.
- Sign bytes use `(AccountNumber, key.Sequence)` — not `SequenceSum`.

### AnteHandler flow

The AnteHandler is extended with a dual path:

1. **Modern path** (when `SignerInfo` is present): Resolves the signing
   key from the signature's `PubKey`, looks up the matching `AccountKey`
   on the account, uses the key's sequence for sign bytes verification.
2. **Legacy path** (genesis/backward compat): Falls back to
   `tx.GetSigners()` address-based resolution.

After signature verification, both the key's `Sequence` and the
account's `SequenceSum` are incremented.

### Realm access control

Sessions can be restricted to specific realm paths using
`filepath.Match` glob patterns on `RealmsWhitelist`:
- `"gno.land/r/demo/*"` — all sub-realms of demo
- `"gno.land/r/boards"` — exact match
- Empty whitelist = unrestricted access

### Transfer capacity

`CoinsTransferCapacity` is a decreasing balance: each transfer deducts
from it. When capacity reaches zero, further transfers are denied.
`flagSessionUnlimitedTransferCapacity` bypasses this check entirely.

### Expiration

Sessions have an `ExpirationTime`. Garbage collection (`gc()`) runs
lazily at the `gno.land` layer (not tm2), letting appchains customize
expiration logic.

## Consequences

### Positive

- Protocol-level sessions — no VM execution needed for auth.
- Flexible permission model with granular flags.
- One key controlling multiple accounts is a natural fit for agents
  and automated systems.
- Transfer capacity provides fine-grained spending control.
- `filepath.Match` patterns offer expressive realm restriction.

### Negative

- **Large BaseAccount blast radius**: Every consumer of the `Account`
  interface must be updated. `PubKey` and `Sequence` accessors change
  semantics. This is the primary reason this exploration was not merged.
- **SequenceSum interference**: Concurrent sessions incrementing
  `SequenceSum` creates ordering dependencies between unrelated signers.
- **`time.Now()` in consensus code**: The `gc()` function uses
  `time.Now()` for expiration checks, which is nondeterministic and
  would cause AppHash mismatches between validators.
- **`filepath.Match` surprises**: Glob semantics (`*` doesn't match
  `/`) can silently fail to match expected paths.
- **Session validation incomplete**: Many checks are marked `XXX`/`TODO`
  — expiry enforcement in AnteHandler, realm restriction enforcement,
  spend limit enforcement are not wired up.

### Neutral

- 64-session cap per account (arbitrary, can be tuned).
- Sessions stored inline on the account — simple but doesn't scale for
  enumeration or prefix-based revoke-all.

## Lessons learned

This exploration validated the core insight that protocol-level sessions
with independent sequences is the right approach. The main takeaways:

1. **Don't restructure BaseAccount** — the blast radius is too large.
   Separate subaccounts (as in #5307) avoid this entirely.
2. **Use `ctx.BlockTime()`** — never `time.Now()` in consensus code.
3. **Prefix matching > glob matching** — simpler, no surprises.
4. **Spend limits need periodic reset** — one-shot decreasing capacity
   is too rigid for real-world sessions.
5. **Flags vs message-type restriction** — the flag approach is more
   expressive but harder to reason about security; "exec-only" as a
   default is safer.

## References

- Issue: [gnolang/gno#1499](https://github.com/gnolang/gno/issues/1499)
  — Account Sessions System (Cookie-Like)
- Successor: [gnolang/gno#5307](https://github.com/gnolang/gno/pull/5307)
  — Session subaccounts (v4 design, by jaekwon)
- Related: [gnolang/gno#4218](https://github.com/gnolang/gno/pull/4218)
  — Account session (by notJoon)
