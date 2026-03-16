# Session Accounts: Design Analysis

## Problem Statement

Many dApp interactions — chat, games, social feeds, governance voting —
require dozens or hundreds of on-chain actions per user session. Today
each action requires the master key to sign, which typically means a
Ledger confirmation or at minimum a keyring popup. This makes interactive
dApps unusable.

The goal: a user's master key authorizes a short-lived session key
(e.g., generated in a browser) to act on behalf of the master account,
subject to constraints, without further master-key confirmations.

From issue #6 (2021-08):

> "The big issue with chat is that it's unrealistic to require signing
> (and ledger verification of) for every chat message. There ought to be
> root-of-trust/CA like system where one can delegate some trust to
> another crypto system for messages (and other things like limited
> sends)."

## Requirements

1. **Fast auth in mempool.** The AnteHandler runs on every incoming tx in
   CheckTx. Signature verification is ~0.5ms (secp256k1). Any session
   mechanism that adds more than ~1ms per tx to the AnteHandler is
   untenable — it directly impacts mempool throughput and creates a DoS
   vector.

2. **Identity continuity.** Realms should see `std.PrevRealm()` and
   `std.OriginCaller()` as the master address, not the session key.
   The session is transparent unless the realm explicitly asks.

3. **Master always wins.** The master key can always revoke sessions,
   override session actions, and is never locked out.

4. **Replay protection.** Sessions need independent sequence numbers so
   that the master key and multiple sessions can submit txs concurrently
   without interfering.

5. **Minimal protocol surface.** The protocol should enforce what Gno
   code cannot (authentication, replay protection, time expiry) and
   leave domain-specific authorization to realms.

6. **No Tx format changes if possible.** Changes to the Tx wire format
   affect all clients, SDKs, hardware wallets, and signing libraries.
   Avoid if we can.

## The AnteHandler Constraint (Why Pure Gno Auth Fails)

**Note:** This section argues against the realm-level auth callback
approach (n0izn0iz's suggestion on PRs #3970 and #4218). The existing
PRs themselves (moul's #3970 and notJoon's #4218) are protocol-level
approaches that do session checks in Go — they do NOT suffer from this
performance problem. The criticisms of #3970 and #4218 later in this
document are about design complexity and scope, not speed.

This is the single most important constraint and it eliminates an entire
class of approaches, so it deserves detailed treatment.

The AnteHandler currently does:
1. Gas limit check — O(1)
2. Mempool fee check — O(1)
3. Tx.ValidateBasic — O(N) where N = num sigs
4. Memo validation — O(1)
5. For each signer: account lookup (IAVL read ~0.1ms), sig verify
   (secp256k1 ~0.5ms), sequence increment

Total: ~1ms per signer. At 1000 txs/sec mempool throughput, that's ~1
second of CPU per second. Tight but workable.

**Running the Gno VM in the AnteHandler would add:**
- Realm code loading from store: ~0.5ms (cached)
- VM machine setup: ~1-5ms
- Gno code execution: ~0.5-10ms depending on auth logic
- Total: 2-15ms per tx

At 1000 txs/sec, that's 2-15 *seconds* of CPU per second just for auth.
Nodes would fall behind. Worse: an attacker can craft txs referencing
auth realms with expensive computation, creating a targeted DoS on
mempool admission.

**Splitting CheckTx/DeliverTx doesn't help.** If we only run Gno auth
during DeliverTx, then the mempool admits any tx with a valid signature
regardless of session permissions. An attacker floods the mempool with
valid-sig, wrong-permissions txs. Validators include them, they fail
during execution, block space is wasted, the attacker only pays gas.
In the worst case this degrades network throughput.

**There's a deeper problem too:** even if we accepted the performance
cost, pure Gno auth still needs *something* in the AnteHandler. The
AnteHandler must verify that the tx signer has a right to claim the
caller identity. If we skip this, anyone can forge txs as anyone else —
the Gno auth realm only runs during DeliverTx, so the mempool has no
authentication at all. This is the "authentication must be protocol,
authorization can be flexible" principle.

**Conclusion:** Session authentication (who is this key, does it belong
to this account, is it unexpired) must happen in Go, in the AnteHandler.
Session authorization (what can this key do within a specific realm) can
optionally happen in Gno, during DeliverTx.

## Approach Analysis

### Approach A: Protocol-Level Sessions (Embedded in Account)

Add sessions directly to BaseAccount.

```go
type BaseAccount struct {
    Address       crypto.Address
    Coins         Coins
    PubKey        crypto.PubKey  // master key
    AccountNumber uint64
    Sequence      uint64
    Sessions      []Session      // NEW
}

type Session struct {
    PubKey     crypto.PubKey
    Sequence   uint64
    ExpiresAt  int64     // unix timestamp (block time)
    AllowPaths []string  // realm path prefixes; empty = all
}
```

**AnteHandler changes.** Currently `ProcessPubKey` checks that
`sig.PubKey.Address() == acc.GetAddress()`. For session keys this won't
match. The fix is straightforward — after the master key check fails,
scan `acc.Sessions` for a matching PubKey:

```go
func processSig(acc, sig, signBytes) {
    pubKey := sig.PubKey
    if pubKey.Address() == acc.GetAddress() {
        // Master key path (existing code, no change)
        verifyAndIncrementSequence(acc, sig, signBytes)
    } else if session := findSession(acc, pubKey); session != nil {
        // Session key path (new)
        checkExpiry(session)
        checkAllowedPaths(session, tx.Msgs)
        verifySignature(session.PubKey, sessionSignBytes, sig)
        session.Sequence++
    } else {
        return ErrUnauthorized
    }
}
```

The session scan is O(N) where N ≤ 16 (capped). Each iteration is a
pubkey bytes comparison (~32 bytes). Total: ~0.5µs. Negligible.

**No Tx format changes.** The Signature already carries a PubKey field.
The AnteHandler uses it to determine master-vs-session. Clients just
need to include the session PubKey in the Signature and use the session's
sequence number in sign bytes.

**Sign bytes.** The session key signs over the same SignDoc structure but
with the session's sequence number (not the master account's). The
AnteHandler reconstructs sign bytes using the session's sequence for
verification. This requires knowing which sequence to use *before*
signature verification — the PubKey in the Signature tells us which
session, which gives us the sequence.

**New message types:**
```go
type MsgCreateSession struct {
    MasterAddr crypto.Address  // signs with master key
    SessionKey crypto.PubKey
    ExpiresAt  int64
    AllowPaths []string
}

type MsgRevokeSession struct {
    MasterAddr crypto.Address
    SessionKey crypto.PubKey
}
```

Only the master key can create/revoke sessions.

**Pros:**
- Fastest possible auth. No new store lookups — session data is on the
  account we already loaded.
- No Tx format changes.
- Simple mental model: account has keys, some are sessions.
- Revocation is immediate (next block).
- All session state in one place (the account object).

**Cons:**
- Account objects grow. With 16 sessions × ~100 bytes = ~1.6KB added.
  For an account read on every tx, this increases IAVL read size. In
  practice most accounts will have 0-2 sessions; the cap keeps it
  bounded.
- Fixed permission model. AllowPaths is a string list of realm path
  prefixes. If we later want "allowed functions" or "max gas per tx" as
  protocol constraints, that's a state migration.
- Session state is consensus-critical. Bugs in session logic can halt
  the chain.

### Approach B: Subaccount Pattern (Separate Account Type)

Create a `SessionAccount` as a first-class account in the store:

```go
type SessionAccount struct {
    Address       crypto.Address  // derived from session pubkey
    MasterAddress crypto.Address
    PubKey        crypto.PubKey
    AccountNumber uint64
    Sequence      uint64
    ExpiresAt     int64
    AllowPaths    []string
    // No Coins field — fees paid by master
}
```

**AnteHandler flow:**
1. `MsgCall.Caller` = master address. `tx.GetSigners()` = [master addr].
2. AnteHandler loads account by master address → gets BaseAccount.
3. Signature PubKey doesn't match master PubKey.
4. Now what? We need to search for the session. But unlike Approach A,
   the session isn't on this account — it's in a separate store entry.
5. We'd need a secondary index: pubkey hash → session account address.
   Or: change `MsgCall.Caller` to session address.

If `MsgCall.Caller` = session address:
1. AnteHandler loads SessionAccount by session address.
2. Verifies signature — works (PubKey matches).
3. Deducts fees from... SessionAccount.MasterAddress's account. This
   requires loading a *second* account. Two IAVL reads per tx.
4. During DeliverTx, VM keeper must translate session address → master
   address for `OriginCaller`.

**Pros:**
- BaseAccount is unchanged.
- O(1) lookup by session address (standard account store).
- Clean separation: session is its own object.

**Cons:**
- Two IAVL reads per session tx (session account + master account for
  fees). In Approach A it's one read (master account with sessions
  embedded).
- Session awareness spreads to more layers: AnteHandler, VM keeper
  (caller translation), bank keeper (cross-account fee deduction).
  Approach A contains it to AnteHandler + new msg types.
- `MsgCall.Caller` semantics change. Currently Caller = the identity
  that signs. With subaccounts, Caller = session address but identity =
  master address. This semantic split leaks through the stack.
- Account store pollution. Millions of session accounts that are
  mostly expired, cluttering iteration, IAVL, and account number space.

### Approach C: Realm-Level Auth Callback (Pure Gno)

Add `AuthRealm string` + `AuthData []byte` to Tx. AnteHandler calls the
auth realm to verify.

**Why this doesn't work — detailed:**

1. **AnteHandler VM execution is too slow.** (Covered above.) 2-15ms
   overhead per tx destroys mempool throughput.

2. **DoS amplification.** An attacker submits txs referencing auth
   realms with expensive `Authenticate()` functions. Even with gas
   limits on auth calls, the VM setup cost alone (~2ms) is 4× the cost
   of a normal signature check.

3. **Authentication gap.** If we defer auth realm execution to DeliverTx
   and do a simple sig check in CheckTx, the mempool accepts any
   validly-signed tx regardless of authorization. The "caller" field in
   MsgCall is unverified until block execution. This means:
   - Mempool has no identity verification (anyone can claim any caller)
   - Block space is wasted on txs that fail authorization
   - Fee deduction is impossible before auth (who pays?)

4. **Circular dependency.** The auth realm is Gno code stored on-chain.
   To load it, you need the VM. To run the VM, you need an execution
   context. To build the execution context, you need the caller identity.
   To verify the caller identity... you need the auth realm. Bootstrapping
   this in the AnteHandler requires special-casing that undermines the
   "everything is Gno" premise.

5. **Upgrade risk.** If the auth realm has a bug, accounts using it
   could be permanently locked (no tx can authenticate). A "master key
   bypass" escape hatch means the protocol has session-awareness anyway,
   voiding the benefit.

**The "everything is Gno code" argument.** The appeal is clear:
composable, auditable, upgradable without hard forks. But authentication
is categorically different from application logic:

- Authentication runs on *untrusted input* (incoming txs from the
  network). Application logic runs on *authenticated input* (txs that
  passed the AnteHandler).
- Authentication must be fast because it's the DoS boundary. Application
  logic has gas metering for cost control.
- Authentication failure means "this tx doesn't exist" (drop from
  mempool). Application failure means "this tx failed" (charge gas,
  record on chain).

Putting authentication in Gno code moves the DoS boundary from a fast
Go path to a slow VM path. This is the fundamental issue.

### Approach D: Protocol Auth + Realm Authz (Hybrid)

Protocol handles authentication (fast, AnteHandler):
- Session key registration, signature verification, expiry, sequence
- Basic realm path constraints (optional)

Realms handle authorization (flexible, DeliverTx):
- Fine-grained per-action permissions
- Transfer caps, rate limits, custom logic
- Via `std.GetSessionInfo()` — returns nil for master-key txs

This is really Approach A with an added native function for opt-in
realm-level checks. The protocol changes are identical to Approach A.
The difference is philosophical: we explicitly acknowledge that the
protocol provides authentication + coarse authorization, and realms
layer finer authorization on top.

```gno
// In a DeFi realm:
func Trade(cur realm, pair string, amount int64) {
    caller := std.PrevRealm().Address()
    si := std.GetSessionInfo()
    if si != nil {
        // Session tx — enforce session-specific limits
        key := caller.String() + ":" + string(si.PubKeyHash)
        spent := sessionSpending.Get(key)
        if spent + amount > MAX_SESSION_TRADE {
            panic("session trade limit exceeded")
        }
        sessionSpending.Set(key, spent + amount)
    }
    // ... proceed with trade (same code for master and session)
}
```

For simple realms (chat, boards), no session awareness needed — the
protocol's realm path constraint is sufficient.

For complex realms (DeFi, treasury), `std.GetSessionInfo()` gives them
the tools to implement domain-specific limits in Gno.

**Pros:**
- All of Approach A's pros.
- Realms that need complex session authorization can implement it in
  Gno, without protocol changes.
- Clean auth/authz separation. Protocol = auth. Realm = authz.
- Follows the "centralized AuthControl" principle: the protocol IS the
  central auth authority; realms add domain-specific policy.

**Cons:**
- Realms that don't check `std.GetSessionInfo()` are fully open to
  any session with the right path prefix. This is fine for most cases
  (the user chose to authorize the session for that path) but could
  surprise realm authors who don't know about sessions.
- The `std.GetSessionInfo()` API is a new native function — not a large
  cost, but it does add to the stdlibs surface.

## Scoring

```
                        A (Embed)  B (Subacct)  C (Gno Auth)  D (Hybrid)
AnteHandler speed         +++        ++           --            +++
No Tx format changes      +++        +            --            +++
Implementation simplicity +++        +            -             ++
BaseAccount unchanged      -         +++          +++            -
Permission flexibility     +         +            +++           +++
DoS resistance            +++        ++           -             +++
Conceptual cleanliness     ++        ++           +++            ++
Store efficiency          ++          -           +++            ++
```

## Recommended Design

**Approach D (Hybrid): Protocol-level session authentication with
optional realm-level authorization.**

In practice, the protocol changes are identical to Approach A. The
addition is one native function (`std.GetSessionInfo()`) and the
explicit design intent that realms can layer authorization.

## Detailed Specification

### Account State

```go
type BaseAccount struct {
    Address       crypto.Address
    Coins         Coins
    PubKey        crypto.PubKey
    AccountNumber uint64
    Sequence      uint64
    Sessions      []Session       // NEW — max 16 entries
}

type Session struct {
    PubKey     crypto.PubKey   // session public key
    Sequence   uint64          // independent replay protection
    CreatedAt  int64           // block timestamp of creation
    ExpiresAt  int64           // block timestamp of expiry
    AllowPaths []string        // realm path prefixes; empty = unrestricted
}
```

Max 16 sessions per account. Each session is ~120 bytes (pubkey 33 +
sequence 8 + timestamps 16 + allowpaths ~63). Worst case account
growth: ~1.9KB. Typical: 0-2 sessions, ~240 bytes.

Session capacity could start at 8 for launch and be increased by
governance parameter if needed.

### Session Lifecycle

**Creation.** Master key signs `MsgCreateSession`:

```go
type MsgCreateSession struct {
    Creator    crypto.Address  `json:"creator"`
    SessionKey crypto.PubKey   `json:"session_key"`
    ExpiresAt  int64           `json:"expires_at"`
    AllowPaths []string        `json:"allow_paths"`
}

func (msg MsgCreateSession) GetSigners() []crypto.Address {
    return []crypto.Address{msg.Creator}
}
```

Validation:
- Creator's master key must sign the tx
- SessionKey must not already exist on this account
- ExpiresAt must be in the future (but not more than MaxSessionDuration)
- len(AllowPaths) ≤ 8
- len(Sessions) < MaxSessionsPerAccount (16)
- SessionKey.Address() must not collide with any existing account address

**Revocation.** Master key signs `MsgRevokeSession`:

```go
type MsgRevokeSession struct {
    Creator    crypto.Address  `json:"creator"`
    SessionKey crypto.PubKey   `json:"session_key"`
}
```

Removes the session from the account. Takes effect next block.

**Emergency: Revoke All.**

```go
type MsgRevokeAllSessions struct {
    Creator crypto.Address `json:"creator"`
}
```

Removes all sessions. Single tx, immediate (next block).

**Expiry.** Expired sessions are checked during the AnteHandler (reject
tx if session is expired) and cleaned up lazily during account writes.
No background goroutine, no separate GC. When the account is next
written (any tx involving this account), expired sessions are pruned.

### AnteHandler Modifications

The critical path. Changes are confined to `processSig` and a helper:

```go
func processSig(ctx, acc, sig, signBytes, simulate, params, sigGasConsumer) {
    pubKey := sig.PubKey

    // Try master key first (existing path)
    masterPK := acc.GetPubKey()
    if masterPK != nil && bytes.Equal(masterPK.Bytes(), pubKey.Bytes()) {
        // Existing master key verification — no changes
        return existingProcessSig(ctx, acc, sig, signBytes, ...)
    }

    // Try session key
    baseAcc, ok := acc.(*std.BaseAccount)
    if !ok {
        return ErrUnauthorized
    }
    session := baseAcc.FindSession(pubKey)
    if session == nil {
        return ErrUnauthorized("unknown signing key")
    }

    // Check expiry
    blockTime := ctx.BlockTime().Unix()
    if session.ExpiresAt > 0 && blockTime >= session.ExpiresAt {
        return ErrUnauthorized("session expired")
    }

    // Check allowed paths
    if len(session.AllowPaths) > 0 {
        for _, msg := range tx.Msgs {
            if !sessionAllowsMsg(session, msg) {
                return ErrUnauthorized("msg not allowed for session")
            }
        }
    }

    // Rebuild sign bytes with SESSION sequence (not account sequence)
    sessionSignBytes := GetSignBytesWithSequence(
        ctx.ChainID(), tx, acc.GetAccountNumber(), session.Sequence,
    )

    // Gas for signature verification
    sigGasConsumer(ctx.GasMeter(), sig.Signature, pubKey, params)

    // Verify signature
    if !simulate && !pubKey.VerifyBytes(sessionSignBytes, sig.Signature) {
        return ErrUnauthorized("session signature verification failed")
    }

    // Increment session sequence
    session.Sequence++
    ak.SetAccount(ctx, acc)

    return sdk.Result{}
}

func sessionAllowsMsg(session *Session, msg Msg) bool {
    // Extract realm path from msg
    path := msgRealmPath(msg)  // e.g., "gno.land/r/demo/boards"
    if path == "" {
        return false  // non-realm msgs (bank sends, etc.) not allowed
    }
    for _, allowed := range session.AllowPaths {
        if strings.HasPrefix(path, allowed) {
            return true
        }
    }
    return false
}
```

**Performance impact:** One extra pubkey comparison (~32 byte memcmp)
for the master key check, then O(N) session scan where N ≤ 16, each
iteration ~32 byte memcmp. Total added cost for session txs: ~1µs.
For master key txs: ~50ns (one extra branch).

### Sign Bytes

The `SignDoc` already includes `Sequence`. For session txs, the client
uses the session's sequence number:

```go
type SignDoc struct {
    ChainID       string
    AccountNumber uint64
    Sequence      uint64    // session sequence for session txs
    Fee           Fee
    Msgs          []Msg
    Memo          string
}
```

No structural change. The AccountNumber remains the master account's
number (replay protection across account deletion). Only the Sequence
differs.

### Fee Payment

Fees are always deducted from the master account. The current code
already does this correctly because:

1. `tx.GetSigners()` returns `[msg.Caller]` which is the master address
2. `signerAccs[0]` is loaded by master address
3. Fees deducted from `signerAccs[0]`

The session key difference only affects signature verification, not
fee deduction. No changes needed.

### VM Integration

**OriginCaller.** `MsgCall.Caller` = master address (unchanged). The VM
sets `OriginCaller = msg.Caller.Bech32()` as before. Realms see the
master address. Sessions are transparent.

**std.GetSessionInfo().** New native function for opt-in realm authz:

```gno
package std

type SessionInfo struct {
    PubKeyAddr Address   // session key's address (for tracking)
    ExpiresAt  int64
    AllowPaths []string
}

// GetSessionInfo returns session metadata if the current tx was signed
// by a session key. Returns nil for master-key transactions.
func GetSessionInfo() *SessionInfo
```

Implementation: the AnteHandler stores session info in the SDK Context
when processing a session-signed tx. The VM reads it from ExecContext.

```go
// In ExecContext (new field):
type ExecContext struct {
    // ... existing fields ...
    SessionInfo *SessionInfo  // nil for master-key txs
}
```

### What About Non-MsgCall Messages?

Session keys might need to call `MsgSend` (bank transfers) or
`MsgAddPackage`. Policy options:

**MsgSend:** Default deny for session keys. If `AllowPaths` is
for MsgCall realm paths, MsgSend doesn't have a realm path. A session
that needs to send coins should do it through a realm that wraps the
banker. Alternatively, add an explicit `AllowSend bool` to Session.

**MsgAddPackage:** Always deny for session keys. Package deployment
should require the master key.

**MsgRun:** Deny for session keys. MsgRun executes arbitrary code —
session constraints are meaningless if arbitrary code runs.

This is enforced in the AnteHandler's `sessionAllowsMsg`:

```go
func sessionAllowsMsg(session *Session, msg Msg) bool {
    switch msg.(type) {
    case vm.MsgCall:
        // Check AllowPaths against msg.PkgPath
        return checkPaths(session, msg.(vm.MsgCall).PkgPath)
    case vm.MsgAddPackage, vm.MsgRun:
        return false
    case bank.MsgSend:
        return false  // or: session.AllowSend
    default:
        return false
    }
}
```

### Cleanup and Pruning

Expired sessions are not actively removed. They're pruned lazily:

1. **On write:** When an account is written to store (after any tx
   involving the account), expired sessions are removed from the
   Sessions slice before serialization.
2. **On session tx:** If a session tx hits an expired session in the
   AnteHandler, the tx is rejected. The expired session remains until
   the account is next written (any successful tx).

No background goroutines. No EndBlocker work. Deterministic.

### Query Interface

No new query endpoint needed. The existing `/auth/accounts/{address}`
returns the full account object (amino JSON), which now includes the
`Sessions` array with each session's PubKey, Sequence, ExpiresAt,
AllowPaths, and spend fields. Clients use this to discover their
session's current sequence before signing.

### Migration

Existing accounts have `Sessions: nil`. The amino encoding is backward
compatible — a nil slice field simply isn't present in the encoded bytes.
Old account bytes decode into new BaseAccount with Sessions == nil.

No state migration needed. This is additive.

## What About Starknet / ERC-4337 / EIP-7701?

These are worth understanding but don't directly apply:

**ERC-4337 (Account Abstraction via UserOperations):** Introduces a
separate mempool for "user operations" that are verified by an
EntryPoint contract. This avoids the AnteHandler problem by creating
a parallel submission path. For Gno, this would mean a second mempool
with its own verification — significant infrastructure complexity.

**EIP-7701 (Native AA with EOF):** Embeds validation logic in account
code via EOF (Ethereum's new bytecode format). The EVM runs validation
code during tx processing. Ethereum can afford this because EVM
signature verification is already an EVM operation (ecrecover
precompile). For Gno, the VM is much heavier to spin up than a single
precompile call.

**Starknet Account Abstraction:** Every account is a contract with a
`__validate__` method. Starknet's sequencer runs validation on every tx.
This works because Starknet's STARK VM is designed for it and accounts
have always been contracts. Gno's architecture is fundamentally
different — accounts are data structures, not code.

The common thread: these systems were designed from the ground up with
in-VM authentication. Gno wasn't, and retrofitting it at the VM level
introduces the AnteHandler performance problem. The protocol-level
approach is the pragmatic answer for Gno's architecture.

## Security Review Findings (Post-Implementation)

Four issues found and fixed during review:

### Fixed: Session path unreachable (Critical)

The original `processSig` used `ProcessPubKey` to decide master-vs-session.
But `ProcessPubKey` returns OK whenever the account already has a pubkey
set — it just returns the stored pubkey without checking whether
`sig.PubKey` matches it. This meant the master key path always ran, its
`VerifyBytes` failed (wrong key), and the session fallback was never
reached. **Sessions were completely non-functional.**

**Fix:** Rewrote `processSig` to explicitly compare `sig.PubKey` against
`acc.GetPubKey()` using `bytes.Equal`. Only enters master path if they
match (or if sig.PubKey is nil and the account has a stored key, or on
first-tx pubkey setup). Otherwise falls through to session path.

### Fixed: AllowPaths prefix matching too loose

`strings.HasPrefix("gno.land/r/demo_evil", "gno.land/r/demo")` was
true. A session for "gno.land/r/demo" could call "gno.land/r/demo_evil".

**Fix:** `pathMatchesPrefix(path, prefix)` checks `path == prefix ||
strings.HasPrefix(path, prefix+"/")`. Only exact match or sub-path.

### Fixed: MsgCall.Send drains master account

Session keys could call MsgCall with `Send: <all coins>`, draining the
master account into a realm. A compromised session for r/demo/boards
could send all coins to the boards realm.

**Fix:** `sessionAllowsMsg` now checks for the `GetReceived()` interface
and rejects session txs with non-zero Send coins. Session keys can call
realm functions but cannot attach coins. If a realm needs deposits, it
uses the banker (controlled by MaxDeposit, which the user sets per-tx).

### Fixed: Session key re-add replay attack

If pubkey X was created (seq 0), signed txs, revoked, then re-added,
it would start at seq 0 again. Old signatures at the same sequence
over identical tx content could replay.

**Fix:** `BaseAccount.NextSessionSeqHint` is a monotonically increasing
counter. Every session sequence increment updates it if higher.
`AddSession` initializes new sessions at `NextSessionSeqHint`, so
re-added keys start at a sequence higher than any previous session ever
used. Old signatures are cryptographically invalid at the new sequence.

### Remaining known limitations

1. **Multi-signer txs with mixed auth modes.** If a tx has multiple
   signers and some use session keys while others use master keys,
   the `SessionInfo` in the context reflects the last session signer.
   The VM sees the same SessionInfo for all msgs in the tx. In practice,
   multi-signer session txs don't occur (each MsgCall has one signer).

2. **MaxDeposit not restricted.** Session keys can set MaxDeposit on
   MsgCall, allowing the target realm to pull coins via the banker.
   This is by design — the realm controls withdrawal logic, and the
   user authorized the session for that realm. Realm authors should use
   `runtime.GetSessionInfo()` to limit withdrawals for session callers.

3. **Gas fee drainage.** A compromised session key can submit txs
   that consume gas, draining the master account's gas balance. No
   per-session gas budget exists at the protocol level. Mitigation:
   set short ExpiresAt, use AllowPaths to limit scope.

## Open Questions

1. **Max session duration.** Should there be a governance parameter for
   the maximum allowed `ExpiresAt`? Currently hardcoded to 30 days.

2. **Session key crypto type.** Should session keys be limited to
   secp256k1? Ed25519 session keys would be useful for browser-based
   clients (WebCrypto API has better ed25519 support). The AnteHandler's
   `sigGasConsumer` already handles both types.

3. **Session-to-session delegation.** Should a session key be able to
   create sub-sessions? Probably not — this opens up delegation chains
   that are hard to reason about. Master key only.

4. **Rate limiting.** Should the protocol enforce per-session tx rate
   limits? Or is gas cost sufficient? A simple `MaxTxPerBlock uint16`
   field on Session could help.

5. **MsgCall.Send for sessions.** Currently blocked entirely. Should we
   add an optional `MaxSend Coins` field to Session for controlled
   coin transfers?

6. **Governance parameters.** What should be governable?
   - MaxSessionsPerAccount (default: 16)
   - MaxSessionDuration (default: 30 days)
   - MaxAllowPaths (default: 8)
   - SessionCreationGasCost (extra gas for MsgCreateSession)

## Critique of PR #3970 (moul) and PR #4218 (notJoon)

Both PRs take the right overall approach (protocol-level session auth
in Go, checked in the AnteHandler). Neither suffers from the performance
problems of the realm-callback approach. The issues are about design
complexity, consensus safety, and scope.

### PR #3970: "Account sessions support (exploration 1)"

**What it gets right:** The core idea is sound — accounts gain session
keys with independent sequences, the AnteHandler checks them, identity
is transparent to realms. The PR also includes an ADR document and
integration test scaffolding.

**Issue 1: Wholesale restructuring of BaseAccount.**

The PR replaces the simple `PubKey`/`Sequence` fields on BaseAccount
with a new abstraction layer:

```go
// Before (current):
type BaseAccount struct {
    Address       crypto.Address
    Coins         Coins
    PubKey        crypto.PubKey
    AccountNumber uint64
    Sequence      uint64
}

// After (#3970):
type BaseAccount struct {
    Address      crypto.Address
    MasterKey    AccountKey      // new interface
    RootSequence uint64
    Sessions     []AccountKey
    Coins        Coins
    AccountNumber uint64
    SequenceSum   uint64
}
```

The Account interface gains 8 new methods: `SetMasterKey`,
`GetMasterKey`, `AddSession`, `DelSession`, `DelAllSessions`, `GetKey`,
`GetAllKeys`, `SequenceByPubKey`, `GetSequenceSum`, `SetSequenceSum`.

This means every piece of code that calls `acc.GetPubKey()` or
`acc.GetSequence()` must be updated — gnoclient, gnokey maketx, the
AnteHandler, keeper, integration tests. The blast radius is large for
what is fundamentally an additive feature. Sessions can be added without
restructuring the base case. The existing PubKey/Sequence fields should
stay; sessions are a new field alongside them.

**Issue 2: SequenceSum is confusing and possibly unnecessary.**

Every tx increments both the key's own sequence AND a global
`SequenceSum` on the account. The sign bytes use SequenceSum in the
legacy path:

```go
signBytes, err = tx.GetSignBytes(
    newCtx.ChainID(), sacc.GetAccountNumber(), sacc.GetSequenceSum())
```

If all keys sign over the shared SequenceSum, then a tx from key A
changes the SequenceSum, invalidating any pending tx from key B that
was signed over the old SequenceSum. This defeats the purpose of
independent per-key sequences. The "modern path" (via SignerInfo) may
use per-key sequences, but having two signing paths — one that works
and one that doesn't — is a source of bugs.

The stated purpose of SequenceSum is replay protection when a session
key is removed and re-added. But the simpler fix is: don't allow
re-adding the same pubkey. Check at session creation time.

**Issue 3: Incomplete session validation.**

The AnteHandler has a literal TODO comment:

```go
// XXX: Check if the session is valid (not expired, etc)?
```

The core security check — is this session expired? does it have
permission for this message? — is not implemented. The PR modifies the
account model and signing infrastructure but doesn't implement the
actual constraints that make sessions safe. This suggests the
abstraction was designed before the requirements were clear.

**Issue 4: Unresolved shared-pubkey problem.**

The `AnteOptions` struct has:

```go
DenySharedPubkeys bool
// XXX: probably not possible, just because a session can be
// created with a pubkey that then become the root key of
// another account.
```

The same pubkey being both a session key on account A and the master
key on account B is a real security concern (a tx signed by that key
could be attributed to either account). The XXX comment acknowledges
the problem but leaves it unresolved. This needs a clear design
decision, not a flag.

**Issue 5: SignerInfo changes the Tx format.**

The PR adds `SignerInfo` (PubKey + Address) to support the "modern
path." This is a Tx wire format change that affects all clients, SDKs,
hardware wallet integrations, and signing libraries. The document's
recommended design avoids this — the existing Signature.PubKey field
already tells the AnteHandler which key signed.

### PR #4218: "Account session" (notJoon, building on #3970)

This PR is a more complete implementation with gno.land-specific session
types. It adds ~1600 lines across 18 files.

**What it gets right:** Separates tm2-level base types from
gno.land-level extensions. Has tests. The GnoAccount methods (gc,
CreateSession, GetSession, RevokeSession) are structurally correct.

**Issue 1: `time.Now()` in consensus-critical code. (Bug.)**

Six calls to `time.Now()` in code that runs during transaction
processing:

```go
// GnoAccount.gc() — called during account writes:
func (ga *GnoAccount) gc() int {
    now := time.Now()            // non-deterministic!
    // ... removes sessions where now > ExpirationTime
}

// BaseSession.IsExpired():
func (s *BaseSession) IsExpired() bool {
    if !s.ExpirationTime.IsZero() && time.Now().After(s.ExpirationTime) {
        s.State = SessionStateExpired    // mutates state!
        return true
    }
}

// BaseSession.NewBaseSession():
func NewBaseSession(...) *BaseSession {
    now := time.Now()
    return &BaseSession{CreatedAt: now, LastUsedAt: now, ...}
}
```

`time.Now()` returns the node's wall clock, which differs across nodes.
If `gc()` runs during `SetAccount` (called from the AnteHandler during
DeliverTx), different nodes may prune different sets of sessions,
producing different account state, causing an **AppHash mismatch and
chain halt**.

The correct approach: use `ctx.BlockTime()` (the block's deterministic
timestamp), passed explicitly to session-checking functions. This is
non-negotiable for consensus code.

Additionally, `IsExpired()` *mutates* the session's State field as a
side effect of a read operation. This is surprising and means a "check"
operation has write semantics, complicating reasoning about when state
changes.

**Issue 2: SessionManager with background goroutine.**

```go
func NewSessionManager(config SessionConfig) *SessionManager {
    sm := &SessionManager{
        activeSessions: make(map[string]*BaseSession),
        config:         config,
    }
    if config.CleanupInterval > 0 {
        go sm.cleanupLoop()    // background goroutine
    }
    return sm
}
```

A background goroutine with `sync.RWMutex` in `tm2/pkg/std` — the most
foundational consensus package. Even if the SessionManager isn't
currently wired into the AnteHandler (the GnoAccount.gc() method is
used instead), its presence in `tm2/pkg/std` means:

- It could be accidentally used in consensus paths.
- It establishes a pattern (in-memory session caches) that's at odds
  with deterministic IAVL-based state.
- The goroutine has no shutdown mechanism (no context, no stop channel).
  It leaks on node restart.

This should either be removed or moved to a non-consensus utility
package with clear warnings.

**Issue 3: `filepath.Match` for realm path matching.**

```go
func (s *GnoSession) HasRealmAccess(realm string) bool {
    for _, pattern := range s.RealmsWhitelist {
        matched, err := filepath.Match(pattern, realm)
        if err != nil {
            continue  // silently skip malformed patterns
        }
        if matched { return true }
    }
    return false
}
```

Three concerns:

1. `filepath.Match`'s `*` does **not** match path separators (`/`). So
   the pattern `"gno.land/r/demo*"` would NOT match
   `"gno.land/r/demo/boards"`. Users would need `"gno.land/r/demo/*"`
   and even then only one level deep. This is likely to surprise
   everyone.

2. Malformed patterns are silently ignored. A typo in a realm whitelist
   pattern means the session can't access anything — fail-closed, which
   is safe, but the user gets no feedback about why.

3. `filepath.Match` is a complex function with platform-specific edge
   cases (escaping, character classes, etc.). In consensus code, simpler
   is safer. `strings.HasPrefix` is trivially verifiable and covers the
   real use case (path prefix matching).

**Issue 4: Transfer capacity tracking in the protocol.**

```go
type GnoSession struct {
    // ...
    CoinsTransferCapacity std.Coins
}

func (s *GnoSession) ConsumeTransferCapacity(amount std.Coins) error {
    // Subtract amount from capacity
    newCapacity := capacity.Sub(amount)
    s.CoinsTransferCapacity = newCapacity
    return nil
}
```

This requires the bank keeper (or VM banker) to call
`ConsumeTransferCapacity` on every coin transfer, spreading session
awareness to yet another module. But what counts as a "transfer"?

- `MsgSend` — clearly yes.
- `banker.SendCoins` called from within a realm — should this count?
  The session authorized calling the realm; the realm decided to move
  coins internally.
- A DEX swap that moves coins between pools — is that a "transfer" by
  the session?
- Gas fee deduction — that moves coins from the master account to the
  fee collector.

The protocol has no clean boundary for "transfer initiated by the
session." This is fundamentally domain logic that belongs in realm code,
where the context is understood. A DeFi realm knows what a "trade" is;
the protocol does not.

**Issue 5: Freeze/Unfreeze state machine.**

```go
const (
    SessionStateActive  SessionState = "active"
    SessionStateFrozen  SessionState = "frozen"
    SessionStateExpired SessionState = "expired"
)

func (s *BaseSession) Freeze()   { s.State = SessionStateFrozen }
func (s *BaseSession) Unfreeze() { s.State = SessionStateActive }
```

Three states instead of two (exists vs. revoked). Questions:

- Who can freeze? Only the master key? Or can a session with
  `flagSessionCanManageSessions` freeze other sessions?
- What happens to in-flight txs signed while the session was active,
  submitted after it's frozen? They're rejected. If the session is
  later unfrozen, the user doesn't know the current sequence.
- What's the use case for freeze-then-unfreeze that isn't served by
  revoke-then-create-new? Freezing preserves the session's sequence
  and configuration. But creating a new session with the same
  configuration is equally easy.

The freeze feature adds state transitions, edge cases, and attack
surface for a use case that can be handled by revoke + recreate.

**Issue 6: Max 64 sessions.**

Each GnoSession carries: BaseAccountKey (~41 bytes), Flags (8 bytes),
ExpirationTime (~15 bytes), CoinsTransferCapacity (variable,
potentially large with multiple denominations), RealmsWhitelist
(variable, each path ~30-50 bytes). Conservative: ~150 bytes/session.

At 64 sessions: ~9.6KB per account. Every account read from IAVL loads
this. For hot accounts (validators, popular realm owners), this is read
on every tx they're involved in. A cap of 8-16 covers realistic use
cases (one per active dApp) without the state bloat.

**Issue 7: Hardcoded permission flags.**

```go
const (
    flagSessionUnlimitedTransferCapacity BitSet = 1 << iota
    flagSessionCanManageSessions
    flagSessionCanManagePackages
    flagSessionValidationOnly
)
```

Each flag is a protocol-level concept. Adding "can make IBC calls"
(commented as XXX in the code) requires a protocol upgrade. The flag
approach assumes we can enumerate all permission types upfront. History
suggests we can't — new features create new permission needs. The
simpler approach (realm path restriction + realm-level authorization)
is open-ended without protocol changes.

### Summary: What To Take From Each PR

From **#3970**: The core insight that sessions are protocol-level
entities with independent sequences, verified in the AnteHandler.
Discard the AccountKey interface restructuring and SequenceSum.

From **#4218**: The GnoAccount-level session management (gc, create,
revoke methods), the amino registration pattern, and the test structure.
Discard the SessionManager, time.Now() usage, filepath.Match,
transfer capacity, freeze states, and hardcoded permission flags.

## Implementation Estimate

The changes touch:
1. `tm2/pkg/std/account.go` — add Session struct, Sessions field, helper methods
2. `tm2/pkg/sdk/auth/ante.go` — session branch in processSig
3. `gno.land/pkg/sdk/vm/msgs.go` — add MsgCreateSession, MsgRevokeSession
4. `gno.land/pkg/sdk/vm/keeper.go` — handle session msgs, propagate session info
5. `gnovm/stdlibs/std/` — add GetSessionInfo native function
6. `tm2/pkg/sdk/auth/keeper.go` — session query endpoint
7. Tests

Core logic is ~300-400 lines of Go. The rest is message routing,
amino registration, query endpoints, and tests.

## Implemented Beyond the Original Spec

During implementation, two areas were extended beyond the spec above:

### Spend Limits (SpendLimit / SpendPeriod / SpendUsed / SpendReset)

The spec's `Session` struct only had `AllowPaths` for permission scoping.
Open Question #5 asked whether to add `MaxSend Coins` for controlled coin
transfers. The implementation answered this with a full spend-tracking
system on `Session`:

```go
SpendLimit  Coins  // max spend per period; empty = no spending allowed
SpendPeriod int64  // seconds; 0 = lifetime cap
SpendUsed   Coins  // spent in current period
SpendReset  int64  // block time when current period started
```

The AnteHandler's `sessionCheckSpend` aggregates `GetReceived()` across
all msgs in the tx, checks against the remaining allowance, resets the
period if expired, and deducts on success. If `SpendLimit` is empty,
no coin transfers are allowed (fail-closed). The `SessionInfo` passed
to the VM includes spend fields so realms can inspect remaining budget.

### NextSessionSeqHint (Replay Protection on Re-Add)

Covered in the Security Review above. `BaseAccount.NextSessionSeqHint`
is a monotonically increasing counter that tracks the highest session
sequence ever used. New sessions (including re-added pubkeys) start at
this value, making old signatures at lower sequences invalid.
