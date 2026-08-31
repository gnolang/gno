# ADR: Phase A — Manual Validator Freeze/Unfreeze in `r/sys/validators/v3`

Companion to [PR #5886](https://github.com/gnolang/gno/pull/5886).
Phase B (automated freeze via off-chain monitors) is
[PR #5887](https://github.com/gnolang/gno/pull/5887).

## Context

`r/sys/validators/v3` today has exactly one way to change the valset: a GovDAO
proposal (`NewValidatorProposalRequest`), which applies operator-keyed
`ValoperChange` deltas and republishes the full set through
`sysparams.SetValsetProposal`.

That is the right shape for planned changes and the wrong shape for an
emergency. "This validator is double-signing right now" needs a lever that acts
in one transaction, and it needs to be reversible — a `{op, Power: 0}` remove
forgets the operator's voting power, so putting them back means someone has to
remember and re-propose the number.

## Decision

A **freeze** holds an operator out of the effective valset while recording what
it took away, so an unfreeze can put it back.

### Authority: two paths, both always available

1. **GovDAO proposal** — `NewFreezeProposalRequest` / `NewUnfreezeProposalRequest`.
2. **Direct call by a freeze admin** — `FreezeValidator` / `UnfreezeValidator`.
   `freezeAdmins` is managed exclusively by GovDAO and starts **empty**, so a
   fresh chain is in GovDAO-only mode until governance explicitly opts into the
   faster path.

Admin auth is `cur.Previous()` plus `cur.IsCurrent()`, **not**
`unsafe.OriginCaller()`. Tx-origin identity would let any realm an admin happens
to call freeze validators on their behalf; requiring the immediate caller to be
a user realm means the admin's own transaction is the caller and no realm can
sit in between and launder the authority.

### Operator-keyed, like `ValoperChange`

Callers name an **operator** address. The signing key is resolved from
`valoperCache` at freeze time and **re-resolved at unfreeze time**.

Keying on the signing address would republish a retired pubkey for an operator
who rotated while frozen — possibly the exact key they rotated away from because
it leaked. This is the stale-key hazard `newValoperChangeExecutor` already
documents and defends against, and it is the reason the whole design is
operator-keyed rather than address-keyed.

Unfreeze also re-checks `KeepRunning`, matching the executor: an operator who
opted out while frozen is not silently reinstated.

### Liveness floor: by count **and** by power

At least 2/3 of the validator roster must remain live after a freeze, measured
both ways.

A count-only floor is close to meaningless under weighted voting: freezing one
of three validators passes the count check and halts the chain if that validator
holds 40% of the power. Tendermint needs >2/3 of *power* to make progress. The
count floor stays as a cheap backstop against a roster shrinking to a handful of
members.

**Denominator is the roster, not the live set.** `GetValsetEffective` already
excludes frozen validators, so measuring against it alone would let the
denominator shrink with every freeze and erode the guarantee one validator at a
time. `heldOutCount` / `heldOutPower` put the roster back together.

### A held-out operator has exactly one owner

`NewValidatorProposalRequest` and `newValoperChangeExecutor` refuse any
`ValoperChange` naming a held-out operator, at **create and execute** time
(mirroring the existing `KeepRunning` re-check). Without the guard, an add would
put the operator back while `frozenSet` still holds the entry — two owners of
the same slot — and a later unfreeze would silently overwrite the power the
proposal set.

The check goes through a single `heldOutBy(op) string` helper rather than an
inline `frozenSet.Has`, so Phase B widens it to auto-frozen entries in one place
instead of four.

### `NewDropFrozenProposalRequest`

Forgets a frozen entry **without** restoring. It exists because a frozen
operator is otherwise stuck: they are absent from the effective valset, so a
`{op, Power: 0}` `ValoperChange` panics with "validator does not exist", and
unfreeze would be the only exit. Dropping leaves the operator out for good;
re-admission goes through the normal add path.

### Free-form strings are escaped at every markdown sink

Freeze reasons are free-form and reach two different markdown sinks, each
needing a different escaper:

- `Render`'s table cells → `sanitize.TableCell`. Applied even though the writer
  is privileged: a bare `|` alone silently reshapes the table, which is a
  correctness problem before it is a security one.
- GovDAO proposal descriptions → `sanitize.InlineText`. GovDAO renders
  descriptions completely raw (no clamp), so a leading `#` or `>` in a reason
  restructures the proposal page voters read.

Escaping happens at the sink, never on the way into state: stored values and
`chain.Emit` payloads must stay verbatim, and `sanitize` is explicitly **not
idempotent**, so any scheme where a value could be escaped twice is wrong by
construction. The rule is stated once in `freeze.gno`'s header so a new sink has
somewhere to look. This is also why the tables are not routed through
`p/moul/mdtable`: it escapes `|` as `&#124;` with no `#`/`[]()` handling, so
`TableCell` would still be needed on top and the two escapes would compose
badly.

Phase B makes all of this load-bearing — there the reason comes from a key the
design assumes is stolen.

## Alternatives considered

**Reuse `{op, Power: 0}` removes and re-adds.** Rejected: the operator's power
is not recorded anywhere, so restoring is a human remembering a number. It also
provides no liveness floor — a remove can halt the chain — and no distinct
"held out, coming back" state for observers.

**Signing-address-keyed freezes.** Rejected: republishes retired pubkeys on
rotate-while-frozen. `TestUnfreeze_RestoresUnderRotatedKeyNotTheCapturedOne`
pins the regression.

**GovDAO-only, no admin list.** Rejected as the *only* option, but it is the
default: `freezeAdmins` starts empty, so a chain that never adds an admin gets
exactly this. Making the fast path opt-in rather than absent avoids a second
migration when governance decides it wants one.

**Count-only liveness floor.** Rejected; see above.
`TestFreeze_PowerFloorCatchesWhatTheCountFloorMisses` pins a case the count-only
floor waves through.

**Auto-expiry on manual freezes.** Deliberately not included — a manual freeze
persists until an explicit unfreeze or drop, because a human decided it. Phase B
adds expiry for the monitor path, where no human did. Whether the manual path
should also get a TTL is an open question on the PR.

## Consequences

- Two new realm-global AVL trees (`frozenSet`, `freezeAdmins`). Both empty on a
  chain that never uses the lever, and `Render` stays silent in that case.
- `publishValset` is a third copy of the "encode set → `SetValsetProposal`"
  logic, alongside `newValoperChangeExecutor` and `RotateValoperSigningKey`.
  Left un-deduped to keep this diff reviewable; worth folding together
  separately.
- No slashing. There is no bonded stake in gno.land today, so a freeze is the
  entire consequence of misbehaviour. If stake lands later, freeze is the
  natural hook to attach it to.
- The liveness floor means a sufficiently concentrated valset cannot freeze its
  largest member at all. That is the intended failure mode: the alternative is
  halting the chain.
