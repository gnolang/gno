# RFC: Per-call coin value for cross-realm calls (gno's `msg.value`)

Status: draft / request for comment — **phase 1 implemented on this branch**
Scope: GnoVM (frames, cross-call), gno.land (`chain/banker`)
Relates to: the origin-call hardening in flight, and the `inert` policy's
payment realms.

## Implementation status

- **Phase 1 (done, this branch): the message-entry receipt.** `banker.CallSend()`
  returns the transaction's send *only to the realm the keeper credited it to*
  (`OriginSendRecipientPath`), and zero to any relayed realm. This needs no
  VM-core change — the coins are already moved by the keeper and the recipient
  is already recorded — so it is a small, self-contained native. It already
  makes the direct-deposit case safe without `AssertOriginCall`: a relay is not
  the recipient, so it sees zero. Files: `gnovm/stdlibs/chain/banker/banker.go`
  (`X_bankerCallSend`), `banker.gno` (`CallSend`), `generated.go` binding,
  `native_gas.go` row; test: `gno.land/pkg/integration/testdata/callsend.txtar`.
- **Phase 2 (designed below, not yet built): the cross-call value channel.**
  Letting a realm *forward* coins on a crossing call (the router case) needs
  per-call coin movement inside the interpreter and a way to attach value to a
  `cross`. That is the larger VM change described in "Design" §1 and §3 and is
  left as the next step.

Phase 1 covers "a realm knows what the message paid it"; phase 2 adds "a realm
can pay another realm within a call." Phase 1 alone already removes wugnot's
need for the origin check on direct deposits.

## Summary

Give a cross-realm call the ability to carry coins, credited to the callee
atomically as part of the call, and let the callee read exactly what it
received in *this* call. This is the coin analog of Ethereum's `msg.value`.
gno already has the caller analog — `cur.Previous()` is `msg.sender`. It has no
value analog, and that gap is why payment realms must lean on `AssertOriginCall`.

This is a feature, not a bug fix. The origin-call hardening is handled
separately (see #6211); this proposal removes the *reason* payment realms depend
on the origin check at all.

## Motivation

Today the only payment signal a realm has is `unsafe.OriginSend()`: the coins
attached to the transaction root, shared identically by every realm in the call
chain. It is the money twin of `tx.origin`, not of `msg.value`. So a realm that
mints or credits against a payment cannot ask "what did *I* receive in this
call?" — it can only read the transaction-wide envelope and then *infer*, from
an authorization check, that the envelope must correspond to coins it actually
holds.

That inference is the whole problem:

- **It is fragile.** The realm is safe only while `AssertOriginCall` holds the
  invariant it depends on — no realm interposed, so the credited envelope is
  the realm's own. If the internals of that check change, the payment guarantee
  changes with it. Payment safety rides on a stack-shape heuristic.
- **It forbids composition.** To keep the inference valid, `AssertOriginCall`
  must reject *any* intermediary realm. So a realm cannot accept a payment
  relayed by another realm — even a legitimate one, such as a swap router that
  takes native coin and wraps it as one step of a user's single transaction.

A per-call receipt replaces the inference with a fact.

## The gap, stated precisely

| concept | Ethereum | gno today |
| --- | --- | --- |
| immediate caller | `msg.sender` | `cur.Previous()` — present |
| value delivered in this call | `msg.value` | — missing |
| transaction initiator | `tx.origin` | `runtime.OriginCaller()` |
| value at the tx root | — | `unsafe.OriginSend()` (tx-wide) |

gno has the per-call *sender*; it lacks the per-call *value*. This RFC adds it.

## Design

### 1. A value channel on cross-calls

A crossing call may carry a `Coins` value, drawn from the *caller realm's own
balance* and credited to the callee's realm address as part of the call:

```go
// caller realm A, holding coins it received earlier
wugnot.Deposit(cross(cur), send(coins))   // coins: A's balance → wugnot's address
```

The value must be a *call attribute*, not an ordinary argument (the callee's
signature is `Deposit(cur realm)` and must not change). Exact spelling is open
(see Open questions); `send(coins)` alongside `cross(cur)` is the strawman,
mirroring Solidity's `f{value: v}()`.

Rules:

- Only a crossing call may carry value; a same-realm call cannot (there is no
  realm boundary to move coins across).
- The coins come from the caller realm's address; the caller must hold them, or
  the call panics before entering the callee (no partial state).
- The transfer and the call are one atomic unit: if the callee panics, the
  transfer reverts with it.

### 2. A per-call receipt

The callee reads what this call delivered:

```go
received := banker.CallSend()   // chain.Coins credited to me in THIS call
```

Bound to the current call frame, not the transaction. A nested cross-call sees
its own value; returning pops it. Non-forgeable: an intermediary that wants to
trigger a downstream credit must actually forward the coins in its own crossing
call, and is then the party on record as having delivered them.

### 3. VM mechanics

- Add `Received chain.Coins` (or an equivalent) to `Frame`, next to the existing
  per-call data (`WithCross`, `Cur`).
- When `doOpPrecall` sets up a crossing call that carries value: (a) move the
  coins caller-address → callee-address through the banker, (b) record them on
  the callee's frame. Charge gas for the transfer.
- `banker.CallSend()` (native) reads the current frame's `Received`.
- On panic/revert of the call, unwind the transfer with the frame — same
  atomicity the VM already gives to state changes.

### 4. Unify the message `-send` as the entry frame's receipt

A `MsgCall -send N` already credits N to the entry realm. Surface that as the
*entry frame's* `CallSend()`, so the primitive answers uniformly: "coins
delivered to the current realm by the call that entered it," whether that call
is the transaction message or a cross-call. `OriginSend()` remains for genuine
tx-root uses (fee attribution, `tx.origin`-scoped accounting) but is no longer
what a payment realm reaches for.

## What it looks like

Payment realm, no origin check, no envelope:

```go
func Deposit(cur realm) {
    received := banker.CallSend().AmountOf("ugnot")   // fact, not inference
    require(received >= ugnotMinDeposit, "...")
    adm.Mint(cur.Previous().Address(), received)       // credit whoever paid
}
```

Composition, now safe by construction:

```go
// a router forwarding a user's ugnot into wugnot as one atomic step
func SwapAndWrap(cur realm, ...) {
    // ... router received the user's ugnot via the MsgCall envelope ...
    wugnot.Deposit(cross(cur), send(chain.Coins{{"ugnot", amt}}))
    // wugnot mints `amt` to the router, which actually delivered it
}
```

## Security considerations

- **Removes the fragility class.** Minting is tied to a coin fact, not to a
  frame count. An origin-check regression can no longer become a mint bug,
  because payment no longer consults the origin check.
- **Receipt is non-forgeable and atomic.** Value moves and its acknowledgment
  are the same event; no caller or intermediary can report coins it did not
  deliver.
- **This is receipt/attribution safety, not blanket payment safety.** Reentrancy
  is orthogonal (Ethereum has `msg.value` and still has reentrancy). Callees
  keep the checks-effects-interactions discipline; this RFC does not change
  reentrancy exposure.
- **Interaction with readonly/borrow rules** must be specified: a value-carrying
  call is a state mutation (a transfer) and must be rejected in read-only
  contexts, like any other write.

## Backward compatibility

Additive and opt-in. `OriginSend`, `AssertOriginCall`, and `cur.Previous()` are
unchanged. Existing realms keep working. A realm migrates by reading
`CallSend()` and dropping its origin/envelope logic when it chooses to. No
consensus change to existing message types; the value channel lives in the VM's
in-flight cross-call, not on the serialized message.

## Alternatives considered

- **Balance-delta (`vault - totalSupply`).** Approximates receipt without a new
  primitive, but captures stray donations to the next depositor and bricks
  deposits if the balance ever dips below supply. No per-caller attribution.
- **Keep the origin check (status quo + fix).** Correct and shippable, but
  leaves payment safety derived from call structure and forbids composition.
- **`OriginSend()` + `cur.Previous().IsUserCall()`.** The current safe-but-
  restrictive pattern: safe only by disallowing intermediaries.

## Open questions

- **Syntax.** How value attaches to a crossing call without changing callee
  signatures — a `send(coins)` attribute, an extended `cross`, or a banker-armed
  form. Grammar/typechecker impact to be scoped.
- **Denom set.** Any `Coins`, or ugnot-only initially.
- **Gas schedule** for the transfer and the receipt read.
- **Message `-send` unification** (design 4): adopt now or keep `OriginSend`
  separate for the entry frame.
- **Failure semantics** at the boundary: revert-with-frame is proposed; confirm
  it composes with defers and cross-call panics.

## Rollout

1. VM: `Frame.Received`, value-carrying `doOpPrecall`, `banker.CallSend()`, gas.
2. Tests: filetests for per-call scoping, nesting, revert atomicity, read-only
   rejection; integration txtar for router-style composition.
3. Docs: `chain/banker` reference; a security-guide entry recommending
   `CallSend()` over envelope+origin-check for new payment realms.
4. Optional follow-up: migrate `wugnot.Deposit` to `CallSend()` and retire its
   origin-check dependence.
