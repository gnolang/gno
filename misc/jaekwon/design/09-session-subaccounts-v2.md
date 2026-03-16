# Session Subaccounts: Implementation Design (v2)

Companion to `08-session-accounts.md`. Sessions are subaccounts keyed
under the master, not embedded in BaseAccount.

## Store Keys

```go
// tm2/pkg/sdk/auth/consts.go
AddressStoreKeyPrefix = "/a/"
SessionStoreKeyInfix  = "/s/"

func AddressStoreKey(addr crypto.Address) []byte             // /a/<master>
func SessionStoreKey(master, session crypto.Address) []byte  // /a/<master>/s/<session>
func SessionPrefixKey(master crypto.Address) []byte          // /a/<master>/s/
```

Master and session share an IAVL prefix path — the second read reuses
cached tree nodes from the first. `RevokeAll` is a prefix delete on
`/a/<master>/s/`. Session enumeration is a prefix iterate on the same.

## Tx Format

```go
// tm2/pkg/std/doc.go
type Signature struct {
    PubKey      crypto.PubKey  `json:"pub_key"`                // optional
    Signature   []byte         `json:"signature"`
    SessionAddr crypto.Address `json:"session_addr,omitempty"` // NEW
}
```

`SessionAddr` is zero for master-key signatures. When set, the
AnteHandler constructs the session key as
`/a/<signer>/s/<SessionAddr>`. `PubKey` is always optional for session
txs — the session account stores it at creation time.

## Account Types

```go
// tm2/pkg/std/account.go — UNCHANGED
type BaseAccount struct {
    Address       crypto.Address
    Coins         Coins
    PubKey        crypto.PubKey
    AccountNumber uint64
    Sequence      uint64
}

// tm2/pkg/std/account.go — NEW
type BaseSessionAccount struct {
    BaseAccount
    MasterAddress crypto.Address `json:"master_address"`
    ExpiresAt     int64          `json:"expires_at"`
    SpendLimit    Coins          `json:"spend_limit,omitempty"`
    SpendPeriod   int64          `json:"spend_period,omitempty"`
    SpendUsed     Coins          `json:"spend_used,omitempty"`
    SpendReset    int64          `json:"spend_reset,omitempty"`
}

// tm2/pkg/std/account.go — NEW
type DelegatedAccount interface {
    Account
    GetMasterAddress() crypto.Address
    GetExpiresAt() int64
    GetSpendLimit() Coins
    GetSpendPeriod() int64
    GetSpendUsed() Coins
    SetSpendUsed(Coins) error
    GetSpendReset() int64
    SetSpendReset(int64) error
}

// For setting AllowPaths from tm2 handler without importing gno.land.
type PathRestricter interface {
    SetAllowPaths([]string)
}

// Context key for passing session accounts downstream.
type SessionAccountsContextKey struct{}
// Context value: map[crypto.Address]DelegatedAccount (signer addr → session)
```

```go
// gno.land/pkg/gnoland/types.go — UNCHANGED
type GnoAccount struct {
    BaseAccount
    Attributes uint64
}

// gno.land/pkg/gnoland/types.go — NEW
type GnoSessionAccount struct {
    BaseSessionAccount
    Attributes uint64   `json:"attributes,omitempty"`
    AllowPaths []string `json:"allow_paths,omitempty"`
}

func (gsa *GnoSessionAccount) SetAllowPaths(paths []string) {
    gsa.AllowPaths = paths
}

func ProtoGnoSessionAccount() std.Account {
    return &GnoSessionAccount{}
}
```

Layer separation:

- **tm2**: `BaseSessionAccount` — master link, expiry, spend limits.
  Generic; any tm2 chain can use delegated signing with spend caps.
- **gno.land**: `GnoSessionAccount` — adds `AllowPaths` (realm path
  prefixes). Only meaningful where realm paths exist.

## AccountKeeper

```go
// tm2/pkg/sdk/auth/keeper.go
type AccountKeeper struct {
    key          store.StoreKey
    prmk         params.ParamsKeeperI
    proto        func() std.Account  // existing — e.g. ProtoGnoAccount
    sessionProto func() std.Account  // NEW — e.g. ProtoGnoSessionAccount
}

func NewAccountKeeper(
    key store.StoreKey, pk params.ParamsKeeperI,
    proto func() std.Account,
    sessionProto func() std.Account,  // NEW
) AccountKeeper

// New methods:
func (ak AccountKeeper) GetSessionAccount(ctx sdk.Context, master, session crypto.Address) std.Account
func (ak AccountKeeper) SetSessionAccount(ctx sdk.Context, master crypto.Address, acc std.Account)
func (ak AccountKeeper) RemoveSessionAccount(ctx sdk.Context, master, session crypto.Address)
func (ak AccountKeeper) RemoveAllSessions(ctx sdk.Context, master crypto.Address)
func (ak AccountKeeper) IterateSessions(ctx sdk.Context, master crypto.Address, cb func(std.Account) bool)

// NewSessionAccount creates a new session account using the session prototype.
func (ak AccountKeeper) NewSessionAccount(ctx sdk.Context, master crypto.Address, pubKey crypto.PubKey) std.Account {
    acc := ak.sessionProto()
    acc.SetAddress(pubKey.Address())
    acc.SetPubKey(pubKey)  // set at creation — always known
    acc.SetAccountNumber(ak.GetNextAccountNumber(ctx))
    da := acc.(std.DelegatedAccount)
    da.SetMasterAddress(master)
    return acc
}
```

## gno.land Initialization

```go
// gno.land/pkg/gnoland/app.go
acck := auth.NewAccountKeeper(
    mainKey,
    prmk.ForModule(auth.ModuleName),
    ProtoGnoAccount,          // existing
    ProtoGnoSessionAccount,   // NEW
)
```

## AnteHandler (tm2)

One flow for master and session signatures.

```go
// tm2/pkg/sdk/auth/ante.go
func NewAnteHandler(...) sdk.AnteHandler {
    return func(ctx sdk.Context, tx std.Tx, simulate bool) (newCtx sdk.Context, res sdk.Result, abort bool) {
        // ... gas meter, memo validation ...

        signerAddrs := tx.GetSigners()
        signerAccs := make([]std.Account, len(signerAddrs))
        stdSigs := tx.GetSignatures()
        sessionAccounts := map[crypto.Address]std.DelegatedAccount{}

        // ——— Phase 1: Resolve all signers ———

        for i, signerAddr := range signerAddrs {
            signerAccs[i] = ak.GetAccount(newCtx, signerAddr)
            if signerAccs[i] == nil {
                return // ErrUnknownAddress
            }

            if !stdSigs[i].SessionAddr.IsZero() {
                sa := ak.GetSessionAccount(newCtx, signerAddr, stdSigs[i].SessionAddr)
                if sa == nil {
                    return // ErrUnauthorized("unknown session")
                }
                da := sa.(std.DelegatedAccount)
                if newCtx.BlockTime().Unix() >= da.GetExpiresAt() {
                    return // ErrSessionExpired
                }
                sessionAccounts[signerAddr] = da
            }
        }

        // ——— Phase 2: Deduct fees from first signer (always master) ———

        if !tx.Fee.GasFee.IsZero() {
            res = DeductFees(bank, newCtx, signerAccs[0], ...)
            if !res.IsOK() {
                return
            }
            signerAccs[0] = ak.GetAccount(newCtx, signerAddrs[0])
        }

        // ——— Phase 3: Verify signatures, check spend, persist ———

        for i, sig := range stdSigs {
            if isGenesis && !opts.VerifyGenesisSignatures {
                continue
            }

            da, isSession := sessionAccounts[signerAddrs[i]]

            // Pick the account that holds the pubkey + sequence.
            var sigAcc std.Account
            if isSession {
                sigAcc = da.(std.Account)
            } else {
                sigAcc = signerAccs[i]
            }

            // Resolve pubkey. For sessions, PubKey was set at creation,
            // so sigAcc.GetPubKey() is always non-nil. For master keys,
            // existing logic: first tx sets PubKey on the account.
            pubKey := sig.PubKey
            if pubKey == nil {
                pubKey = sigAcc.GetPubKey()
            } else if sigAcc.GetPubKey() == nil {
                // First tx for this master account.
                sigAcc.SetPubKey(pubKey)
            }
            if pubKey == nil {
                return // ErrInvalidPubKey
            }

            // Sign bytes: sigAcc's own AccountNumber and Sequence.
            signBytes, _ := tx.GetSignBytes(
                newCtx.ChainID(),
                sigAcc.GetAccountNumber(),
                sigAcc.GetSequence(),
            )

            sigGasConsumer(newCtx.GasMeter(), sig.Signature, pubKey, params)

            if !simulate && !pubKey.VerifyBytes(signBytes, sig.Signature) {
                return // ErrUnauthorized
            }

            // Check spend limits before persisting (sessions only).
            if isSession {
                if errMsg := checkSessionSpend(da, tx, signerAddrs[i], newCtx.BlockTime().Unix()); errMsg != "" {
                    return // ErrSessionNotAllowed(errMsg)
                }
            }

            // Increment sequence and persist — spend updates included.
            sigAcc.SetSequence(sigAcc.GetSequence() + 1)

            if isSession {
                ak.SetSessionAccount(newCtx, signerAddrs[i], sigAcc)
            } else {
                ak.SetAccount(newCtx, signerAccs[i])
            }
        }

        // ——— Propagate session accounts in context ———

        if len(sessionAccounts) > 0 {
            newCtx = newCtx.WithValue(std.SessionAccountsContextKey{}, sessionAccounts)
        }

        return newCtx, sdk.Result{GasWanted: tx.Fee.GasWanted}, false
    }
}

func checkSessionSpend(da std.DelegatedAccount, tx std.Tx, signerAddr crypto.Address, blockTime int64) string {
    // Aggregate GetReceived() across msgs signed by signerAddr.
    // Reset period if expired: da.SetSpendUsed(nil), da.SetSpendReset(blockTime).
    // Check against da.GetSpendLimit().
    // Deduct: da.SetSpendUsed(used.Add(totalSpend)).
    // All mutations are on the in-memory account object;
    // caller persists afterward via SetSessionAccount.
}
```

## gno.land Ante Wrapper

```go
// gno.land/pkg/gnoland/app.go
baseApp.SetAnteHandler(func(ctx sdk.Context, tx std.Tx, simulate bool) (...) {
    // ... gas price, genesis setup ...

    newCtx, res, abort = authAnteHandler(ctx, tx, simulate)
    if abort {
        return
    }

    // ——— AllowPaths check (gno.land layer only) ———

    sa := newCtx.Value(std.SessionAccountsContextKey{})
    if sa != nil {
        sessions := sa.(map[crypto.Address]std.DelegatedAccount)
        for _, msg := range tx.GetMsgs() {
            for _, signer := range msg.GetSigners() {
                da, ok := sessions[signer]
                if !ok {
                    continue
                }
                gsa, ok := da.(*GnoSessionAccount)
                if !ok || len(gsa.AllowPaths) == 0 {
                    continue
                }
                if !sessionAllowsMsg(gsa.AllowPaths, msg) {
                    return newCtx, abciResult(std.ErrSessionNotAllowed(...)), true
                }
            }
        }
    }
    return
})

func sessionAllowsMsg(allowPaths []string, msg std.Msg) bool {
    if msg.Type() != "exec" {
        return false
    }
    pp, ok := msg.(interface{ GetPkgPath() string })
    if !ok {
        return false
    }
    path := pp.GetPkgPath()
    for _, prefix := range allowPaths {
        if path == prefix || strings.HasPrefix(path, prefix+"/") {
            return true
        }
    }
    return false
}
```

## Messages

```go
// tm2/pkg/sdk/auth/msgs.go
type MsgCreateSession struct {
    Creator     crypto.Address `json:"creator"`
    SessionKey  crypto.PubKey  `json:"session_key"`
    ExpiresAt   int64          `json:"expires_at"`
    SpendLimit  std.Coins      `json:"spend_limit,omitempty"`
    SpendPeriod int64          `json:"spend_period,omitempty"`
    AllowPaths  []string       `json:"allow_paths,omitempty"`
}

type MsgRevokeSession struct {
    Creator    crypto.Address `json:"creator"`
    SessionKey crypto.PubKey  `json:"session_key"`
}

type MsgRevokeAllSessions struct {
    Creator crypto.Address `json:"creator"`
}
```

## Handler

```go
// tm2/pkg/sdk/auth/handler.go
func (ah authHandler) handleMsgCreateSession(ctx sdk.Context, msg MsgCreateSession) sdk.Result {
    acc := ah.acck.GetAccount(ctx, msg.Creator)
    if acc == nil {
        return // ErrUnknownAddress
    }

    blockTime := ctx.BlockTime().Unix()
    if msg.ExpiresAt <= blockTime {
        return // ErrUnauthorized("session already expired")
    }
    if msg.ExpiresAt > blockTime + std.MaxSessionDuration {
        return // ErrUnauthorized("session duration exceeds maximum")
    }

    sessionAddr := msg.SessionKey.Address()

    // Check collision with existing regular account.
    if ah.acck.GetAccount(ctx, sessionAddr) != nil {
        return // ErrUnauthorized("session key address collides with existing account")
    }
    // Check duplicate session.
    if ah.acck.GetSessionAccount(ctx, msg.Creator, sessionAddr) != nil {
        return // ErrUnauthorized("session key already exists")
    }
    // Check session count.
    count := 0
    ah.acck.IterateSessions(ctx, msg.Creator, func(_ std.Account) bool {
        count++
        return count >= std.MaxSessionsPerAccount
    })
    if count >= std.MaxSessionsPerAccount {
        return // ErrSessionLimit
    }
    // Check AllowPaths count.
    if len(msg.AllowPaths) > std.MaxAllowPathsPerSession {
        return // ErrUnauthorized("too many allow paths")
    }

    // Create session account via prototype.
    sa := ah.acck.NewSessionAccount(ctx, msg.Creator, msg.SessionKey)
    da := sa.(std.DelegatedAccount)
    da.SetExpiresAt(msg.ExpiresAt)
    if len(msg.SpendLimit) > 0 {
        da.SetSpendLimit(msg.SpendLimit)
    }
    if msg.SpendPeriod > 0 {
        da.SetSpendPeriod(msg.SpendPeriod)
    }
    da.SetSpendReset(blockTime)

    // Set AllowPaths via interface — concrete type is GnoSessionAccount.
    if pr, ok := sa.(std.PathRestricter); ok && len(msg.AllowPaths) > 0 {
        pr.SetAllowPaths(msg.AllowPaths)
    }

    ah.acck.SetSessionAccount(ctx, msg.Creator, sa)
    return sdk.Result{}
}

func (ah authHandler) handleMsgRevokeSession(ctx sdk.Context, msg MsgRevokeSession) sdk.Result {
    ah.acck.RemoveSessionAccount(ctx, msg.Creator, msg.SessionKey.Address())
    return sdk.Result{}
}

func (ah authHandler) handleMsgRevokeAllSessions(ctx sdk.Context, msg MsgRevokeAllSessions) sdk.Result {
    ah.acck.RemoveAllSessions(ctx, msg.Creator)
    return sdk.Result{}
}
```

## VM Integration

```go
// gnovm/stdlibs/chain/runtime/native.go
func X_getSessionInfo(m *gno.Machine) (pubKeyAddr string, expiresAt int64, allowPaths []string, isSession bool) {
    ctx := execctx.GetContext(m)
    sa := ctx.Value(std.SessionAccountsContextKey{})
    if sa == nil {
        return "", 0, nil, false
    }
    sessions := sa.(map[crypto.Address]std.DelegatedAccount)
    da, ok := sessions[originCaller]
    if !ok {
        return "", 0, nil, false
    }
    addr := da.(std.Account).GetAddress()
    var paths []string
    if gsa, ok := da.(*GnoSessionAccount); ok {
        paths = gsa.AllowPaths
    }
    return addr.String(), da.GetExpiresAt(), paths, true
}
```

## Query

```
/auth/accounts/{master}/sessions         → prefix iterate /a/<master>/s/
/auth/accounts/{master}/sessions/{addr}  → direct lookup  /a/<master>/s/<addr>
```

## Replay Protection

No `NextSessionSeqHint`. Session accounts have their own globally
monotonic `AccountNumber`. A revoked and re-created session gets a new
`AccountNumber`; sign bytes include it, so old signatures are invalid.

## Expiry Cleanup

Expired sessions are rejected in the AnteHandler (Phase 1). Lazy
pruning: `handleMsgCreateSession` can prune expired sessions during
its count iteration. No background goroutines.

## Comparison to Embedded Sessions (08-session-accounts.md)

| | Embedded (08) | Subaccounts (this doc) |
|---|---|---|
| BaseAccount changes | Sessions, NextSessionSeqHint | None |
| IAVL reads/session tx | 1 | 2 (shared prefix, cheap) |
| Session lookup | O(N) scan, N ≤ 16 | O(1) key lookup |
| AllowPaths layer | tm2 (violation) | gno.land (correct) |
| Replay on re-add | NextSessionSeqHint | Free (AccountNumber) |
| RevokeAll | `acc.Sessions = nil` | Prefix delete |
| sig.PubKey | Required for sessions | Always optional (set at creation) |
| Master account bloat | ~120 bytes/session | None |
| Tx format | No change | +SessionAddr on Signature |
| Session proto | N/A | ProtoGnoSessionAccount |
