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

Add `banker.CallSend()`: it returns the message's send **only to the realm the
keeper credited** (`m.Realm.Path == OriginSendRecipientPath`), and nothing to any
other realm. It reads no frames.

This makes "what did I receive in this call" a direct, credited, non-forgeable
fact:

- A realm reached by any interposition — relay, closure, func-value alias, bound
  or interface method, or a shape no rule enumerates — is not the credited
  recipient, so `CallSend` reads zero and the realm mints/credits nothing. It
  closes *both* of #6211's blind spots for a payment realm without reasoning
  about the stack at all, because it keys on *who was paid*, not on call shape.
- The receipt is atomic and credited: the keeper's transfer and the recorded
  recipient are one message-level fact; `CallSend` reports exactly that, gated to
  the payee.

A standalone showcase realm
(`gno.land/pkg/integration/testdata/callsend.txtar`) demonstrates it: a payment
realm reads the credited send on a direct call, and a realm reached through an
intermediary reads zero. This PR changes no deployed realm; migrating a genesis
realm such as `wugnot` to `CallSend` is left as a follow-up (its own fix ships
via #6211).

Phase 2 — realm → realm forwarding — is implemented on the same primitive:
`banker.PayCall(toPkgPath, rlm, coins)` forwards coins from the caller's realm
to another realm and records a per-message credit for the payee, which the payee
reads (and consumes) through the *same* `CallSend()`. A per-message credit
ledger lives on `ExecContext` (allocated by the keeper); no VM-core, op_call, or
grammar change. This is why `CallSend` is deliberately **call-scoped** ("what
did this call deliver to me") rather than origin-scoped: the one receipt answers
for both an EOA payment (the message) and a realm forward (`PayCall`). Gating
the origin-scoped `OriginSend` instead would close the vulnerability but could
not host forwarding — coins another realm forwarded are not an "origin send".

`NumCallFrames` is untouched: it also feeds the `getRealm` gas model, and this
change needs no frame reasoning at all.

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

- A payment realm can mint/credit against `CallSend()` with no origin check; the
  guarantee no longer depends on frame counting.
- This PR changes no deployed realm. The showcase realm reads the credited send
  on a direct call and zero when reached through an intermediary; a forwarding
  router also reads zero until phase 2 — it mints nothing.
- New native surface: gas row added (mirrors `originSend`, recalibrate with the
  table on the reference machine); `generated.go` binding added by hand and
  should be confirmed against `go generate`.
- Rests on the same VM fact #6211 uses: an aliased/relayed body runs under its
  declaring realm while `OriginSendRecipientPath` stays the message-named path.
