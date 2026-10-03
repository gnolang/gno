# ADR: Jail validators that stop signing, so consensus stops counting them

## Status

Proposed. Not implemented.

## Context

### The problem

gno.land validator membership is Proof of Authority: GovDAO votes an operator
in, and the operator stays in until GovDAO votes them out or they opt out with
`UpdateKeepRunning(false)` — which itself only blocks re-admission; a removal
proposal still has to pass. Nothing on the chain reacts to a validator that
simply stops signing.

Tendermint commits a block once precommits from more than 2/3 of the voting
power are in. A validator that is in the set but silent counts in the
denominator and adds nothing to the numerator. With N validators of equal
power, the chain keeps producing blocks as long as at most ⌊(N−1)/3⌋ of them
are silent: none out of 3, one out of 4, 5 or 6, two out of 7, 8 or 9. The
margin is computed over the validators in the set, dead or alive, and today
the only way to take a dead validator out of the set is a GovDAO proposal that
passes and executes. Until then the dead validator keeps consuming one unit of
margin.

### What Cosmos SDK does

`x/slashing` keeps, per validator, a `ValidatorSigningInfo` (`StartHeight`,
`IndexOffset`, `MissedBlocksCounter`, `JailedUntil`, `Tombstoned`) and a bitmap
over the last `SignedBlocksWindow` blocks (default 100). BeginBlock reads
`LastCommitInfo.Votes`. When `MissedBlocksCounter` exceeds
`SignedBlocksWindow − MinSignedPerWindow × SignedBlocksWindow` (default 50%, so
more than 50 misses out of 100) and the validator has been bonded for at least
one window (`height > StartHeight + SignedBlocksWindow`), it is slashed by
`SlashFractionDowntime`, jailed with
`JailedUntil = block time + DowntimeJailDuration` (default 10 min), and its
counter and bitmap are reset. `x/staking` excludes validators with
`Jailed = true` from the bonded set in `ApplyAndReturnValidatorSetUpdates`, i.e.
from the ABCI `ValidatorUpdates`; the validator record itself persists. Once
`JailedUntil` has passed the operator sends `MsgUnjail`, and if not tombstoned
and still self-delegated the validator rebonds at the next EndBlock.

Two parts of that have no counterpart here: there is no stake to slash and no
bond to fall out of. What remains — sample, threshold, exclude, cool down,
return on request — is the liveness half, and it is what this ADR ports.

### What we build on

- **tm2 already delivers the signal.** `execBlockOnProxyApp` fills
  `RequestBeginBlock.LastCommitInfo` with one
  `VoteInfo{Address, Power, SignedLastBlock}` per validator of the previous
  height (`tm2/pkg/bft/state/execution.go`, `getBeginBlockLastCommitInfo`).
  `SignedLastBlock` is `block.LastCommit.Precommits[i] != nil`: a validator
  counts as absent when its precommit for the committed block is not in the
  canonical commit — offline, voted nil, or its precommit reached the next
  proposer too late to be included. `BaseApp` keeps the votes (`app.voteInfos`)
  and exposes them as `ctx.VoteInfos()`, but nothing in gno.land reads them.
- **Valset changes flow through params, not events**
  (`pr5485_valset_params.md`). `r/sys/validators/v0` publishes a full target
  set to `node:valset:proposed` + `node:valset:dirty` through `r/sys/params`;
  `EndBlocker` diffs it against the chain-managed `node:valset:current` and
  returns `ValidatorUpdates`. Only chain code can write `valset:current` (ctx
  sentinel in `nodeParamsKeeper.WillSetParam`). Realms read the set through
  `sysparams.GetValsetEffective()`: `v0.IsValidator`, the proposal executor and
  `RotateValoperSigningKey` all build the next target on top of it.
- **Operator identity lives in `r/gnops/valopers`**: operator address, current
  signing pubkey/address, auth list, `KeepRunning`. v0 mirrors it in
  `valoperCache` (pushed by `NotifyValoperChanged`). A key rotation republishes
  the target set under the new signing address.

### Non-goals

- Slashing of any kind. There is no stake; exclusion plus a cool-down is the
  entire penalty.
- Double-signing, for now. tm2 has the evidence types but no pool, no
  evidence in blocks, and `RequestBeginBlock.Violations` is commented out
  (`NoOpEvidencePool` in `node.go`); that plumbing is the prerequisite. Once
  it lands, a violation jails through the same overlay with
  `jailed_until = MaxInt64`, the tombstone value reserved in §1: the plain
  `block time < jailed_until` check in `MsgUnjail` then refuses forever with
  no special case and no state migration. A key rotation carries the entry
  over (§7) and a governance removal leaves it in place (§6), so neither
  clears it. The only way back is a new registration — a new operator
  address, since valopers refuses to re-register an existing one, and a new
  key — admitted by GovDAO, which is where Cosmos ends up too.
- Admission policy. GovDAO keeps deciding who is a member, and the tooling
  that automates that decision keeps its job. katana
  (<https://github.com/samouraiworld/katana>; samouraiworld, Go, pre-release:
  created July 2026, no license yet) is a daemon that reads gnomonitoring's
  per-validator health score — `100 × signed/total blocks` minus alert and
  downtime penalties — and files GovDAO proposals: `eject` below 30 over the
  current month, `lower_vp` below 60 over the week, `raise_vp` at 85 and
  above over the year, with a 30-day cooldown per validator, a rejection cap,
  dry-run and notifications. It automates the *proposal* half of today's
  path; the vote and its delay remain, and its eject fires once a validator
  has missed most of the month-to-date window: from minutes after a window
  reset to about ten days, depending on the day of the month it died. That is
  a policy loop at the scale of days to weeks, which is what a membership
  decision should be. Jailing is the reflex at the scale of minutes, needs no
  vote, and feeds it: a validator that keeps getting jailed
  is katana's eject case, readable from the `liveness/jailed` query or the
  events. One interaction to keep in mind: katana confirms its target is in
  the live consensus set (`/validators`) before proposing, and a jailed
  validator is not — to eject one it must read membership,
  `node:valset:current`, instead.

## Decision

Port the liveness half of `x/slashing` into a new `gno.land/pkg/sdk/liveness`
module, applied by gnoland's EndBlocker as a **jail overlay**: a separate list
of jailed addresses kept next to the governed set, never merged into it.
`node:valset:current` stays exactly what GovDAO decided, and consensus
receives `current` minus the jailed validators. A jailed validator is still a
member; it is only masked out of the signing set until unjailed.

### 1. A BeginBlocker tracks signatures

A new sdk module, `gno.land/pkg/sdk/liveness`, laid out like
`tm2/pkg/sdk/bank`:

- `keeper.go` — state and the BeginBlocker;
- `msgs.go`, `handler.go` — `MsgUnjail` and the `liveness/jailed` query (§3);
- `hook.go` — the end-of-tx hook (§7);
- `package.go` — amino registration.

`gnoland/app.go` constructs its keeper, `lk` below, after `acck`, `bankk`,
`vmk` and `prmk`, and wires it:

- `SetBeginBlocker(lk.BeginBlocker)`;
- `AddRoute("liveness", …)`;
- `prmk.Register("liveness", lk)`;
- the existing `EndTxHook` chains to `lk.EndTxHook`;
- the EndBlocker reads it (§2).

The BeginBlocker reads `req.LastCommitInfo.Votes`: the votes on block H−1,
cast by the set of H−1.

All liveness state lives in the module's own prefixed keyspace of `mainKey`
(`/liveness/…`, next to auth's `/a/`, bank's `/b/` and params' `/pv/`), not
in params, and is written by the keeper only. Per-validator keys use the
consensus address, the one the `valset:*` entries derive from their pubkey:

```
key                                value    meaning
/liveness/window                   int64    window the bitmaps were built for
                                            (reset on change, §6)
/liveness/dirty                    bool     set by a jail or an unjail,
                                            cleared by the EndBlocker
/liveness/jailed/<address>         int64    jailed_until, unix seconds;
                                            presence in this prefix = jailed;
                                            MaxInt64 = tombstoned (never)
/liveness/info/<address>           amino    signingInfo, below
/liveness/bitmap/<address>/<chunk> []byte   one chunk of the miss bitmap
/liveness/rename/<old>             address  new key continuing <old>'s
                                            record after a rotation (§7)

signingInfo {
    StartHeight int64 // height of the first vote seen since (re-)entry
    MissedCount int64 // number of set bits in the bitmap
}
```

The bitmap has `signed_blocks_window` bits, one per block, at index
`height % window`; it is split into fixed-size chunks so a miss rewrites one
chunk, not the whole window.

`lk.Jailed(ctx)` iterates the `/liveness/jailed/` prefix and returns the
addresses with their `jailed_until`.

Per vote, at BeginBlock(H):

1. `missed := !SignedLastBlock`, `idx := (H−1) % window`, `prev := bit(idx)`.
2. `prev == missed`: nothing to write. This is the steady state of an honest
   validator, whose bit is already clear.
3. `!prev && missed`: set the bit, `MissedCount++`.
   `prev && !missed`: clear the bit, `MissedCount--`.
   (A miss that leaves the window is thereby forgotten exactly when a new block
   enters it — the Cosmos sliding-window invariant, without a per-validator
   `IndexOffset`.)
4. If `missed`, `H−1 > StartHeight + window` and
   `MissedCount > window − window × min_signed_per_window / 100`: **jail**.
   Add the address to the jailed set with
   `jailed_until = header.Time + downtime_jail_duration`, drop its signing
   info and bitmap, set the keeper's dirty flag so the EndBlocker recomputes
   the consensus set this block, and emit a `ValidatorJailed` event in
   `ResponseBeginBlock` with `address`, `missed`, `window` and `jailed_until`.
   The decision and its reason are known here and nowhere else; the removal
   itself shows up in tm2's `EventValidatorSetUpdates` at EndBlock.

`H−1 > StartHeight + window` is the **grace period**, taken from Cosmos
(`height > StartHeight + SignedBlocksWindow`): no jail decision until a full
window has been observed since the validator's first vote. A validator that
has just entered has no history, and the bitmap cannot tell "no history" from
"missed", so judging it earlier would judge an empty window. Misses still
count during the grace, which is what the `missed` condition is for: without
it, a validator dark for most of its first window and back online when the
grace ends would be jailed on a block it signed — the SDK does exactly that.
With it, a jail always happens on a block the validator missed, that block
was committed without it, and removing it can never lower the signing share;
this is what lets the power floor go (see Alternatives).

Determinism: `LastCommitInfo` is derived from the committed block,
`header.Time` is the block time, and every write lands in the deliver state,
so all nodes compute the same result and it is covered by the app hash.

Signing info exists only for validators in `consensus`. A vote from an
address that has none creates it — with `StartHeight` set to that height —
only if the address is in `consensus`. When the EndBlocker's diff
removes an address (jail, governance removal), its signing info and bitmap
are deleted; a key rotation is the one exception, see §7. The address still
appears in `LastCommitInfo` for two more blocks, since tm2 applies the
removal at H+2; with no signing info and no `consensus` entry those votes are
ignored. A validator that re-enters starts fresh, with a new grace window. In
steady state every voter has signing info, so the `consensus` lookup never
runs.

### 2. EndBlocker applies the overlay

Chain-managed `node` keys (ctx sentinel, like `current` today):

| Key                     | Written by                     | Meaning                                                                       |
|-------------------------|--------------------------------|-------------------------------------------------------------------------------|
| `node:valset:current`   | EndBlocker (init: InitChainer) | Governed membership; what realms read (proposal baseline, `IsValidator`).     |
| `node:valset:consensus` | EndBlocker (init: InitChainer) | The set last handed to tm2 (the V_{H+2} bookkeeping `current` carries today). |

Nothing is written by realms and nothing new is read by them; `proposed` and
`dirty` keep their exact meaning. The jailed set is the keeper's
`/liveness/jailed/` prefix (§1).

EndBlocker, every block. The early return becomes "`dirty` is false and the
liveness keeper is not dirty":

1. If `dirty`: parse and validate `proposed` exactly as today (pubkey
   allow-list, whole-reject, empty-set floor), then `current = proposed`.
2. `target = current − lk.Jailed(ctx)`, by address. A jailed entry whose
   address is not in `current` is simply not matched. Keep the existing floor
   and refuse to emit a set with no live power (unreachable through jailing,
   see Consequences; the backstop stays).
3. `diff = consensus.UpdatesFrom(target)`, write `consensus = target`, delete
   the signing info of every address `diff` removes — or move it to its
   successor when a rename exists (§7) — clear both flags, return `diff`.

No step reconciles the jailed set with `current`: an entry stays until
`MsgUnjail` clears it, whatever governance does to the membership in between.

### 3. `MsgUnjail`: the operator's way back

A new gno.land message, handled in Go, so no realm changes (see "Unjail
through the realms" under Alternatives considered):

```
MsgUnjail {
    Operator crypto.Address // signer; the valoper's operator address
}
```

Defined in `gno.land/pkg/sdk/liveness/msgs.go` — route `liveness`, type
`unjail`, `GetSigners` returns `Operator` — next to `MsgSend` in
`tm2/pkg/sdk/bank/msgs.go` and `MsgCall` in `gno.land/pkg/sdk/vm/msgs.go`.
Handled in the module's `handler.go`; amino-registered in its `package.go`,
which `gnoland.Package` lists as a dependency like `vm.Package`; sent with a
new `gnokey maketx unjail` subcommand in `gno.land/pkg/keyscli`. The keeper
takes the VM keeper's `QueryEvalJSON` as an interface for step 2.

The handler, in order:

1. The jailed set is empty: reject `no validator is jailed`. Cheap, and it
   keeps the VM out of the path when there is nothing to do.
2. Resolve the operator: one `vmk.QueryEvalJSON` of
   `valopers.GetByAddr(<operator>)`, read `SigningAddress` and `KeepRunning`.
   No profile: reject. The eval must run under the transaction's gas meter
   rather than the query budget, so the sender pays for it.
3. `SigningAddress` not in the jailed set: reject `not jailed`.
4. `KeepRunning == false`: reject. An opted-out validator does not come back,
   consistent with the add rule in v0.
5. `ctx.BlockTime() < jailed_until`: reject `jailed until <t>`.
6. Remove the entry from the jailed set, set the keeper's dirty flag, emit
   `ValidatorUnjailed`. If the address is still in `current`, the EndBlocker
   of the same block re-adds it with its governed power and tm2 applies that
   at H+2; a fresh signing info and grace window start with its first vote.
   If governance removed it meanwhile, the entry is simply gone.

The chain learns who the operator is the way `assertGenesisValopersConsistent`
already reads v0 from Go: through the VM keeper, calling an exported function
of a realm that exists today. `GetByAddr` is non-crossing, so it is an eval,
not a `MsgCall`.

`MsgUnjail` is signed by the operator account. The auth list that
`UpdateKeepRunning` and `UpdateSigningKey` honour lives inside the profile's
`Authorizable` and is not exported, so it cannot be consulted from Go. The
operator can instead delegate the message to an account session
(`gno.land/adr/adr-001-session-subaccounts.md`) scoped to `liveness/unjail`,
which requires adding that route type to the `validSessionRouteTypes`
whitelist in `gno.land/pkg/gnoland/allow_paths.go`. Cosmos operators do the
same with an `x/authz` grant for `MsgUnjail`.

Unjail is an operator action, not a timer, on purpose. With no stake, an
explicit transaction is the only evidence available that someone is at the
keyboard. Automatic re-admission would put a still-dead node back, lower the
margin for another window, jail it again, and loop.

Realms see none of this until a hardfork exposes it (a mirror in
`r/sys/params` or a stdlib native); `v0.Render` shows a jailed validator as an
ordinary member meanwhile. Off-chain readers have the `liveness/jailed` ABCI
query, served by the module's handler like `bank/balances` and `vm/qeval` and
returning every jailed address with its `jailed_until`, plus the
`ValidatorJailed` / `ValidatorUnjailed` events.

### 4. Parameters

The liveness keeper registers on the params keeper like `auth`, `bank` and
`vm` (`prmk.Register("liveness", lk)`), so its params follow the
`<module>:p:<name>` convention, `liveness:p:*`, live in a typed `Params`
struct read with `lk.GetParams(ctx)`, are validated by the keeper's own
`WillSetParam`, and are set by GovDAO through the generic `r/sys/params`
factories (`NewSysParamInt64PropRequest(cur, "liveness", "p", …)`):

| Key                      | Type           | Default            | Constraint            |
|--------------------------|----------------|--------------------|-----------------------|
| `signed_blocks_window`   | int64, blocks  | `0` = **disabled** | 0 ≤ w ≤ cap (100 000) |
| `min_signed_per_window`  | int64, percent | `5`                | 0–100                 |
| `downtime_jail_duration` | int64, seconds | `600`              | ≥ 0                   |

An integer percent rather than a decimal: the params keeper has no decimal type
and nothing here needs one.

Absent keys read as zero, so a node that upgrades to this code changes nothing
until governance sets a window. That is the rollout: ship, then enable by
proposal, the way the valoper fees are staged.

Suggested first values, next to what the two chains gno.land's validators
come from actually run (queried 2026-09-29):

| chain          | window | min signed | jail  | block time | silence before jail |
|----------------|--------|------------|-------|------------|---------------------|
| Cosmos Hub     | 10 000 | 5%         | 600 s | ~5.7 s     | ~15 h               |
| AtomOne        | 10 000 | 5%         | 600 s | ~5.8 s     | ~15 h               |
| gno.land, here | 4 000  | 5%         | 600 s | 5 s        | ~5.3 h              |

The shape is theirs: a large window with a low threshold, so that a short
blip never counts and only a sustained absence does — the operators this set
is drawn from are used to that and to nothing tighter. The window is smaller
because the set is: a validator missing for 15 hours is noise among 180 and
the whole margin among five. The quantity to reason about is
`window × (1 − min_signed/100) × block time`, the continuous silence that
jails: it must exceed the longest routine maintenance an operator is expected
to do, and can be loosened toward the Hub's value by proposal as the set
grows. The 5 s is mainnet's `timeout_commit` (the tm2 default); test5 runs
1 s blocks, where the same window jails after about an hour, so networks
with faster blocks scale the window up.

### 5. Timing

A validator that misses block H−1 is evaluated at BeginBlock(H); if that
crosses the threshold its removal is returned at EndBlock(H) and tm2 applies it
at H+2. Same lag as any valset change today.

Wall-clock stalls do not count. Misses are recorded only for blocks that were
committed. If the chain halts (fewer than 2/3 signing), no block is produced,
nothing is recorded, and nobody is jailed for the outage when it resumes. That
is the intended reading of "offline": jailing is for validators the rest of the
set out-lives, not for outages the whole set shares.

### 6. Edge cases, decided

- **Governance removes a jailed validator.** `current` drops it; consensus
  was already without it. Its entry in the jailed set stays, so a later re-add
  is masked until the operator sends `MsgUnjail` — as in Cosmos, where
  `Jailed` survives unbonding.
- **Governance changes a jailed validator's power.** `current` updates; the
  validator stays jailed; consensus is unchanged until unjail.
- **`KeepRunning=false` while jailed.** Nothing changes chain-side. The
  operator has said they are leaving; `MsgUnjail` refuses until they flip it
  back, and GovDAO removes them as today.
- **Window change.** `signed_blocks_window` decides what `height % window`
  means. The tracker records the window it was built with; on mismatch at
  BeginBlock it clears every bitmap and counter and resets every `StartHeight`
  to H. Cost: one grace period. Setting the window to `0` clears the tracker
  and stops tracking; existing jail entries stay until unjailed.
- **Hardfork / genesis.** `InitChainer` seeds
  `consensus = current = req.Validators`; the jailed set starts empty. Signing
  infos are built lazily from the first commits. Nothing is replayed and
  nothing is migrated, so `valset_checkpoint_replay.md` is unaffected.
- **Late precommits.** A validator that is online but whose precommit the next
  proposer did not include is recorded absent for that block. That is noise,
  not downtime, and it is why the threshold is a fraction of a window and not a
  run of consecutive misses.

### 7. Key rotation

`UpdateSigningKey` replaces a validator's consensus key without changing its
operator: v0's `RotateValoperSigningKey` republishes `current` with the new
address in place of the old, and the EndBlocker's diff removes one address
and adds another. Treated like any removal, that would delete the record and
give the new key a fresh `StartHeight`, hence a new grace window — and with a
rotation throttle (`node:valoper:rotation_period_blocks`, default 600) shorter
than the window, an operator could keep a dead key in consensus by rotating
before every jail, or end a jail early by rotating out of it.

The chain cannot tell a rotation from a governance swap by looking at the
diff, and nothing in params records it. What does record it is the event v0
emits, `ValoperRotated{op, oldAddr, newAddr, height, applied}`. Realm events
carry the emitting package path, stamped by the VM
(`gnovm/stdlibs/chain/emit_event.go`, `currentPkgPath`), so it cannot be
forged by another realm. The module consumes it:

1. `lk.EndTxHook`, chained to the hook gnoland already installs, runs in
   DeliverTx only, after the messages, on a context whose writes are
   committed with the block iff the transaction succeeded. On an event with
   `PkgPath == "gno.land/r/sys/validators/v0"`, `Type == "ValoperRotated"` and
   `applied == "true"`, it writes `/liveness/rename/<oldAddr> = newAddr`.
   Nothing else.
2. In §2 step 3, when the diff removes `old` and a rename exists whose target
   the diff adds, the signing info, bitmap and jailed entry of `old` are
   moved to `new` instead of deleted. Renames are markers for that one
   block: the EndBlocker always runs in it, since the rotation set `dirty`,
   and clears the whole `/liveness/rename/` prefix at the end of its run,
   used or not. A rotation whose proposal is whole-rejected therefore
   leaves nothing behind.

The new key inherits `StartHeight`, the misses and the jail, tombstone
included: rotation neither grants a grace window nor shortens a jail. An
operator with a lost or compromised key still rotates freely and unjails the
normal way afterwards.

Two things this makes explicit:

- It is event-driven state in gnoland, which `pr5485_valset_params.md`
  removed for the valset. The difference: the event carries an annotation
  (which new key continues which record), not the source of truth. The
  valset still flows through params, the write is synchronous and durable,
  and a missing or malformed event degrades to today's behavior — a fresh
  record — never to a wrong validator set.
- v0's event is now a chain contract. The attribute names and the `applied`
  flag are read by Go, so a realm-side change would silently reopen the
  hole; a test pins the emit in `cache.gno` the way
  `TestValsetConstsDoNotDrift` pins the param keys, and the package-path
  gate moves with the realm version, like `valsetAuthorizedRealm`.

## Alternatives considered

- **Track liveness in the realm** (chain calls `v0.NotifyCommit(votes)` each
  block through `vmk.Call`). Rejected: one VM invocation per block on the
  consensus path, gas-less realm writes every block, and it is the chain→VM
  callback shape that `pr5485_valset_params.md` removed. The realm needs the
  verdict, not the votes.
- **Mirror `LastCommitInfo` into params for realms.** Same objection, with a
  store write per validator per block for state no realm reads.
- **Consecutive-miss counter instead of a window.** One integer per validator,
  no bitmap. Rejected: a validator that signs one block in ten never trips it
  yet lowers the margin nine blocks out of ten, and late-precommit noise makes
  "consecutive" flaky in the other direction. The window is also what operators
  and monitoring already understand from Cosmos. The bitmap costs 500 bytes per
  validator at a 4000-block window, and a clean signer never writes it.
- **Jail by removing from `current`.** Fewer keys, but every proposal is built
  on `current`; a jailed validator absent from it would be silently dropped
  from membership by the next unrelated proposal, and re-admission would need
  a vote. Membership and liveness must stay orthogonal — the same split as
  Cosmos's `Jailed` flag on a validator record that persists.
- **Auto-unjail when the cool-down ends.** Rejected in §3: it re-admits dead
  nodes on a loop.
- **Unjail through the realms** (`valopers.Unjail` → `v0.RequestUnjail` → a
  realm-written `node:valset:unjail` key drained by the EndBlocker). Same
  `maketx call` flow as `Register`, and the auth list would apply. Rejected:
  it edits three public realms, which only change at a hardfork, and it needs
  a reconcile step in the EndBlocker between realm-written requests and
  chain-written jail entries. `MsgUnjail` is one Go handler and ships with a
  node upgrade.
- **Sign `MsgUnjail` with the consensus key.** No VM lookup: the signer's
  address is the jailed address. Rejected: the consensus key belongs in
  `priv_validator_key.json` or a KMS, not in a `gnokey` keyring, and every
  other operator action is signed by the operator account.
- **Close the rotation hole another way.** A v0 check refusing rotation
  while jailed is a hardfork and blocks the lost-key case; raising
  `rotation_period_blocks` above the window only leaves a rotated dead key
  jailed for part of each cycle. Reading the event (§7) needs neither.
- **Unjail by GovDAO proposal.** Rejected: a full vote cycle to recover from a
  reboot turns a liveness mechanism into a punishment. GovDAO keeps the power
  to remove a jailed validator.
- **Drop offline validators inside tm2.** Rejected: membership policy belongs
  to the ABCI application, where Cosmos keeps it too; tm2 only reports votes.
- **A power-weighted floor on jailing** ("never jail more than 1/3 of the
  power in one block", or a cumulative cap on jailed power). Not needed: a
  validator is jailed only on a block it missed (§1 step 4), that block was
  committed, so its signers alone held more than 2/3 — and every validator
  jailed in the same block is a non-signer of it, however many there are.
  Removing them raises the signers' share; a cap would only keep provably
  absent power in the denominator. Seven validators with two dead: 5 of 7
  sign, 5 needed, margin zero, and after both are jailed 5 of 5 sign, 4
  needed, margin one. A cap protects against a verdict that can be wrong,
  such as an off-chain monitor's; this one is read from the commits. The
  existing empty-set floor stays as the backstop.

## Consequences

Positive:
- A dead validator stops counting against the quorum as soon as it has missed
  more than the allowed share of a window, plus two blocks, with nobody in the
  loop. The margin recovers by itself.
- Recovery is one operator transaction after the cool-down; no vote.
- No realm is modified; the whole feature ships with a node upgrade.
- Membership semantics for realms (`IsValidator`, proposal baselines, the
  valopers front-run guard) are unchanged: jailed validators are still members.
- Disabled by default; existing networks opt in by proposal.

Negative / trade-offs:
- gno.land gains a BeginBlocker and per-block store writes for absent
  validators, bounded by the 100-entry valset cap. Signers with a clean bit
  cost no write.
- One more chain-managed `node:valset:*` key (`consensus`) and a new
  `liveness` module with its own store, `liveness:p:*` params and the
  `liveness/jailed` query. Realm-side constants are untouched, so
  `TestValsetConstsDoNotDrift` is unaffected.
- A new sdk module and message type: amino registration, a `liveness` route,
  a `gnokey` subcommand, and a VM eval inside a Go handler, bounded by the
  sender's gas.
- `liveness/unjail` joins the `validSessionRouteTypes` whitelist, so unjail
  can be delegated to a scoped session; valopers' auth list, which is not
  readable from Go, does not apply to it.
- Realms cannot see jail state until a hardfork; `v0.Render` shows jailed
  validators as members meanwhile.
- The "active at H+2" meaning moves from `current` to `consensus`; `current`
  becomes "governed membership". The known readers to update are
  `InitChainer`, `EndBlocker` and their doc comments; realm readers already
  treat `current` as membership.
- A validator that is online but persistently excluded by proposers can be
  jailed. The window and the threshold are the tolerance; if a network hits
  this, the fix is the window, not the mechanism.
- The valset can now change without a vote. Light clients and relayers already
  handle arbitrary adjacent transitions
  (`pr5767_revert_valset_trust_level_cooldown.md`), so nothing new is required
  of them.

## Validation plan

- Unit tests for the tracker: bit set/clear as a miss enters and leaves the
  window, `MissedCount` equals the popcount, grace period, jail exactly at
  `MissedCount > window − window × pct / 100` and never on a signed block,
  reset on jail, window-change reset, no-op when the window is 0.
- `TestEndBlocker` extensions on the existing `valsetState` mock: a jail
  removes from the diff but not from `current`; clearing the entry re-adds
  with the governed power; a governance removal of a jailed validator leaves
  the entry and a re-add stays masked; a rotation moves the record and the
  jail to the new key and a rejected proposal drops the rename (§7); the
  empty-set floor holds.
- A drift test that pins the `ValoperRotated` emit in v0's `cache.gno`: event
  type, attribute names, and the `applied` flag (§7).
- Handler unit tests for `MsgUnjail`, one per rejection in §3 plus the accept
  path, against a seeded jailed-set entry and a stubbed operator lookup.
- An app-level sequencing test driving `BeginBlock`/`EndBlock` with crafted
  `LastCommitInfo` across a window, the way `x/slashing` is tested. The txtar
  harness runs one in-memory validator and cannot make one go silent.
- txtar for `MsgUnjail` end to end: wrong signer, operator without a profile,
  `no validator is jailed`, the real `GetByAddr` lookup through the VM, a
  session scoped to `liveness/unjail` that is accepted and one scoped to
  `bank/send` that is refused, and `gnokey query liveness/jailed` returning
  an empty set. The accept path and the `jailed until` refusal need a jail
  entry, which the txtar chain never produces; they stay at the Go level.
- A multi-validator devnet run with one node stopped, checking the
  `ValidatorJailed` event, the `consensus` set, and the unjail round trip;
  recorded in the PR, since CI cannot run it today.

## References

- Cosmos SDK `x/slashing` README (signing info, bitmap, `SignedBlocksWindow`,
  `MinSignedPerWindow`, `DowntimeJailDuration`, `MsgUnjail`) and `x/staking`
  `ApplyAndReturnValidatorSetUpdates`.
- `gno.land/adr/pr5485_valset_params.md`,
  `pr5767_revert_valset_trust_level_cooldown.md`,
  `valset_checkpoint_replay.md`.
- `tm2/pkg/bft/state/execution.go` (`getBeginBlockLastCommitInfo`),
  `tm2/pkg/sdk/baseapp.go` (`BeginBlock`),
  `gno.land/pkg/gnoland/node_params.go`.

## AI assistance

Written with AI assistance (Claude Code) from the code and ADRs cited above
and the Cosmos SDK `x/slashing` specification. The human author reviewed and
owns the design.
