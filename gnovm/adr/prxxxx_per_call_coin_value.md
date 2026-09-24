# ADR: a credited per-call receipt for inbound coins (banker.CallSend)

## Context

Two independent things go wrong when a realm accepts a payment. One is an
*authorization* question — "was I entered straight from the message?" — answered
by `AssertOriginCall`, and repaired at the VM level by PR #6211 (it counts
realm-crossing closures and anchors the entry to the message-named package).

This ADR is about the *other* one, which #6211 does not touch and cannot: the
*value* layer. Even with a perfect origin check, a realm still has no way to ask
"how much did I receive in this call, and from whom." Its only payment signal is
`unsafe.OriginSend()`, and that signal is structurally the wrong shape:

- **It is transaction-wide, not per-call.** `OriginSend` is the envelope
  attached at the chain root, kept in the exec context, and read identically by
  every realm in the call chain. It is the money twin of `tx.origin`, not of
  `msg.value`. It reports what the transaction carried, never what *this* realm
  received.
- **So payment safety has to be *derived from* the auth check.** A realm reasons
  "AssertOriginCall passed, therefore no realm interposed, therefore the envelope
  the keeper credited must be mine." Payment correctness rides on a stack-shape
  heuristic. That coupling is fragile: if the auth check's internals change,
  the payment guarantee silently changes with it.
- **The receipt is not atomic.** The keeper moves the coins to the recipient
  realm and *separately* the realm reads a shared envelope; nothing ties "the
  coins moved" to "I can prove I got them" as one fact. An intermediary sits in
  the gap.

The chain already records the missing fact. Every `MsgCall` credits its `-send`
to exactly one realm and the keeper stores that realm as
`OriginSendRecipientPath` (set with or without coins, empty where there is no
message — fail-closed). Nothing exposed that fact to realm code.

## Decision

A per-message **call-credit ledger** on `ExecContext`, keyed by `(payer, payee)`
package path and consumed on read, behind two natives:

- `banker.CallSend()` takes the entry keyed by (`cur.Previous()`, `cur`), using
  the presented identities (`execctx.GetRealm`, which agrees with `cur`). The
  keeper seeds the message send as (user `""`, entry realm) for `MsgCall` and a
  funded `MsgAddPackage`. A relayed realm, a re-entrant call back into the entry
  realm, or a second read in the same call all take nothing. Reading marks the
  envelope observed for the unobserved-send guard.
- `banker.PayCall(toPkgPath, rlm, coins)` moves coins from the live current
  realm's address to the payee's and records (payer, payee). Only the payee,
  entered by that payer, can read it. The keeper fails the message if a forward
  is never read (`ErrUnclaimedPayCall`, deterministic text: the ledger is an
  ordered slice, never a ranged map). `toPkgPath` must be a realm path; a
  context with no ledger refuses before any coins move.

`CallSend()` is call-scoped, not origin-scoped, so one receipt serves both an
EOA payment and a realm forward; a gated `OriginSend` could not host forwarding.
No VM-core, op_call, or grammar change; `NumCallFrames` is untouched.

Why consume and key by payer: an unconsumed, path-only receipt let a re-entrant
call mint twice against one envelope, and a path-only forward credit went to
whichever realm entered the payee next. Both were demonstrated on this branch
before the ledger took this shape.

## Relationship to #6211

Standalone and complementary, not a second copy of the same fix. #6211 repairs
the *auth* primitive `AssertOriginCall`, which every realm using it needs —
including non-payment authorization, where `CallSend` does not apply. `CallSend`
removes a *payment* realm's dependence on that primitive entirely. For a payment
realm the two overlap (CallSend would also make it safe); everywhere else they do
different jobs.
Sequencing: #6211 ships the immediate, universal fix; this lands after it as the
payment primitive.

## Alternatives considered

- **OriginSend + AssertOriginCall (status quo).** Correct only while the auth
  heuristic is; couples payment to call shape; forbids composition.
- **Mint the vault balance delta (`balance - totalSupply`).** No new primitive,
  but captures stray donations to the next depositor and bricks deposits if the
  balance ever dips below supply. No per-caller attribution.
- **Rely on #6211 alone.** Fixes the auth layer, leaves payment derived from it
  rather than a first-class fact.

## Consequences

- A payment realm credits against `CallSend()` with no origin check; the
  guarantee rests on a consumed, payer-keyed ledger entry, not frame counting.
- A router that pays a realm and does not then cross into it gets an error, not
  a lost payment. Internal callouts (`callRealmBool`) carry no ledger, so a
  `PayCall` there fails closed. `gnovm/pkg/test.Context` seeds the ledger so
  `gno test` matches the chain.
- New natives shift the stdlib state hash (pinned app-hash updated); gas rows
  mirror `originSend`/`bankerSendCoins` and need recalibration.
- Tests: `callsend.txtar` (relay reads zero), `callsend_forward.txtar`,
  `callsend_reentrant.txtar`, `callsend_unclaimed.txtar`, `callsend_addpkg.txtar`.
