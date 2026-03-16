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

1. **Message type restriction.** All session-signed messages must have
   `msg.Type() == "exec"` (MsgCall). MsgAddPackage, MsgRun, MsgSend,
   and all other types are denied. This check uses the
   `DelegatedAccount` presence in the context map, not a type assertion
   to `*GnoSessionAccount`, so it cannot be bypassed by alternative
   session account types.

2. **AllowPaths restriction.** If `AllowPaths` is non-empty, the
   MsgCall's `PkgPath` must match one of the allowed prefixes. Read
   via a local `pathRestricted` interface to avoid tm2 importing
   gno.land types.

### AllowPaths validation at creation

`handleMsgCreateSession` validates AllowPaths entries:
- Max `MaxAllowPathsPerSession` (8) entries
- No empty strings
- No trailing slashes (would silently make paths unreachable)

### Spend limit enforcement in VM keeper

`VMKeeper.Call()`, `Run()`, and `AddPackage()` all call
`auth.CheckAndDeductSessionSpend()` when `msg.Send` is non-zero.
This is defense-in-depth — `Run` and `AddPackage` are already blocked
by the "exec only" restriction, but the spend check protects against
future relaxation.

### VM integration

`ExecContext.SessionAccount` carries the `std.DelegatedAccount`
directly. The Gno stdlib exposes it via:

```gno
// chain/runtime
func GetSessionInfo() (pubKeyAddr address, expiresAt int64, allowPaths []string, isSession bool)
```

Return values:
- `isSession == false`: not a session tx (master key signed)
- `isSession == true, allowPaths == nil`: session with no path restrictions
- `isSession == true, allowPaths != nil`: session restricted to these paths

`pubKeyAddr` is the session key's address (for per-session tracking).
The master address is available via `OriginCaller()` as usual.

Realms that want session-specific authorization should check
`isSession` and apply their own limits:

```gno
func Trade(cur realm, pair string, amount int64) {
    _, _, _, isSession := runtime.GetSessionInfo()
    if isSession {
        // enforce session-specific trade limits
    }
}
```

### Session prototype

`AccountKeeper` accepts a `sessionProto` function:

```go
acck := auth.NewAccountKeeper(mainKey, prmk, ProtoGnoAccount, ProtoGnoSessionAccount)
```

The tm2 handler creates sessions via `acck.NewSessionAccount()` and
sets `AllowPaths` through a local `pathRestricter` interface, keeping
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

- Unrestricted sessions (empty AllowPaths) can call any realm. The
  user chose to authorize this; the session's other constraints
  (expiry, spend limit) still apply.

## References

- tm2 ADR: `tm2/adr/adr-002-session-subaccounts.md`
- Issue gnolang/gno#1499: Account Sessions System (Cookie-Like)
