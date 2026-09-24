# ADR: per-call coin value for realms (banker.CallSend / banker.PayCall)

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
interposed and so the envelope is its own. That inference is fragile (payment
safety rides on a stack-shape heuristic, repaired in #6211 but still a
heuristic) and forbids composition (a router cannot legitimately relay a
payment). The chain already records the missing fact: every `MsgCall` credits
its send to one realm, stored as `OriginSendRecipientPath`.

## Decision

A per-message **call-credit ledger** on `ExecContext`, keyed by `(payer, payee)`
package path and consumed on read, behind two natives in `chain/banker`:

- `CallSend()` takes the entry keyed by (`cur.Previous()`, `cur`), using the
  presented identities (`execctx.GetRealm`, which agrees with `cur`). The keeper
  seeds the message send as (user `""`, entry realm) for `MsgCall` and a funded
  `MsgAddPackage`. A relayed realm, a re-entrant call back into the entry realm,
  or a second read in the same call all take nothing. Reading marks the envelope
  observed for the unobserved-send guard.
- `PayCall(toPkgPath, rlm, coins)` moves coins from the live current realm's
  address to the payee's and records (payer, payee). Only the payee, entered by
  that payer, can read it. The keeper fails the message if a forward is never
  read (`ErrUnclaimedPayCall`; deterministic text, the ledger is an ordered
  slice, never a ranged map). `toPkgPath` must be a realm path; a context with
  no ledger refuses before any coins move.

```go
func Deposit(cur realm) {                       // payee: fact, not inference
    got := banker.CallSend().AmountOf("ugnot")
    mint(cur.Previous().Address(), got)
}
func Route(cur realm) {                         // router: forward, then cross
    got := banker.CallSend().AmountOf("ugnot")
    banker.PayCall("gno.land/r/x/vault", cur, chain.Coins{{"ugnot", got}})
    vault.Deposit(cross(cur))                   // vault reads exactly got
}
```

`CallSend()` is call-scoped, not origin-scoped, so one receipt serves both an
EOA payment and a realm forward; a gated `OriginSend` could not host forwarding.
No VM-core, op_call, or grammar change; `NumCallFrames` is untouched. `rlm` is
not `PayCall`'s first parameter because a realm-first signature is the
crossing-function form this non-realm package may not declare.

Why consume and key by payer: an unconsumed, path-only receipt let a re-entrant
call mint twice against one envelope, and a path-only forward credit went to
whichever realm entered the payee next. Both were demonstrated on this branch
before the ledger took this shape. Reentrancy is otherwise orthogonal; callees
keep checks-effects-interactions.

## Relationship to #6211

Complementary. #6211 repairs the *auth* primitive `AssertOriginCall`, needed by
every realm using it including non-payment authorization. This removes a
*payment* realm's dependence on that primitive; for such a realm the two overlap.

## Alternatives considered

- **OriginSend + AssertOriginCall (status quo).** Correct only while the auth
  heuristic is; forbids composition.
- **Gate `OriginSend` to the credited recipient.** Minimal fix for existing
  realms, but origin-scoped by name and meaning, so it cannot carry a realm
  forward. Could still ship alongside.
- **Value as a call attribute (frame form).** The target syntax; see below.
- **Vault balance delta (`balance - totalSupply`).** Captures stray donations
  and bricks deposits if the balance dips below supply; no attribution.

## Frame form: `f(cross(cur, send(coins)))`

`PayCall` binds value to the (payer, payee) pair for the rest of the message,
so between two calls from the payer into the payee the first `CallSend` read
wins, and a forgotten call surfaces only as the unclaimed error. Binding value
to one call frame, like `msg.value` or CosmWasm `funds`, removes both. User
code collapses to one line and the payee is unchanged:

```go
wrap.Deposit(cross(cur, send(chain.Coins{{"ugnot", got}})))   // A: chosen
wrap.Deposit(cross(cur), send(coins))                           // B: 2nd call attribute
wrap.Deposit(cross(cur).send(coins))                            // C: method on realm
```

A keeps `send` inside `cross`: the preprocessor already validates `cross` at
`Args[0]` and the realm's liveness there, and hangs the coins on the CallExpr;
precall moves them and sets `Frame.Received` beside `Frame.Cur`; `CallSend`
reads the nearest crossing frame. B adds a second positional slot to every
call site and needs its own rejection rules; C makes `send` look like a
`realm` method, so `x := cur.send(coins)` becomes a storable value that
carries money. `send` is syntax the preprocessor consumes, not a transfer and
not part of `realm`.

Cost: grammar, typechecker shim, a banker hook in `doOpPrecall` (the VM core
does not know the banker), and a decision on a callee panic recovered by the
caller, where the coins have already moved like any other Gno state change.
Sequencing: land `PayCall` to review the semantics; the frame form then wires
precall to the same transfer and deletes the ledger, `PayCall`, the unclaimed
guard and their error type.

## Consequences

- A payment realm credits against `CallSend()` with no origin check; the
  guarantee rests on a consumed, payer-keyed ledger entry, not frame counting.
- A router that pays and does not then cross into the payee gets an error, not
  a lost payment. Internal callouts (`callRealmBool`) carry no ledger, so a
  `PayCall` there fails closed. `gnovm/pkg/test.Context` seeds the ledger so
  `gno test` matches the chain.
- New natives shift the stdlib state hash (pinned app-hash updated); gas rows
  mirror `getRealm`/`bankerSendCoins`/`packageAddress` and need recalibration.
- Tests: `callsend.txtar` (relay reads zero), `callsend_forward.txtar`,
  `callsend_reentrant.txtar`, `callsend_unclaimed.txtar`, `callsend_addpkg.txtar`.
- Follow-ups: migrate `wugnot.Deposit` to `CallSend()`; the frame form above.
