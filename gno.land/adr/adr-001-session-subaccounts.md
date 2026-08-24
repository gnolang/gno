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
    AllowPaths []string  // typed entries; required at create-time
}
```

`AllowPaths` is a **required** per-session msg-type allow-list using a
typed grammar. Empty `AllowPaths` is rejected at both the CLI and
chain layers — every session must make an explicit choice between the
wildcard and a specific entry list.

The grammar:

```
entry      := "*" | route_type [":" path]
route_type ∈ {"vm/exec", "vm/run", "bank/send", "bank/multisend"}
path        — non-empty, no trailing slash, only legal after "vm/exec:"
```

The `*` wildcard matches any msg type, **subject to the always-denied
list below**. It's the deliberate "trust this session like a master
except for privilege-escalation msgs" form, suitable for device login.

The four canonical `<route>/<type>` pairs are the only fully-qualified
forms accepted today. Bare-route entries (e.g. `bank` to mean any bank
msg) and unknown types are rejected. A future relaxation may permit
bare-route entries via a fork; today's sessions never use that form.

Only `vm/exec` may carry an optional `:<path>` suffix that restricts
the call target by realm prefix:

- `vm/exec:gno.land/r/jae/blog` matches `MsgCall` to that realm or any
  sub-realm.
- `vm/exec` (no path) matches any `MsgCall`, regardless of realm.
- `*:<path>` is rejected — the wildcard does not accept a path.

Path matching preserves the prefix-attack guard:
`path == prefix || strings.HasPrefix(path, prefix+"/")`. This prevents
`gno.land/r/demo` from matching `gno.land/r/demo_evil`.

### Always-denied list

Two filters apply at ante-time, in order:

1. **Privilege-escalation deny-list** (hard floor, regardless of
   AllowPaths or `*`):
   - any `auth/*` msg — covers `create_session`, `revoke_session`,
     `revoke_all_sessions`, and any future auth msgs by route.
   - `vm/add_package` — sessions cannot claim namespace under master.

2. **AllowPaths match** — a msg passes if any entry matches its
   `(Route(), Type()[, PkgPath()])`. The wildcard `*` matches anything
   (after step 1).

The deny-list is forward-compatible: a future fork that adds a new
auth-module msg type defaults to "denied" (because the rule denies the
whole `auth/` route). New `vm/*` types default to allowed and may be
granted via AllowPaths; if a future `vm/*` type is dangerous, it must
be added to the deny-list explicitly.

### Examples

- `["*"]` — any msg except auth/* and vm/add_package. Device-login pattern.
- `["vm/exec:gno.land/r/jae/blog"]` — only `MsgCall` to that realm.
  `MsgRun`/`MsgSend`/`MsgMultiSend` rejected.
- `["bank/send"]` — only coin transfers. No realm calls.
- `["vm/exec:gno.land/r/jae/blog", "bank/send"]` — both: realm-scoped
  call AND coin transfer.
- `[]` — **invalid** at create-time.

### Ante wrapper: checkSessionRestrictions

The two filters above apply per-msg, per-session-signer. The check
uses `DelegatedAccount` presence in the context map, not a type
assertion to `*GnoSessionAccount`, so it cannot be bypassed by
alternative session account types.

### AllowPaths validation at creation

`handleMsgCreateSession` calls `*GnoSessionAccount.ValidateAllowPaths`
through a local `allowPathsValidator` interface (parallel to
`allowPathsSetter`), which delegates to the parser. Rejected at
create-time:
- Empty `AllowPaths` slice (required field)
- Max `MaxAllowPathsPerSession` (8) entries
- Empty entry strings
- Trailing slashes
- Unknown route_types (e.g. `bank/foo`, `auth/create_session`,
  `vm/add_package`)
- Bare routes (e.g. `bank`, `vm`) — reserved for future relaxation
- Path suffix on non-`vm/exec` entries (e.g. `bank/send:foo`)
- Empty path after `:` (e.g. `vm/exec:`)
- `*:<path>` (wildcard with path)

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
- `isSession == true`: `allowPaths` carries the session's typed entries
  (e.g. `["*"]` for wildcard, or `["vm/exec:gno.land/r/foo", "bank/send"]`).
  AllowPaths is required at create-time, so an empty slice is unreachable.

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

The tm2 handler creates sessions via `acck.NewSessionAccount()`,
validates the AllowPaths grammar through a local `allowPathsValidator`
interface (delegating to the parser in
`gno.land/pkg/gnoland/allow_paths.go`), then sets `AllowPaths` through
a local `allowPathsSetter` interface — keeping tm2 unaware of the
grammar as a concept.

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
- Realm paths in `vm/exec:<path>` entries are not validated as
  realm-like — garbage paths simply fail to match anything (fail-closed).

### Neutral

- Wildcard sessions (`AllowPaths: ["*"]`) can call any realm AND send
  coins (still subject to the always-denied list). **Device login is
  the primary use case for this mode**: a user creates a session on a
  less-secure device (browser, mobile app), sets `AllowPaths` to
  `["*"]`, and relies on time expiry + `SpendLimit` to bound damage
  if the device is compromised. Because `SpendLimit` is authoritative
  across all outflows (see "Spend limit enforcement" above), this is
  safe even when the session can call arbitrary attacker-deployed
  realms.
- A session with `bank/send` listed needs the same `SpendLimit` care:
  the entry permits coin transfers to any address, bounded only by
  `SpendLimit`.
- `MaxSessionsPerAccount` (16), `MaxAllowPathsPerSession` (8),
  `MaxSessionDuration` (~4 years), and `MaxSpendPeriod` (30 days) are
  compile-time constants in `tm2/pkg/std/account.go`, not tunable params.
  Changing them requires a coordinated upgrade of all nodes.

## References

- tm2 ADR: `tm2/adr/adr-002-session-subaccounts.md`
- Issue gnolang/gno#1499: Account Sessions System (Cookie-Like)
