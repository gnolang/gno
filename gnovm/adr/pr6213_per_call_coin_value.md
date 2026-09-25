# ADR: per-call coin value for realms (`cross(cur, coins)` / `banker.CallSend`)

## Context

gno has the per-call *sender* but not the per-call *value*:

| concept | Ethereum | gno |
| --- | --- | --- |
| immediate caller | `msg.sender` | `cur.Previous()` |
| value delivered in this call | `msg.value` | missing |
| transaction initiator | `tx.origin` | `runtime.OriginCaller()` |
| value at the tx root | — | `unsafe.OriginSend()` (tx-wide) |

A realm's only payment signal is `unsafe.OriginSend()`, the envelope attached at
the chain root and read identically by every realm in the call chain. A realm
that mints against it must *infer*, from `AssertOriginCall`, that no realm
interposed and so the envelope is its own. That inference is fragile (repaired
in #6211, but still a stack-shape heuristic) and forbids composition: a router
cannot legitimately relay a payment, so nothing in `examples/` calls
`wugnot.Deposit` from another realm.

## Decision

Value is a property of one call, carried on the cross that makes it:

```go
func Deposit(cur realm) {                        // payee: a fact, not an inference
    got := banker.CallSend().AmountOf("ugnot")
    mint(cur.Previous().Address(), got)
}
func Route(cur realm) {                          // router: forward on the cross
    got := banker.CallSend().AmountOf("ugnot")
    wrap.Deposit(cross(cur, chain.Coins{{"ugnot", got}}))
}
```

- **Syntax.** `cross(rlm, coins)` takes an optional second argument of static
  type `chain.Coins`. Preprocess moves it to `CallExpr.Send` and keeps `cross`
  one-argument at runtime; the Go shim is `cross(rlm realm, coins ...any)`.
  Anything else in that position, or `cross` outside `Args[0]` of a crossing
  call, is a preprocess error.
- **Transfer.** `Send` is evaluated after the arguments; `doOpPrecall` pops it
  and, after `installCrossingCur`, hands it to `gno.CrossSendHandler` with the
  caller's and the freshly minted callee cur's addresses. `execctx` implements
  the hook: decode the `chain.Coins`, drop zero coins (no send), validate as
  `std` does, move through the context's banker, and return what moved, which
  becomes `Frame.Received`. No banker means any send panics. Gas:
  `OpCPUCrossSendBase/Slope`, mirrored from `bankerSendCoins`.
- **Entry call.** `Machine.CallReceived` walks to the nearest `WithCross`
  frame and returns its `Received`, or reports the call as the message entry:
  the MsgCall `.origin` frame (`Frame.EntryCall`, set by `installCrossingCur`)
  or code with no cross above it (`init`, `main`, tests). For the entry, the
  banker native returns the context's send when the current realm is the one
  the keeper credited (`OriginSendRecipientPath`, now also set by AddPackage),
  and marks it observed. MsgRun sets no recipient, so its self-transfer is not
  a receipt. Nothing is latched: every entry-level frame of a message sees the
  send, as `OriginSend` does today.
- **Reading.** `banker.CallSend()` is that: plain `fn(cur)` same-realm calls
  are transparent; a further cross, even into the same realm, a re-entrant
  call, or a relayed call reads zero.

No ledger, no payee path as a string, no end-of-message state: the coins and
the receipt exist exactly for the duration of the call, like `msg.value` and
CosmWasm `funds`.

## Alternatives considered

- **OriginSend + AssertOriginCall (status quo).** Correct only while the auth
  heuristic is; forbids composition.
- **Gate `OriginSend` to the credited recipient.** Minimal fix for existing
  realms, but origin-scoped by meaning, so it cannot carry a realm forward.
- **A banker-layer ledger (`banker.PayCall`), tried first on this branch.**
  A per-message credit keyed by (payer, payee) and consumed on read. It needed
  an unclaimed-credit guard, a nil-ledger refusal, harness seeding, and still
  let the first of several calls into the payee take the credit. Every one of
  those is what frame binding removes.
- **A `send(coins)` marker, `cross(cur, send(coins))`.** Reads well, but a
  uverse `send` cannot be shadowed and `send` is an ordinary identifier in
  nineteen existing realms and tests. The bare `chain.Coins` argument is typed
  just as strictly.
- **Vault balance delta (`balance - totalSupply`).** Captures stray donations
  and bricks deposits if the balance dips below supply; no attribution.

## Consequences

- A payment realm credits against `CallSend()` with no origin check; the
  guarantee rests on the frame, not frame counting.
- A callee panic recovered by the caller has already moved the coins, as any
  other Gno state change; only a failed message reverts them.
- `CallExpr` gains a persisted `Send` field (amino regenerated); `TransField`
  gains `TRANS_CALL_SEND`; `Frame` gains `Received`.
- A funded `MsgAddPackage` reads its send from `init`, crossing or not.
  `testing.SetOriginSend` drives `CallSend` at the entry level of a test.
- Stdlib source changed, so the pinned app hash is re-derived. `getRealm`-shaped
  gas row for `bankerCallSend`; all rows mirrored, recalibrate.
- Tests: `zrealm_cross_send{0,1}.gno`, `callsend.txtar` (relay reads zero),
  `callsend_forward.txtar`, `callsend_attribution.txtar`,
  `callsend_reentrant.txtar`, `callsend_addpkg.txtar`.
- Follow-ups: migrate `wugnot.Deposit` to `CallSend()`; a `Render`-safe
  read-only cross (a send is a write) once read-only crosses exist.
