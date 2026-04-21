# ADR-002: Session Subaccounts

## Status

Accepted

## Context

Many dApp interactions — chat, games, social feeds, governance voting —
require dozens or hundreds of on-chain actions per user session. Today
each action requires the master key to sign, which typically means a
Ledger confirmation or a keyring popup. This makes interactive dApps
unusable.

The goal: a user's master key authorizes a short-lived session key
(e.g., generated in a browser) to act on behalf of the master account,
subject to constraints, without further master-key confirmations.

### Requirements

1. **Fast auth in mempool.** Session verification must happen in the
   AnteHandler without running the Gno VM (~0.5ms overhead max).
2. **Identity continuity.** `OriginCaller` and `PrevRealm` see the
   master address, not the session key. Sessions are transparent unless
   the realm explicitly asks.
3. **Master always wins.** The master key can always revoke sessions.
4. **Replay protection.** Independent sequence numbers so master and
   sessions can submit txs concurrently.
5. **No Tx format changes if possible.** Changes affect all clients,
   SDKs, and signing libraries.

### Alternatives considered

- **Embedded sessions on BaseAccount.** Adds `Sessions []Session` to
  `BaseAccount`. Simpler (one IAVL read) but pollutes `BaseAccount`
  with session bookkeeping, requires `NextSessionSeqHint` for replay
  protection, and puts `AllowPaths` (a gno.land concept) in tm2.

- **Realm-level auth callbacks.** Running Gno code in the AnteHandler
  is too slow (2-15ms per tx) and creates a DoS vector.

- **Separate `SessionAccount` at `/s/<session>`** (flat namespace).
  Loses the ability to enumerate or revoke-all by master without
  scanning the entire store.

## Decision

Sessions are separate accounts stored at `/a/<master>/s/<session>` in
the IAVL store, sharing the `/a/` prefix path with the master account
for cheap sequential reads.

### Store key scheme

```
Regular accounts:   /a/<20-byte addr>                     (23 bytes)
Session accounts:   /a/<20-byte master>/s/<20-byte session> (46 bytes)
```

`IterateAccounts` filters by key length (`AccountStoreKeyLen = 23`) to
exclude session sub-keys. The `/s/` infix acts as a visual delimiter
for debugging raw store dumps.

### Account types

**tm2 layer** (`tm2/pkg/std`):

```go
type BaseSessionAccount struct {
    BaseAccount                                    // Address, PubKey, AccountNumber, Sequence
    MasterAddress crypto.Address                   // linked master
    ExpiresAt     int64                            // 0 = no expiry
    SpendLimit    Coins                            // empty = no spending allowed
    SpendPeriod   int64                            // seconds; 0 = lifetime cap
    SpendUsed     Coins                            // current period usage
    SpendReset    int64                            // period start time
}
```

- `GetCoins()` returns nil; `SetCoins()` rejects non-empty. Sessions
  do not hold coins.
- `SetMasterAddress()` has an immutability guard (cannot re-set).
- `DelegatedAccount` interface extends `Account` with session getters/setters.

**gno.land layer** (`gno.land/pkg/gnoland`):

```go
type GnoSessionAccount struct {
    BaseSessionAccount
    AllowPaths []string  // realm path prefixes; empty = unrestricted
}
```

`AllowPaths` is gno.land-specific. tm2 never interprets it — it's set
via a local `allowPathsSetter` interface at creation time and read via a
local `pathRestricted` interface in the VM native function.

### Tx format

```go
type Signature struct {
    PubKey      crypto.PubKey   // optional (stored on session at creation)
    Signature   []byte
    SessionAddr crypto.Address  // zero for master-key signatures
}
```

`crypto.Address` with `json:",omitempty"`. Amino skips the field in
both binary and JSON when the address is zero (`[20]byte{}`), so
master-signed txs pay no wire-size overhead. The AnteHandler checks
`!SessionAddr.IsZero()`.

### AnteHandler flow

One unified flow for master and session signatures:

1. **Phase 1: Resolve signers.** Load master accounts. If
   `sig.SessionAddr` is non-zero, load session from
   `/a/<master>/s/<SessionAddr>` and check expiry (`ExpiresAt > 0 &&
   blockTime >= ExpiresAt`; `ExpiresAt = 0` skips the check = no expiry).
2. **Phase 2: Deduct gas fees.** Always from master. Gas fees count
   against session spend limits via `DeductSessionSpend`.
3. **Phase 3: Verify signatures.** Use session's own `AccountNumber`
   and `Sequence` for sign bytes (both zero at genesis). Pubkey
   resolution: if `sig.PubKey` is nil, use stored key; if stored key
   is nil, set `sig.PubKey` on account (first tx); if both exist,
   they must match (reject mismatch). Increment sequence and persist.
4. **Phase 4: Propagate context.** Store session accounts map in
   `sdk.Context` for downstream handlers.

### Responsibility split

```
AnteHandler (tm2):      sig verify, sequence, gas fees + spend check, expiry
Ante wrapper (gno.land): msg type allowlist, AllowPaths
Bank keeper (tm2):      spend limit check on all SendCoins + InputOutputCoins
VM keeper (gno.land):   spend limit check on lockStorageDeposit
Realm code (gno):       optional business-logic authz via GetSessionInfo()
```

`SpendLimit` is authoritative at the bank-keeper layer: any outflow
from master's balance that goes through `bank.Keeper.SendCoins` or
`bank.Keeper.InputOutputCoins` is debited against the session. This
covers outer `msg.Send`, `bank.MsgSend`, `bank.MsgMultiSend`, and
in-realm `std.Send` from gno code. `SendCoinsUnrestricted` (gas
collection, storage deposit refunds) is the only intentional bypass.

### Replay protection

No `NextSessionSeqHint`. Session accounts have their own globally
monotonic `AccountNumber`. A revoked and re-created session gets a new
`AccountNumber`; sign bytes include it, so old signatures are invalid.

### Session lifecycle messages

```go
MsgCreateSession{Creator, SessionKey, ExpiresAt, SpendLimit, SpendPeriod, AllowPaths}
MsgRevokeSession{Creator, SessionKey}
MsgRevokeAllSessions{Creator}
```

`RevokeAll` is a prefix delete on `/a/<master>/s/`.

### Validation

`MsgCreateSession.ValidateBasic` rejects:
- Missing creator or session key
- Negative `ExpiresAt` (0 is allowed = no expiry)
- Negative `SpendPeriod`
- More than `MaxAllowPathsPerSession` (8) AllowPaths entries
- Invalid `SpendLimit` coins (negative amounts, malformed)

`handleMsgCreateSession` additionally checks:
- `ExpiresAt` in the past or beyond `MaxSessionDuration` (30 days)
- `SpendPeriod` beyond `MaxSessionDuration`
- Session key address collides with existing regular account
- Duplicate session key
- Session count exceeds `MaxSessionsPerAccount` (16)
- AllowPaths entries: no empty strings, no trailing slashes

### Spend limits

`SpendLimit` must include the gas fee / storage deposit denom
(e.g., `ugnot`) or the session can't pay gas and can't grow realm
storage — spending is checked per-denom, and a missing denom means
zero allowance (fail-closed). Empty `SpendLimit` means no spending at
all, useful when another signer pays gas.

`SpendLimit` is the session's high-water mark, not a net-balance
counter. Refunds (e.g., storage deposit refund when state is freed)
credit coins back to master but do **not** reverse `SpendUsed`. This
prevents a compromised session from churning state to extend its
effective spending past `SpendLimit`. The refunded coins still reach
master; the session's remaining budget stays at its high-water mark.

#### Enforcement points

- **`bank.Keeper.SendCoins`**: calls `CheckAndDeductSessionSpend`
  after `canSendCoins`. Covers outer `msg.Send` on MsgCall/MsgRun,
  `bank.MsgSend`, and in-realm `std.Send`/`banker.SendCoins` from gno
  code inside a session-signed MsgCall or MsgRun.
- **`bank.Keeper.InputOutputCoins`**: per-input check, for
  `bank.MsgMultiSend`.
- **`VMKeeper.lockStorageDeposit`**: explicit check before the
  `SendCoinsUnrestricted` transfer. Storage deposits must bypass the
  restricted-denom check (they're an always-valid system transfer),
  but session accounting still applies.
- **AnteHandler Phase 2**: gas fee deducted via `DeductSessionSpend`
  in memory; the fee transfer itself uses `SendCoinsUnrestricted` so
  it doesn't double-count through the bank-keeper hook.

Intentional bypasses: `bank.Keeper.SendCoinsUnrestricted` (gas fee
collection, storage deposit refunds). Anything that must move coins
as a system-internal transfer should use the unrestricted path.

`DeductSessionSpend` operates on the `DelegatedAccount` directly.
`CheckAndDeductSessionSpend` is a context wrapper that looks up the
session from context and persists after deduction, accepting a
`SessionAccountSetter` interface (not the concrete `AccountKeeper`)
so it can be called from the bank keeper without circular imports.

### Query endpoints

```
/auth/accounts/{master}/sessions         -> list all sessions
/auth/accounts/{master}/session/{addr}   -> specific session
```

### VM integration

`ExecContext.SessionAccount` carries the `DelegatedAccount` directly —
no intermediate `SessionInfo` struct. The Gno native function
`runtime.GetSessionInfo()` returns `(pubKeyAddr, expiresAt, allowPaths,
isSession)` using a local `pathRestricted` interface to read
`AllowPaths` without importing gno.land.

## Consequences

### Positive

- `BaseAccount` is unchanged — no migration needed.
- Clean layer separation: tm2 handles auth + spend, gno.land handles
  paths + msg types.
- O(1) session lookup by (master, session) key.
- Replay protection is free via `AccountNumber`.
- Session signatures are the same size as master signatures when
  `SessionAddr` is zero (field is omitted via amino `omitempty`).

### Negative

- Two IAVL reads per session tx (master + session). Mitigated by
  shared prefix path in the tree.
- `Signature.SessionAddr` is a wire format addition. Backward
  compatible (omitted when zero) but clients need to be aware.
- `IterateAccounts` requires a key-length filter to exclude sessions.
  Documented on `AddressStoreKeyPrefix`, `AccountStoreKeyLen`, and
  `IterateAccounts`.
- `SpendLimit` must include the gas fee denom or the session can't pay
  gas (fail-closed).

### Neutral

- Empty `SpendLimit` means "no spending allowed" — useful when another
  signer pays gas.
- `ExpiresAt = 0` means "no expiry" — valid until revoked.
- At the gno.land layer, sessions can send `exec` (MsgCall), `run`
  (MsgRun), `send` (bank.MsgSend), and `multisend` (bank.MsgMultiSend).
  Other msg types are denied. `add_package` is permanently blocked
  (sessions must not claim namespace under master); session-lifecycle
  msgs (`create_session`, `revoke_session`, `revoke_all_sessions`)
  are permanently blocked to prevent privilege escalation.
- **Device login** is a primary use case for sessions without
  `AllowPaths`. A user creates a session on a less-secure device,
  leaves `AllowPaths` empty, and relies on `SpendLimit` + time expiry
  to bound damage if the device is compromised. Because `SpendLimit`
  is authoritative across every outflow (bank-keeper hook + storage
  deposit hook), this is safe even when the session can call
  arbitrary attacker-deployed realms.
- `MaxSessionsPerAccount` (16), `MaxAllowPathsPerSession` (8), and
  `MaxSessionDuration` (30 days) are compile-time constants in
  `tm2/pkg/std/account.go`, not tunable params. Changing them
  requires a coordinated upgrade of all nodes.

## References

- gno.land ADR: `gno.land/adr/adr-001-session-subaccounts.md`
- PRs #3970 (moul) and #4218 (notJoon): prior session explorations
- Issue gnolang/gno#1499: Account Sessions System (Cookie-Like)
