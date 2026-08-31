# ADR: Phase B — Monitor-Driven Automated Validator Freeze in `r/sys/validators/v3`

Companion to [PR #5887](https://github.com/gnolang/gno/pull/5887).
Builds on Phase A ([PR #5886](https://github.com/gnolang/gno/pull/5886),
[ADR](./pr5886_validator_freeze_phase_a.md)), which this PR is stacked on.

## Context

Phase A gives the chain a freeze lever a human can pull: GovDAO, or an admin
GovDAO has vouched for. Both require a person to notice, decide, and sign.

Double-signing and sustained downtime are detected by machines, in seconds, and
the value of a freeze decays fast. The gap between "a monitor saw it" and "a
human pulled the lever" is the whole problem Phase B addresses.

## Decision

Registered off-chain monitors may freeze directly, and the chain enforces
safety invariants on every attempt.

### The design premise: the monitor is untrusted

A monitor is an unattended hot key running detection logic. The design question
is not "how do we make sure the monitor is honest" — it is **"assume the key is
stolen the day it is registered; what can the thief do?"**

Four on-chain invariants bound the answer:

| # | Invariant | Value |
|---|-----------|-------|
| 1 | Power floor | post-freeze live power ≥ **80%** of **roster** power |
| 2 | Auto-frozen cap | at most **12%** of **roster** power auto-frozen at once |
| 3 | Protected set | GovDAO-designated operators are immune to monitor freezes |
| 4 | Auto-expiry | entries expire after ~86400 blocks (~1 day); anyone can sweep |

Worst case for a stolen key: validators holding up to 12% of power drop out for
up to a day. Never enough to stop the chain. Recovery is a GovDAO
deregistration, a manual unfreeze, or simply waiting.

The floor is stricter than Phase A's 2/3 because no human is in the loop.

**Roster, not live.** `GetValsetEffective` already excludes everything held out,
so measuring the floor and the cap against it would let the denominator shrink
with every freeze and erode the guarantee one validator at a time. Both measure
against `live + held-out`.
`TestAutoFreeze_FloorDenominatorIncludesManuallyFrozenPower` pins a case the
eroding denominator would have waved through.

Monitor auth is `cur.Previous()` plus `cur.IsCurrent()`, not tx-origin. A
monitor signs unattended, so caller laundering matters more here than anywhere
else in the realm.

### Expiry, not monitor-initiated release

A monitor can freeze and **cannot** unfreeze. Release is time-based and
permissionless: `ExpireAutoFrozen` is callable by anyone, only ever gives power
back, and the caller pays the gas.

This is what makes invariant 4 a real bound rather than a promise — there is no
"the monitor forgot to release" failure mode, and no monitor action can extend a
freeze beyond its expiry.

`ExpireAutoFrozen` **skips** an operator it cannot restore cleanly — gone from
`valoperCache`, or `KeepRunning=false` — dropping the entry and emitting
`restored=false`, rather than letting one stuck operator block the sweep for
everyone else. It re-resolves the signing key, so an operator who rotated while
auto-frozen is never restored under a retired pubkey.

### Held-out state is shared with Phase A

`heldOutCount` / `heldOutPower` widen to cover auto-frozen entries, so Phase A's
liveness floor sees them, and `heldOutBy` widens so valoper proposals refuse an
auto-frozen operator at create and execute time.

That last one is load-bearing, not cosmetic. Without it a GovDAO add re-admits
an auto-frozen operator while the entry survives in `autoFrozenSet`, and
`ExpireAutoFrozen` later republishes the stale pre-freeze power over whatever
GovDAO decided. `TestValoperProposal_RejectsAutoFrozenOperator` and
`TestValoperProposal_RejectsOperatorAutoFrozenAfterProposalCreated` pin both
timings.

### Monitor reason strings are sanitized

A monitor's `reason` is rendered into the realm's markdown table. It is the one
place in this realm where an assumed-compromised key writes text a human reads,
so it goes through `sanitize.TableCell`. Without it a stolen monitor key can
inject markdown into `r/sys/validators`' page. The general rule, and the second
sink it covers, are documented in `freeze.gno`'s header (see the Phase A ADR).

`Evidence` is the same untrusted-source string and currently has **no** markdown
sink — it is stored and emitted, never rendered. That is fine today, and it is
the field that reintroduces this case the moment someone adds it to
`renderAutoFreeze`, which is why the header rule names it explicitly.

## Alternatives considered

**Trust the monitor and skip the invariants.** Rejected outright — it makes a
single hot key a chain-halt button.

**N-of-M monitor attestation before a freeze fires (Phase C).** Strictly better
on the single-stolen-key axis and deferred rather than rejected: it needs an
attestation-aggregation design, and the invariants here are what make a
single-monitor deployment safe enough to run in the meantime. The invariants do
not become redundant under N-of-M — they bound a coordinated-collusion case too.

**Monitor-initiated unfreeze instead of expiry.** Rejected: it adds a second
authority to the monitor key for no gain, and it introduces a
freeze-held-forever failure mode. Time-based release is strictly less
authority.

**Measuring the floor against live power.** Rejected; that is the eroding
denominator described above.

**A `NewDropAutoFrozenProposalRequest`, mirroring Phase A's drop.** Not
included: an auto-frozen operator that governance wants gone permanently is
reachable by waiting out the ~1 day expiry and then proposing a normal
`{op, Power: 0}` remove. Phase A needs its drop because a manual freeze has no
expiry at all; here expiry is the exit.

## Open: no count floor on the monitor path

Both monitor-path bounds are on **power**. Phase A additionally has a
2/3-**by-count** floor, and Phase B does not extend it to this path.

The consequence is real: under a skewed roster — one validator holding most of
the power, a long tail of small ones — a monitor can freeze the whole tail and
stay inside the 12% cap the whole way, walking the valset down toward a single
member. The chain keeps producing blocks, so no invariant here is violated, but
the centralization is reachable by monitor action alone.
`TestAutoFreeze_NoCountFloorOnTheMonitorPath` asserts this as **current
behaviour**; it is the test that should flip if a backstop lands.

Extending Phase A's `assertLivenessFloor` to this path was tried and **collides
with invariant 2**: the only rosters where an auto-freeze fits under 12% at all
are skewed ones, and on a small roster the count floor fires before the cap
does. It turns `TestAutoFreeze_CapAccumulatesAcrossFreezes` and
`TestAutoFreeze_FloorDenominatorIncludesManuallyFrozenPower` into count-floor
failures, destroying what they test. At realistic scale (20+ validators) the cap
binds first and a count floor is inert — the collision is a small-roster
artifact, but the tests that pin the cap's semantics live at small roster sizes.

Resolving this needs a decision on the number, not just the mechanism: a count
backstop that does not collide is probably not 2/3. Deferred to review.

## Open: restores can be silently whole-rejected by the pubkey-type allow-list

Landed after this PR was opened: [#5949](https://github.com/gnolang/gno/pull/5949)
(13 Jul 2026) removed secp256k1 support for validators and added a
chain-mirrored validator pubkey-type allow-list. The EndBlocker
(`gno.land/pkg/gnoland/app.go`) now **whole-rejects** a valset proposal if any
add or update in the diff carries a pubkey type outside that list — it logs,
clears `valset:dirty`, and returns. The realm gets no signal.

`r/gnops/valopers` enforces the allow-list at registration and rotation, so a
key that was legal when registered stays in `valoperCache` after the list
tightens. Restore paths republish straight from that cache, as adds:

- `doUnfreeze` (Phase A) removes the `frozenSet` entry, publishes, and returns
  success. If the publish is whole-rejected, the realm believes the operator is
  unfrozen while the valset never changed — and the recorded pre-freeze power is
  gone, so there is no second attempt: the operator is no longer frozen, so
  unfreeze cannot be retried, and re-adding needs someone to know the number.
- `ExpireAutoFrozen` (Phase B) is worse, because it batches. One entry with a
  now-disallowed key type sinks the whole publish, so every other operator in
  that sweep loses its restore too — while their `autoFrozenSet` entries have
  already been removed. This is exactly the failure mode the skip-don't-block
  logic was written to prevent, reappearing one layer below the realm.

The triggering scenario is not hypothetical: it is the secp256k1 removal itself,
on a chain that had secp validators frozen across the upgrade.

Not fixed here. The obvious in-realm fix — check
`sysparams.GetValsetPubKeyTypes()` before republishing and refuse or skip —
means duplicating valopers' bech32-decode + amino-type-URL helper into v3. The
better fix is probably a layer down, in
`r/sys/params.SetValsetProposal`, which would cover all three publishers
(`newValoperChangeExecutor`, `RotateValoperSigningKey`, `publishValset`) at once
and turn a silent EndBlocker drop into a panic the caller sees. Either way it is
a pre-existing hole these PRs widen rather than create, and it wants its own PR.

## Consequences

- Three new realm-global AVL trees (`monitors`, `protectedSet`,
  `autoFrozenSet`), all empty by default; `Render` stays silent until the path
  is used.
- `ExpireAutoFrozen` is unbounded in entries swept per call. The caller pays the
  gas and invariant 2 bounds how many entries can exist, so it is self-limiting
  today, but a batch limit would be cheap insurance.
- Invariant 2 caps total damage; it does **not** rate-limit a single monitor,
  which can exhaust the cap on its own. Per-monitor limits are deferred.
- The *set* of hold mechanisms is still enumerated in four functions
  (`heldOutBy`, `heldOutCount`, `heldOutPower`, and `AutoFreezeValidator`'s
  preconditions). `heldOutBy` centralized the proposal-path lookup — which is
  what closed the bug above — but not the knowledge of what the mechanisms are,
  so Phase C's attestation-driven hold would touch all four again. Worth folding
  into one enumeration before that lands, along with making explicit whether
  `NewFreezeProposalRequest`/`doFreeze` checking only `frozenSet` is deliberate.
- Evidence is a free-form string, recorded and emitted but not validated. A
  structured proof (e.g. a signed double-sign attestation) would let the chain
  verify rather than record. Deferred.
- `autoFreezeExpiryBlocks` assumes ~1s blocks. If block time changes materially
  the constant becomes wrong in wall-clock terms; it is a compile-time constant,
  so changing it needs a realm upgrade rather than a param change.
