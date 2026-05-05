# ADR-001: Session Subaccounts (gno.land Layer)

## Status

Accepted

## Context

This ADR covers the gno.land-specific aspects of session subaccounts.
The core protocol (tm2 layer) is described in `tm2/adr/adr-002-session-subaccounts.md`.

Session subaccounts allow users to authorize limited-capability signing
keys for dApps. The tm2 layer handles authentication (signature
verification, expiry, spend limits). The gno.land layer adds
authorization: which message types and realm paths a session can use.

## Decision

### GnoSessionAccount

```go
type GnoSessionAccount struct {
    std.BaseSessionAccount
    AllowPaths []string  // realm path prefixes; empty = unrestricted
}
```

`AllowPaths` restricts which realm packages the session can call.
Path matching uses exact match or sub-path:
`path == prefix || strings.HasPrefix(path, prefix+"/")`.
This prevents `gno.land/r/demo` from matching `gno.land/r/demo_evil`.

### Ante wrapper: checkSessionRestrictions

After the tm2 AnteHandler passes, a gno.land-specific check runs:

1. **Message type restriction.** Session-signed messages must have
   `msg.Type()` in the allowlist:
   - `"exec"` (MsgCall)
   - `"run"`  (MsgRun)
   - `"send"` (bank.MsgSend)
   - `"multisend"` (bank.MsgMultiSend)

   MsgAddPackage and all auth msgs (create_session, revoke_session,
   revoke_all_sessions) are denied. This check uses the
   `DelegatedAccount` presence in the context map, not a type assertion
   to `*GnoSessionAccount`, so it cannot be bypassed by alternative
   session account types.

2. **AllowPaths restriction.** If `AllowPaths` is non-empty, the
   msg's `PkgPath` must match one of the allowed prefixes. Read via a
   local `pathRestricted` interface to avoid tm2 importing gno.land
   types. Path-less msgs — MsgRun, MsgSend, and MsgMultiSend — do not
   implement `pkgPather`, so a session with non-empty `AllowPaths`
   rejects them. This is intentional: a session with AllowPaths set
   is realm-scoped, and path-less msgs (arbitrary code execution,
   direct value transfers) would escape that scope.

### AllowPaths validation at creation

`handleMsgCreateSession` validates AllowPaths entries:
- Max `MaxAllowPathsPerSession` (8) entries
- No empty strings
- No trailing slashes (would silently make paths unreachable)

### Spend limit enforcement

`SpendLimit` is **authoritative** across every tx-initiated outflow
from master's balance. Enforcement happens at the tm2 bank keeper:

- `bank.Keeper.SendCoins` calls `auth.CheckAndDeductSessionSpend`
  after its `canSendCoins` check.
- `bank.Keeper.InputOutputCoins` does the same per input (for
  `MsgMultiSend`).

This single gate covers:

- Outer `msg.Send` on MsgCall / MsgRun (VM keeper routes the transfer
  through `bank.Keeper.SendCoins`).
- `bank.MsgSend` and `bank.MsgMultiSend`.
- In-realm `std.Send` / `banker.SendCoins` from gno code running inside
  a session-signed MsgCall or MsgRun.
- Storage deposits (see below).

Deliberate bypasses (via `bank.Keeper.SendCoinsUnrestricted`):

- Gas fee collection in the ante handler.
- Storage deposit refunds in `refundStorageDeposit` (a credit to
  master, not a debit).

Storage deposits get an explicit `CheckAndDeductSessionSpend` in
`VMKeeper.lockStorageDeposit` before the `SendCoinsUnrestricted`
transfer — the transfer itself needs to bypass restricted-denom
checks, but session accounting still applies. Refunds do **not**
reverse `SpendUsed`: a session's cumulative spend is a high-water
mark, not a net balance, so a compromised session cannot churn state
(allocate → free → allocate) to drain master beyond `SpendLimit`.
The refunded coins still reach master; only the session's remaining
budget stays at its high-water mark.

`VMKeeper.AddPackage()` retains a redundant pre-check as
defense-in-depth. AddPackage is currently blocked from session signers
at the gno.land allowlist, so the pre-check never fires for sessions
today. If the allowlist is ever relaxed, the pre-check must be
removed or converted to a check-only variant to avoid double-counting
with the bank-keeper hook.

Sessions that intend to spend the chain's gas / storage denom
(`ugnot`) must include it in `SpendLimit` — any outflow whose denom
is absent from `SpendLimit` fails. See `tm2/adr/adr-002` for the gas
fee denom note that applies here as well.

### VM integration

`ExecContext.SessionAccount` carries the `std.DelegatedAccount`
directly. The Gno stdlib exposes it via:

```gno
// chain/runtime
func GetSessionInfo() (pubKeyAddr address, expiresAt int64, allowPaths []string, isSession bool)
```

Return values:
- `isSession == false`: not a session tx (master key signed)
- `isSession == true, len(allowPaths) == 0`: session with no path restrictions
  (nil or empty slice — both mean unrestricted)
- `isSession == true, len(allowPaths) > 0`: session restricted to these paths

`pubKeyAddr` is the session key's address (for per-session tracking).
The master address is available via `OriginCaller()` as usual.

`SpendLimit` is enforced at the protocol level for all coin
movements, so realms **do not** need to check session status for
basic spending bounds. Realms that want to apply additional
business-logic restrictions beyond `SpendLimit` (e.g., rate limits on
trade volume, different limits for session vs master) can check
`isSession`:

```gno
func Trade(cur realm, pair string, amount int64) {
    _, _, _, isSession := runtime.GetSessionInfo()
    if isSession {
        // optional: enforce realm-specific trade limits for sessions
    }
}
```

Note that `banker.RemoveCoin` cannot reach master's native coins:
the gno-side banker at `gnovm/stdlibs/chain/banker/banker.gno`
requires `BankerTypeRealmIssue` and enforces that the denom is
prefixed with the current realm's path. Master's `ugnot` has no such
prefix, so no realm can burn it.

### Session prototype

`AccountKeeper` accepts a `sessionProto` function:

```go
acck := auth.NewAccountKeeper(mainKey, prmk, ProtoGnoAccount, ProtoGnoSessionAccount)
```

The tm2 handler creates sessions via `acck.NewSessionAccount()` and
sets `AllowPaths` through a local `allowPathsSetter` interface, keeping
tm2 unaware of AllowPaths as a concept.

## Consequences

### Positive

- AllowPaths is purely a gno.land concern — tm2 doesn't know about it.
- Message type restriction is fail-safe: new message types are denied
  by default.
- Realms can opt into session awareness via `GetSessionInfo()` without
  any changes for realms that don't care about sessions.

### Negative

- `checkSessionRestrictions` runs on every tx (but short-circuits when
  no sessions are in context).
- AllowPaths entries are not validated as realm-like paths — garbage
  entries simply fail to match anything (fail-closed).

### Neutral

- Unrestricted sessions (empty AllowPaths) can call any realm.
  **Device login is the primary use case for this mode**: a user
  creates a session on a less-secure device (browser, mobile app),
  leaves `AllowPaths` empty, and relies on time expiry + `SpendLimit`
  to bound damage if the device is compromised. Because `SpendLimit`
  is authoritative across all outflows (see "Spend limit enforcement"
  above), this is safe even when the session can call arbitrary
  attacker-deployed realms.
- `MaxSessionsPerAccount` (16), `MaxAllowPathsPerSession` (8), and
  `MaxSessionDuration` (30 days) are compile-time constants in
  `tm2/pkg/std/account.go`, not tunable params. Changing them
  requires a coordinated upgrade of all nodes.

## References

- tm2 ADR: `tm2/adr/adr-002-session-subaccounts.md`
- Issue gnolang/gno#1499: Account Sessions System (Cookie-Like)
