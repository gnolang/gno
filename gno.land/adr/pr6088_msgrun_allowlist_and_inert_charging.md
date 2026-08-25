# ADR: MsgRun allowlist, inert-submission charging, and the oracle's time bound

Builds on [PR #5888](https://github.com/gnolang/gno/pull/5888) (phase 2, inert
packages), which builds on [#5885](https://github.com/gnolang/gno/pull/5885)
(phase 1, `code_submission_policy`). Read
[pr5888_phase2_inert_packages.md](./pr5888_phase2_inert_packages.md) first.

## Context

Reaching the Go type checker is the expensive part of accepting code. It is
native Go, so it is not metered per operation; the only price on it is
`chargePreprocessGas`, a flat per-source-byte charge levied immediately before
it runs.

Phase 2 removes that cost from `MsgAddPackage` by storing packages inert and
type-checking them later, at `MsgEnablePackage`. Its ADR states this "removes
the DoS surface from the typechecker" from block execution.

That claim did not hold, for three separate reasons found while auditing the
combined branch.

1. **`MsgRun` was never in scope.** It type-checks *and executes*
   caller-supplied source immediately, under every policy value including
   `inert`. `inert` does nothing to it. So under the policy that exists to keep
   the type checker off the critical path, `MsgRun` puts it straight back.

2. **The inert path charged nothing for the work it deferred.** The branch
   returns before `chargePreprocessGas`, and nothing else in it is priced by
   source length. A submitter could park an arbitrarily large package for the
   price of one amino write and leave the compile bill to whoever enables it.

3. **The oracle could not tell fast code from slow code.** It only checked that
   a package type-checks — which the validator re-checks at
   `MsgEnablePackage` anyway. An oracle that duplicates a check the chain
   already performs protects nothing.

`.app/simulate` made (1) worse than an allowlist alone could fix. It is a public
query that *executes* messages, and the auth ante skipped signature
verification when simulating, so any signer-derived authorization decision read
an attacker-chosen `Caller`.

## Decision

### 1. `run_submitters`: a new vm param, consulted under every policy

`Params.RunSubmitters []crypto.Address`, proto field 18 (#5888 owns 15/16/17).

**An empty list means the gate is OFF: anyone may `MsgRun`.** That is the
behaviour which predates the field, so it is what the zero value has to mean.
Listing one address turns the gate on.

This is the opposite of `CodeSubmitters`, and the asymmetry is the design, not
an inconsistency:

- `CodeSubmitters` is unreachable until an operator has explicitly moved
  `code_submission_policy` to `permissioned`. Its empty state is therefore a
  half-finished opt-in, and refusing is the safe reading.
- `RunSubmitters` has no such switch in front of it — it is consulted on every
  `MsgRun` from the moment the field exists. A fail-closed empty value would
  disable `MsgRun` on every chain that upgrades without editing genesis,
  including `gnoland start` with stock genesis. Because GovDAO proposal
  *creation* is `MsgRun`-only, that takes governance down with it and leaves no
  in-band repair.

An earlier revision of this work did fail closed, and paid for it in exactly the
places that prove the point: `DefaultTestingGenesisConfig`, the txtar
`gnoland start` merge, `adduser`, `adduserfrom` and gnodev all had to seed the
list, and CI stayed green the whole time — the breakage was reachable only on a
real chain. Seeding a fail-closed default across every harness is the smell, not
the fix.

The cost of the current reading is that the param cannot express "nobody may
`MsgRun`". Nobody has asked for that configuration.

The full matrix:

| policy | `vm/add_package` | `vm/run` |
|---|---|---|
| `permissionless` | allow | `run_submitters`, if non-empty |
| `permissioned` | require `code_submitters` | `run_submitters`, if non-empty |
| `inert` | allow (stored inert) | `run_submitters`, if non-empty |

`MsgRun`'s column does not vary because no policy value makes it safe. It is
also the only code-bearing message with no other gate: `MsgAddPackage` clears
`checkNamespacePermission` and `checkCLASignature`, while `MsgRun`'s path is
forced to `/e/<caller>/run`, so there is no namespace to check it against.

The two rules are **siblings**: `checkRunSubmitters` and
`checkCodeSubmissionPolicy`, evaluated independently, neither nested in the
other. That shape is the point. Phase 1 gated `add_package` and `run` together
behind one policy test, so adding `inert` to the enum silently changed what
happened to both — it would have gated `add_package` on `code_submitters` under
`inert`, contradicting the entire point of `inert`. Fixed here as part of the
graft, and pinned: nesting them again, or making the run check conditional on
policy, each fail their own tests.

They share only inputs, not control flow — the params are read once and the
signers scanned once by the caller, so the replay carve-out and the params read
exist in exactly one place. `checkRunSubmitters` deliberately takes no policy
argument, which makes its unconditionality structural rather than a branch a
later reader might "simplify".

### 2. Signature verification when simulating, for gated message types only

`auth.AnteOptions.RequireSigForSimulate func(std.Tx) bool`, with gno.land
passing a predicate that matches **both** code-bearing messages. Bounding the input is not sufficient: without
this, an unauthenticated caller names any address as `Caller` and the allowlist
check above accepts it. Restricted to those two, so keyless gas estimation keeps working for every
message type that carries no source.

It covered `MsgRun` alone until an audit caught the gap. That was right when
`MsgRun` was the only gated message — `code_submission_policy` existed but
nothing enforced it — and adding that enforcement (§1) made `MsgAddPackage`
gated too while leaving the predicate behind. Under `permissioned`, the policy
whose entire purpose is keeping strangers off the type checker, anyone could
name a listed submitter, attach arbitrary bytes as a signature, and drive a full
type-check plus `init()` per query for free.

This required a gnokey change, contrary to first expectations: gnokey rewrites
`GasWanted` to consensus max before simulating, which invalidates the
signature. It now signs a second tx for simulation (`txWithGasWanted`,
`simTxBytes`). An earlier attempt — a flag to skip the rewrite — degraded
`infinite_loop.txtar`, a #3612 regression test that asserts `run -simulate
only` reports against block max.

### 3. Inert submissions are charged up front

`chargePreprocessGas` now runs in the inert branch, and `params` is read once
above the policy branch so both paths decide from the same snapshot. The work
is deferred, not avoided — `MsgEnablePackage` type-checks and runs exactly
those bytes later — so the charge belongs to the submitter, who chose the byte
count. `EnablePackage` deliberately does not charge it a second time.

### 4. `EnablePackage` runs as the creator, not the approver

`OriginCaller` was `msg.Approver`. `init()` runs at enable, and it commonly
is free to record `chain.OriginCaller()` as the owner, so any package that does
would have handed ownership to the approver — and ownership
would have depended on which approver happened to sign. It also diverged from
the non-inert path, where `init()` sees the deployer, meaning identical source
would initialize differently under a different policy.

The creator is read back from the `gnomod.toml` that `AddPackage` wrote before
storing; `genesis.go` already reads it the same way. `SessionAccount` is keyed
on the creator to match, which normally yields nil — correct, since no session
of theirs authorized the enable.

`Height`/`Timestamp` stay at enable time. Only the caller identity is
inherited: `init()` should observe when it actually ran.

### 5. The inert path applies the same gnomod rules as the normal path

`HasReplaces`, the private-override rule, private-must-be-realm, draft, and the
`gno.mod` deprecation were all skipped on the inert branch, and
`EnablePackage` does not check them either. So `inert` was a way to park a
package no policy would ever accept. Extracted to `checkGnomodConstraints` and
called from both paths.

### 5b. Freezing a parked path, bound to the original submitter

An audit found that a parked inert package can be silently overwritten by a
second submission at the same path. The pre-existing duplicate guard consults
`GetPackage`, which sees only *active* packages — a parked one has no
`PackageValue`, and `AddInertPackage` is an unconditional `Set`. Since
`MsgEnablePackage` names a **path, not a content hash**, an approver can verify
submission A and have submission B activated.

This is now fixed, but only after a first attempt that had to be reverted. The
difference between the two is the whole lesson, so both are recorded.

**What ships**: inside the inert branch, a submission over a parked path is
refused *unless the submitter is the address recorded in the parked package's
own stamped `gnomod.toml`*. Two properties of that shape answer the two
objections that killed the first attempt: it is inside the branch, so no
non-inert chain pays for a lookup only `inert` needs; and it is creator-scoped,
so the retry path stays open and a typo cannot destroy a path.

What it does and does not buy is worth being exact about. It stops a **stranger**
from hijacking a path someone else has under review — which was wide open,
because `checkNamespacePermission` returns nil while the names realm is
undeployed, and under `inert` that is precisely the state a chain boots in. It
does **not** stop the submitter's own bait-and-switch: the same creator may
still replace GOOD with EVIL after an approver has read GOOD, because that same
replacement is the legitimate retry after a failed enable. Only a content hash
in `MsgEnablePackage` closes that, and it remains open below.

**What was reverted**, and why it is recorded rather than deleted: the first
attempt refused *any* submission over a parked path, and sat above the policy
branch. Three reasons, in increasing order of importance:

- **It taxed every deploy on every chain.** `GetInertPackage` is a metered read
  of the `main` store, and the guard ran under all policies. Measured: +59,000
  gas on three txtar goldens, exactly one `ReadCostFlat`. Worse, that is a
  paid by every deploy for a lookup only `inert` needs. An earlier draft of
  this section claimed the figure grows with tree depth; it does not. `NewParams`
  pins `FixedGetReadDepth100 = MinGetReadDepth100`, so the charge is exactly one
  `ReadCostFlat` at every tree size.
- **It bricked paths permanently, in the expected case.** `inert` deliberately
  skips type-checking at submit, so an ill-typed package parks happily and then
  fails at `EnablePackage`. `DelInertPackage` only runs after a *successful*
  enable and `DisablePackage` is unimplemented, so nothing clears the key: the
  path can never be submitted to again and never be enabled. A typo destroys a
  package path forever. My commit called that "an inconvenience", which is wrong
  — under a policy whose whole point is deferring the type-check, ill-typed
  submissions are the normal case.
- **It did not actually close the hole.** gpao verifies the mempackage carried
  *in the transaction*, never the bytes stored at the path, and it does not
  consult transaction results. So: land a failing `MsgAddPackage(X, GOOD)` and a
  succeeding `MsgAddPackage(X, EVIL)` in the same block. Only EVIL parks, the
  guard never fires, gpao verifies GOOD and enables X. The swap survives.

The creator-bound guard shipped above avoids the first two failures by
construction. The third it only narrows: gpao still verifies transaction bytes
rather than stored bytes, so the failed-tx variant survives.

The two fixes that would fully work, neither attempted here:

1. **A content hash in `MsgEnablePackage`**, with the keeper refusing a
   mismatch. This is the only one that closes the time-of-check/time-of-use gap
   completely, because it makes the approval name the bytes rather than the
   path. It is a message-schema change and belongs in #5888.
2. **gpao verifying what it queries from the chain**, not what it scraped from a
   block, and consulting transaction results so it ignores reverted submissions.
   That is entirely within the daemon and fixes the failed-tx variant too, but
   it still leaves a window between fetch and enable.

### 5c. The declared deposit ceiling is carried from submit to enable

`inert` is the only path where the message that **declares** the storage-deposit
ceiling is not the message that **spends** against it. `MsgAddPackage.MaxDeposit`
was simply dropped, so `EnablePackage` fell back to `params.DefaultDeposit` —
meaning a creator who declared they would risk 1000ugnot could be charged up to
the chain default (100 GNOT today) by a transaction they neither sign nor can
refuse. That is a consent bug, and it is the one the escrow question below was
really reaching for.

The ceiling now rides in `gnomod.AddPkg.MaxDeposit`, through the same stamped
round-trip that already carries `Creator` — the section is documented as keeper
bookkeeping, not user content, and the field is `omitempty`, so a submission
that declares nothing stores nothing and behaves exactly as before. No new store
key, no new `Store` method, no apphash movement on non-inert chains.

Carrying it also pins the ceiling to submit time, so a governance raise to
`DefaultDeposit` between submit and enable cannot widen what the creator is
exposed to.

### 5d. A parked submission must not delete a live package's source

`AddPackage` clears a private package's stored mempackage blobs before
redeploying it, because the prod / `#allbutprod` pair is not fully replaced by a
re-add. That delete sat **above** the policy branch — and the inert branch
returns without ever calling `AddMemPackage`. So parking a submission over a
live private package deleted the live package's source while its realm, its
objects and its `pkgidx:` entry all survived.

The consequence is worse than a missing blob. At boot,
`PreprocessAllFilesAndSaveBlockNodes` iterates mempackages, and its producer
`IterMemPackage` silently skips a
nil one, so a restarted node rebuilds no `PackageNode` and panics when the realm
is called — while a node that has not restarted keeps answering from
`cacheNodes`, because `SetBlockNode`'s backend write is commented out. That is a
consensus split keyed on restart history, which is exactly the hazard the
surrounding code was written to prevent. The realm's balance and its locked
storage deposit are unrecoverable too, since refunds are driven by
`RealmStorageDiffs()` and require the realm to execute.

The fix is to move the delete below the inert branch, so it only runs on the
path that re-adds a blob.

### 5e. `DefaultDeposit` drops from 600000000ugnot to 100000000ugnot

`DefaultDeposit` is the fallback CEILING applied when a message declares no
`MaxDeposit`. At the default `100ugnot`/byte storage price, 600000000ugnot let a
single message add 6 MB of realm state; the new value allows 1 MB.

The number is measured, not chosen. Booting a real app over all 321 genesis
packages and recording every `StorageDepositEvent`, the largest single deploy is
`gno.land/r/gnoland/boards2/v1` at 276,098 bytes (27,609,800ugnot); median is
6,647 bytes, p95 is 50,324, and only 7 packages exceed 100 KB. The largest
single `MsgCall` growth exercised anywhere in the suite is 500,326 bytes, in
`storage_deposit_price_change.txtar`, which passes an explicit `-max-deposit`
and is therefore unaffected. So 1 MB clears the worst real deploy by 3.6x and
the worst real call by 2x.

For contrast, 10000000ugnot (100 KB) was tried first and fails 41 of the 321
genesis packages — four directly on the deposit ceiling and 37 as cascading
compile failures from the missing dependencies. That is the floor this value
has to stay above.

This is apphash-visible: `SetStruct` writes `vm:p:default_deposit` as its own
genesis store key, so its contents are in the committed root.
`apphash_crossrealm38_test.go` is re-pinned with a note, and
`update_storage_params.txtar` asserts the literal. Gas goldens do not move — the
param's value does not enter any gas computation, only the comparison that
accepts or refuses a deposit.

Note this is independent of §5c: with the declared ceiling now carried from
submit to enable, the inert path consults `DefaultDeposit` only when a submitter
declared nothing.

### 5f. `EnablePackage` refuses to overwrite a live package

Enable is the deferred second half of a deploy, but its entire precondition set
was *the sender is an approver* and *something is parked at this path*. It never
checked whether the path was already live — the check `AddPackage` performs — and
never checked the policy.

That is a package takeover. A path can be parked and live at the same time: the
two live in different key spaces, and nothing clears a parked blob when
governance moves the policy off `inert`. So an attacker parks at a path and is
never approved; the policy later flips to `permissionless`; someone else deploys
at that path for real; and then any approver's routine enable replaces the live
package with the attacker's bytes, running `init()` with the attacker as
`OriginCaller`, which any `init()` that records an owner will treat as one. It
does not even fail loudly: `runMemPackage` takes its fresh-package branch when
`MachineOptions.PkgPath` is empty, so it rebuilds the block node and package
value over the live ones and orphans the previous realm's objects.

`EnablePackage` now applies the same rule as `AddPackage` — a public package may
not be replaced, a private one may — and performs the same `DeleteMemPackage`
before a private redeploy, for the same stale-sibling reason documented in §5d.

Worth noting what this is *not*: it is not a fork. Every node refuses or accepts
identically. It is an authorization gap, and it was invisible to the app hash
for a reason recorded below.

Nor is it the whole of the deploy's preconditions. `EnablePackage` still does not
check the policy, `checkGnomodConstraints`, the namespace, or the CLA — the first
of those is listed as open below. Only the overwrite is closed here.

The liveness probe deliberately reads the stored blob rather than the live
`PackageValue`. Loading the value populates the object cache, and
`RunMemPackage`'s `SetCachePackage` then panics — which made the first version
of the private-replacement branch dead on arrival, failing every time with
"already exists in cache". `AddPackage` reads the value and survives only
incidentally, because `checkNamespacePermission` re-enters
`getGnoTransactionStore`, whose `ClearObjectCache` evicts the entry in between.
The test for the private replacement exists because that branch had none and was
broken.

### 5g. The declared ceiling is stamped unconditionally, and must name the gas denom

The `[addpkg]` section is keeper bookkeeping, but it lives in a file the
submitter authors, so any field the keeper does not overwrite is
attacker-supplied. Stamping `max_deposit` only when the message declared one
left a hand-written value in place, and §5c then read it back at enable as
though the message had declared it — a submitter could write `max_deposit =
"1ugnot"` into their own `gnomod.toml` and block their package's activation with
a ceiling no message ever carried. `stampGnomod` now assigns every `AddPkg`
field unconditionally, including the empty cases.

The same section also refuses a `MaxDeposit` that carries no `ugnot`. Storage
deposits are denominated in the gas denom and `processStorageDeposit` reads only
that component, so a ceiling of, say, `5foocoin` parses cleanly at enable,
contributes nothing, and falls back to `params.DefaultDeposit` — read at enable
time, which is exactly the pinning §5c exists to provide. Honouring it in name
only would have made §5c's guarantee conditional on a denom nobody checks.

### 6. The oracle gets a wall-clock budget, on its own goroutine

`contribs/gpao` gains `--verify-budget` (default 10s), covering **both** stages
the validator re-runs: type-check and preprocess. Wall-clock is not a consensus
quantity, so "this finishes quickly" is exactly the claim only an off-chain
actor can make — and therefore the only claim worth an oracle. The default is
generous on purpose: a real package verifies in milliseconds, and a borderline
one should pass rather than lose a race with CPU contention.

Preprocess uses `Machine.PreprocessFiles`, not `RunMemPackage`: the latter
executes `init()`, which off-chain and unmetered means a hostile package could
hang the daemon permanently. It runs against a discarded transaction store, so
nothing a candidate does is persisted or visible to the next one.

Two earlier drafts of this work reported preprocess as blocked, claiming
`PreprocessFiles` panicked with an internal "should not happen" that needed more
machine setup than was available. **That was a misdiagnosis and is retracted.**
The panic was `AddMemPackage` rejecting an *unsorted file list* in the test
fixture — the type-check path does not validate file order, so the fault stayed
invisible until preprocess exercised it. The stage needed no new machinery at
all, only the `lint.go` pattern and a correctly ordered fixture.

Because the preprocessor resolves imports through the store rather than through
the type-check's `hybridGetter`, chain-domain imports are seeded into that store
first, transitively, over RPC where the disk cannot supply them. If something
remains unresolvable the stage is **skipped with a log**, not turned into a
rejection: refusing would regress behavior (such a package was previously
approved on the type-check alone) and would refuse for a limitation of the
oracle rather than a fault of the package. What is not acceptable is doing that
silently, so it is logged.

That preprocess catches something the type-check cannot is pinned by a test: a
pure package declaring a crossing function (`cur realm`) is well-typed to
go/types but rejected by the preprocessor, because "crossing functions only
exist in realms" has no Go equivalent. Verified by removing the stage, at which
point that test — and only that test — fails.

The budget is **enforced, by running verification in a child process.**

An earlier revision of this branch could only *measure* it. Go cannot kill a
goroutine, so an in-process deadline can only abandon the work — which leaves it
running, still spending the CPU the budget was meant to bound, and still
mutating stores the next attempt reads. That version waited for the work to
finish and then reported that it had taken too long, which bounds nothing.

A process can be killed, so `exec.CommandContext` turns the budget into a real
deadline. `TestOracleVerifyBudgetKillsTheChild` pins it on a heartbeat rather
than on elapsed time: the child writes a heartbeat file in a loop, and the test
asserts the heartbeat STOPS after the budget expires. (It also carries a
generous elapsed-time backstop, but that bound is not what distinguishes the
implementations.) An implementation that abandons
the wait instead of killing the process fails it, which is the distinction that
matters and the one an elapsed-time assertion cannot make.

Three further properties come from the process boundary rather than from code:

- **Per-attempt isolation is free.** Each child builds its own stores and cache,
  so no state survives to be raced. That is what previously forced verification
  onto a single goroutine, and an earlier draft of this ADR gave the wrong reason
  for that constraint (it blamed the typecheck cache, when `LoadImports` writing
  to a shared store was the larger part). Both are moot now.
- **A crash is contained.** The typechecker and preprocessor report errors by
  panicking, and a native-code crash or OOM would otherwise take the daemon with
  it. Now it takes the child down and reads as a rejection.
- **The approver key never enters the process that compiles untrusted code.** The
  child has no signer and no keystore, so it cannot approve anything.

Verification still runs one at a time, but no longer for safety — only because
it is the expensive thing this daemon does, and running several at once would
make each slower and the budget harder to interpret. It stays off the
block-reading loop so a burst of candidates cannot stall chain following.

A budget overrun is reported distinctly and is **not** a rejection, because the
two deserve opposite treatment: a rejected package is settled, a slow one may
have lost a race with whatever else the machine was doing. Exit status carries
the verdict — clean exit passes, non-zero exit is a rejection with the child's
stderr as the reason, and a deadline is neither. The overrun cap still applies,
since resubmitting is cheap under `inert`.

Dependency preloading and store construction now sit *inside* the measured
window, because they happen in the child. An earlier revision kept them outside,
which was right when the stores were long-lived and warmed across candidates;
with a process per candidate every one pays the full cold cost. That is a real
charge for work the chain does not do — on the order of a second here — and it
is the main reason the default budget is 10s rather than tight.

### 7. Not done: cross-field validation of the policy/list pairings

An earlier version of this branch made `Params.Validate` reject `permissioned`
with an empty `code_submitters`, and `inert` with an empty `pkg_approvers`,
calling them brick configurations. **That was wrong and has been reverted.**

The justification was that restoring the lost capability requires the
capability. It does not, and the reason is this branch's own central change:
`MsgRun` is gated on `run_submitters`, not on `code_submitters`. So with
`permissioned` and nobody listed, deploys stop but governance does not — the
`MsgRun` → propose → vote → execute loop is fully intact, and the grafted
`code_submission_policy.txtar` demonstrates exactly that, with an account
refused an `addpkg` still running successfully. #5888's `inert` case is the
same. Separating the two lists *removed* the self-lockout hazard that #5885
documented an ordering caveat to work around; the validation then guarded
against a lockout that no longer existed.

Three further costs, none of which the original reasoning accounted for:

- **It broke genesis import.** `ValidateGenesis` calls `Validate` and
  `InitGenesis` panics, so a chain that legitimately held that pairing under
  #5885 or #5888 would export a genesis this binary refuses to boot.
- **It blocked a legitimate exit.** Removing the *last* approver — revoking a
  compromised key — panics inside `WillSetParam` after the vote has already
  passed, and the govdao executor has no recover, so the transaction aborts.
- **It was circumventable anyway.** `validateAddressSlice` rejects only zero and
  duplicate addresses, so a list holding one address nobody has a key for is
  operationally identical to an empty one and validates cleanly. The rule
  refused the honest spelling of the state, not the state.

The sharpest form of the objection: those pairings are only *reachable from* a
chain on which param changes work, which is precisely the condition that makes
them recoverable. The predicate's entry condition and its exit condition are
the same condition.

Both are also plausibly deliberate — `permissioned` with nobody listed is the
only vm-level way to express an emergency deploy freeze, and `inert` with no
approver yet is the natural state while adopting an oracle.

And it validated the wrong thing. At the time this was written the one
configuration it should have worried about was `run_submitters` empty — which
`Validate` cannot reject, because it is the default. §1 has since removed that
hazard by making the empty list mean "gate off" rather than "nobody", so there
is now no unrecoverable pairing for a cross-field rule to catch at all. A rule
that refuses two recoverable states and cannot touch the unrecoverable one did
not track the property it claimed to enforce; today it would refuse two
recoverable states and nothing else.

If a fat-finger guard is wanted, the defensible form is a *transition* check in
`WillSetParam` — which alone sees both the old and the new value — refusing a
flip *into* a restrictive policy while the corresponding list is empty, and
leaving the end states themselves legal. That is a separate change, and since
#5885 permits these pairings — its `Validate` deliberately imposes no
cross-field rule — changing that verdict belongs in its own PR rather than in a
branch stacked on top of it. (No test asserts the pairing either way; the
permission is by absence of a check, not by assertion.)

### 8. Not done: making the submitter pay for `init()` at enable

Under `inert` the approver's transaction supplies the gas meter for a parked
package's `init()`, so the submitter chooses the cost and the approver -- in
practice an unattended oracle -- supplies the budget. This was designed,
implemented, and abandoned. It is recorded because the obvious fixes each look
right and are not.

**The money problem is smaller than it looks.** tm2 fees are flat: the ante
deducts `tx.Fee.GasFee` in full regardless of gas used, so the approver's cost
per enable never scaled with `init()` to begin with. What remains is that the
approver must keep `GasWanted` high enough for the largest legitimate `init()`,
and the minimum fee scales with that -- worth roughly a third of the flat fee
gpao already volunteers per enable.

**Recording the declared gas limit does not prove payment.**
`EnsureSufficientMempoolFees` is what ties `GasFee` to `GasWanted`, and the ante
runs it only under `ctx.IsCheckTx() && !simulate`; its own contract says it
"cannot be part of consensus". `std.Tx.ValidateBasic` imposes no relation
between the two and `minGasPrices` is node-local. So on the consensus path a
submitter reaching a proposer directly can declare `Block.MaxGas` for a nominal
fee. Anyone tempted to read `GasWanted` as evidence of payment should stop here.

**Consuming the reservation instead does not survive contact either.**

- The store's gas context and meter are read once and frozen into the gno store
  by `BeginTransaction`, before any handler runs. `ctx.WithGasMeter` copies a
  Context field and cannot reach them, so storage I/O during `init()` still
  bills the approver; only `MachineOptions.GasMeter` and the allocator move.
  Storage dominates a realm-creating `init()`, so the split puts the small half
  on the reservation. The existing `ctx.WithGasMeter(store.NewGasMeter(n))`
  sites are not precedent: every one is a query path that then builds a
  throwaway store. `EnablePackage` must use the committed store.
- `NewMachineWithOptions` mutates the store-owned allocator's gas meter and
  nothing restores it, so after `init()` returns the storage-deposit charge and
  every later allocation bill the side meter.
- Work on a side meter is invisible to `BlockGasMeter`, which is fed post-hoc
  from the transaction's meter. That removes it from the block bound *and* from
  the fee market: `UpdateGasPrice` reads `BlockGasMeter().GasConsumed()` and
  treats a low reading as the strongest "under target" signal, so a block
  saturated with off-meter `init()` looks idle and pushes the gas price down.
- It does not fix the drain. A one-gas reservation on an expensive `init()`
  still costs the approver a full flat fee when the enable fails, which is the
  same rate as before.
- Charging a declared figure with `ConsumeGas` would be the first primitive here
  that burns gas without doing work, so one small transaction could consume a
  whole block's allowance -- and the mempool reaps against *declared* gas, so it
  reserves the block before executing.
- Recording the transaction's limit rather than a per-message figure over-issues
  by the message count, since nothing caps messages per transaction and they
  share one meter.

**`gnomod.toml` is the wrong home regardless.** §5g exists because a
hand-written `max_deposit` was once honoured, and its fix was the invariant that
`stampGnomod` assigns every field unconditionally so nothing user-authored
survives. A submitter-declared `init_gas` read back out of the same section
breaks that invariant in the section that already sprang once.

**What to do instead.** A chain-wide cap is the cheap option -- one `vm` param
and a passthrough meter at the machine construction, which keeps every unit
visible to the block meter and needs no new state. But note what it bounds: the
value reaches only the machine's meter and the allocator, so it caps interpreter
work and not storage, for the same frozen-meter reason above. It bounds the VM
half; it does not bound the approver's exposure. Alternatively move the budget
off chain, extending gpao's verifier child from `PreprocessFiles` to a bounded
full run -- §6 made that safe by turning the budget into a real deadline
enforced by killing a process -- so the fee is never spent on a package that
would fail.

The orthogonal fix, if "the submitter pays" is ever made a requirement, is to
stop having one party execute another's code: split enable into an approval
(approver, O(1), binding a content hash) and an activation sent by the creator
on an ordinary transaction and an ordinary meter. That makes declarer and
spender the same party, and would retire §5c's ceiling carry, §5g's stamping
defence, and the content-hash gap item 12 lists as open.

## Alternatives considered

**Escrow the deposit ceiling at submit and refund the remainder at enable.**
Designed in full, reviewed, and declined. The problem it targets is real — a
creator can drain their account between submit and enable, so the enable fails —
but it is not the problem it looks like. A failed enable is *retryable*: the tx
boundary discards the writes, the package stays parked, and the approver can
enable again once the creator is funded. The only loss is the approver's gas on
one transaction, and `MsgEnablePackage` can be simulated before broadcast.

Against that, escrow would immobilise the ceiling — not a quote, a ceiling: the
sample package here needs 210,200ugnot against a 100,000,000ugnot default cap —
on *every* submission, including the ones a review gate is expected to leave
unapproved, with no cancel message and `DisablePackage` unimplemented. It also
does not deliver the guarantee it promises: `processStorageDeposit` reads
`params.StoragePrice` at enable time, so a governance price rise between submit
and enable reopens the same failure with the funds already locked. And it would
need a distinct derived address, since `DeriveStorageDepositCryptoAddr` is the
live realm's own deposit pool — refunding by that address's *balance* rather
than a recorded amount would have been a theft primitive against a live realm.

What survived from the exercise is §5c: the actual defect was not that the
deposit is taken late, but that the creator's declared ceiling was thrown away
before it could bind the charge.

**Reuse `CodeSubmitters` for `MsgRun` instead of a new param.** Rejected. It
would make one list mean different things per policy, and an operator who
populated it purely to unblock `MsgRun` under `inert` would silently gain
deploy rights the moment governance flipped to `permissioned` — privilege
expansion with no separate governance act and no separate review. It also
cannot express run-set ⊂ deploy-set, which is the operationally desirable
shape: the deploy set is naturally large (every onboarded, CLA-signed dev in
their own namespace) while the run set should be small. Address lists in this
codebase mostly sit one-per-capability, and the exceptions are informative
rather than reassuring: `PkgApprovers` in this very struct gates both
`MsgEnablePackage` and `MsgDisablePackage`, and `r/sys/users`' `controllers`
gates five operations at once. Sharing a list is what makes a later capability
split a migration instead of a param change.

**Enforce in `VMKeeper.Run` as well as the ante.** Deferred, with the tradeoff
recorded because it is a close call. For: the vm module's other authorization
(namespace, CLA, `PkgApprovers`) lives in the keeper, and a keeper check cannot
be lost by app wiring. Against: the ante is the only layer that can refuse
during CheckTx and keep the tx out of the mempool; it is where gno.land's other
signer-derived policy already lives; and a second gate needs its own replay
carve-out, which does not fail in tests when it is missed — it fails the next
time somebody forks the chain. One gate, one carve-out. A reviewer who
disagrees can add the keeper check cheaply; the helper is already factored.

**Bound the qeval/simulate input instead of verifying signatures.** Rejected on
owner instruction and on the merits: gas cannot bound it either, since
`maxGasQuery` is 3e9 ≈ 3s of CPU by design.

**Charge the deferred compile at enable rather than at submit.** Rejected:
it bills the approver for a byte count the submitter chose.

## Consequences

### Bootstrap is not a hazard, because empty means open

A chain boots with `run_submitters` empty, which means the gate is off. Nothing
has to be seeded for governance to work, and `gnoland start` with stock genesis
behaves exactly as it did before this param existed.

Turning the gate on is a deliberate act, and the ordering follows from one fact:
governance *proposal creation* is `MsgRun`-only.

`isValidCall` in `r/gov/dao/v3/impl/govdao.gno` admits a caller only when the
previous realm is a user or the `/e/<addr>/run` ephemeral realm; a deployed
realm's `init()` frame is neither, and that exclusion is deliberate — it is what
prevents vote-laundering through an intermediary realm. Nor can `MsgCall`
substitute: `ProposalRequest`'s fields are unexported and it carries an
`Executor` interface over a `func(realm) error`, while `convertArgToGno` panics
on struct, interface and func arguments, and the keeper synthesizes exactly one
call with no composition.

So whoever populates the list must keep at least one address that can create
proposals, or the *next* change to this param becomes unproposable. Both routes
work:

- **At genesis**, `gnogenesis params set vm.run_submitters <addr>,...`, via the
  reflect-based setter. (The declarative route, `genesis_params.toml` through
  `LoadGenesisParamsFile`, cannot set it — that function accepts only
  `chain_domain` and `sysnames_pkgpath` and carries its own `XXX Write onto ggs
  for other keeper params`. It is knowingly incomplete for every vm param, and
  special-casing the newest one would add a one-off to a function that needs
  generalizing instead.)
- **On a running chain**, a governance proposal, created while the list is still
  empty and therefore permissive. `run_submitters.txtar` walks exactly this
  path: open chain, vote, gate on, stranger refused.

The delegated manager (`r/sys/params`) cannot undo it. It may only add, may only
remove what it added, and may never take the list to zero — see
`RemoveRunSubmitters`. Reaching zero would switch the gate off entirely, which
is not a smaller version of removing one address but the revocation of the whole
restriction GovDAO voted for.

### Test chains are NOT seeded, deliberately

The harness leaves `run_submitters` at its default. `DefaultTestingGenesisConfig`,
the txtar `gnoland start` merge, `adduser`, `adduserfrom` and gnodev all use the
stock VM genesis state, so every existing txtar keeps working unchanged and none
of them is quietly testing a pre-seeded allowlist.

The two txtars that exercise the gate populate it in-script, by governance
proposal, which also documents the bootstrap order above.

One unrelated fix in that area is kept: the `gnoland start` handler now carries
`tsGenesis.VM.Params.ChainDomain` and `.SysNamesPkgPath`. Those are the only two
scalar vm params `LoadGenesisParamsFile` writes, and the merge previously
dropped them — harmless only because the file happens to set what the defaults
already are, so a test would have passed while a real chain used the file's
value. `TestGenesisParamsReachTheHarness` fails if a third field is added to the
loader without being carried here.

Production defaults are untouched.

### Each row of the matrix needs its own refusal test

Every txtar in the suite signs as an account nobody listed, on a chain whose
allowlist is empty — so all of them pass whether the gate runs or not. A test
that expects a **refusal** is the only kind that distinguishes "the gate
authorized this signer" from "there is no gate", and there has to be one per
row.

`run_submitters.txtar` covers the run row end to end, and covers both readings
of the param in the order a chain meets them: empty, a stranger runs;
populated by governance vote, the same stranger is refused. The vote in the
middle is not scaffolding — it is the bootstrap order, and it only works because
the list starts permissive.

This was verified by deleting the `checkCodePolicy` call from the ante closure.
The txtar fails, and so does one Go test — `app_test.go`'s "historical MsgRun
replays under a run_submitters list it is not on", which sends a live MsgRun and
asserts the refusal names `run_submitters`. An earlier draft of this section
claimed every Go test still passed; that was wrong, and the correct version is a
stronger result rather than a weaker one.

The refused account cannot come from `adduser`/`adduserfrom`, since those write
a genesis balance and this has to be an account the seeding proposal never
named; the txtar recovers a key from a fixed mnemonic and funds it by transfer
instead.

The same reasoning applies per row, and the rows are not interchangeable.
`run_submitters.txtar` covers `vm/run`; `code_submission_policy.txtar` (grafted
from #5885) covers `vm/add_package` under `permissioned`. It is the only
*integration* test that does: deleting `case vm.MsgAddPackage` from the ante's
message scan leaves `run_submitters.txtar` green and fails that file. It is not
the only test at all — `TestTxCodeMsgSigners`, `TestTxCarriesCode` and
`code_policy_test.go` all drive the same scan over real messages and fail too.
An earlier draft claimed sole coverage; the uniqueness is only at the
integration level.

That file also shows the sensible ordering — populate `code_submitters`, then
flip the policy — is reachable through ordinary GovDAO proposals. Nothing
enforces that order: the cross-field validation that would have is reverted (see
§7). Doing it the other way is a deploy outage until a second proposal lands,
not a deadlock.

### The inert row is covered end to end, in Go rather than txtar

The two txtars above cover refusals. The `inert` row is not a refusal — it is a
lifecycle, submit → parked → enable — and it had no test that crossed a
transaction boundary at all. The keeper's own tests call `AddPackage` and
`EnablePackage` directly, which skips the ante handler, the tx boundary, and
the commit; §3 and §4 of this document are both claims about behavior those
tests cannot see.

`TestInertPackageLifecycleEndToEnd` (`gno.land/pkg/gnoland`) closes that. It
runs a chain with `code_submission_policy = inert`, one committed block per
transaction, and four distinct accounts — submitter, approver, and two callers —
so a cost can be attributed to a party rather than to "whoever signed". Account
numbers and sequences are read back from the chain before each signature rather
than assumed; assuming them silently produced a passing "must be rejected"
assertion during development, where the rejection was a signature error and the
behavior under test never ran.

It is a Go test and not a txtar for a reason worth recording on its own:
**`MsgEnablePackage` has no `gnokey` subcommand.** The message has a handler and
an amino registration, but no CLI route, so a txtar script cannot send one. The
only ways to enable a package today are a programmatic client (which is what
gpao is) or a hand-assembled transaction.

The two §4/§5c claims are mutation-verified against the keeper. §3's
submit-time preprocess charge is *not* covered by this test:

| Mutation | Assertions that fire |
|---|---|
| `OriginCaller: msg.Approver.Bech32()` instead of the creator | both identity assertions |
| `processStorageDeposit(ctx, msg.Approver, ...)` instead of the creator | both balance assertions |

One negative result came out of writing it. A call to a package that is stored
but not enabled fails — correctly — but by way of a generic VM panic,
`unexpected node with location <path>:0:0`. That is not an inert-specific error
path: a package that was never submitted at all produces the identical message.
So "the parked package is not callable" cannot, by itself, distinguish parked
from absent, and the test says so in a comment. What closes the gap is that
`EnablePackage` refuses with `no inert package at path` when nothing is stored,
so the enable succeeding is the proof that submit persisted the package. The
two steps are load-bearing together and neither should be deleted alone.

### Genesis replay is exempt, deliberately

`deliverGenesisTx` replays historical txs through this same ante with
`BlockHeight > 0`, *after* `InitGenesis` has installed the new params. Without
the carve-out a hardfork would refuse to replay its own history the moment
either list omits a historical signer, and with `StrictReplay` the node would
not boot. Keyed on `auth.GenesisReplayKey`, not `BlockHeight() == 0`, because
forked txs carry their original heights.

This is a real hole — replayed history is not re-authorized — and it is stated
here rather than left to be discovered.

### Authorization is per message, not per transaction

The signers that must hold code-submission rights are the signers of the *code
messages*, not every signer of the transaction. An earlier version of this
branch used `tx.GetSigners()`, which reads naturally and is wrong: a tx bundling
a `MsgRun` with, say, a bank send signed by someone else refused the whole tx
and reported the bystander as "not authorized to send MsgRun" when they had sent
no such thing. #5885 had this right and the mistake was introduced by
rewriting rather than grafting its tests — found only by going back and diffing
the two test suites case by case.

A consequence worth stating, since it is the visible difference from #5885: an
address on `code_submitters` is not thereby allowed to run, and vice versa. The
two lists are independent, so a tx whose `MsgAddPackage` signer is authorized
and whose `MsgRun` signer is not is refused — for the run, naming the run list.

### Sessions are authorized through their master, in two layers

An earlier draft of this ADR said the opposite — that a session key must appear
in `run_submitters` itself, and called that fail-closed. **That is wrong and is
retracted.** `tx.GetSigners()` unions `msg.GetSigners()`, and
`MsgRun.GetSigners()` returns `msg.Caller`, the *master* address; the session
key travels in `Signature.SessionAddr`, a field this check never reads. The
contract on `std.SessionAccountsContextKey` says so in as many words: the map is
keyed on "the master account address returned by `msg.GetSigners()`, NOT the
session pubkey address."

So the real behavior is a two-layer grant. A session's `AllowPaths` decides
whether it may carry a `MsgRun` at all; `run_submitters` decides whether its
master may run code. A session key held by a listed master inherits that
authorization, and listing a session address in `run_submitters` would do
nothing.

That is defensible — arguably better than the narrow reading, since it keeps one
answer to "may this account run code" — but it was never chosen, it fell out of
where the check reads its identity. Narrowing it would mean authorizing from
`Signature.SessionAddr` rather than from `tx.GetSigners()`, which is a different
design and is not attempted here.

Note `vm/run` is grantable to sessions today, while `vm/add_package` is in
`sessionAlwaysDenied` as privilege escalation. Existing coverage does not
distinguish the two readings: `gnoclient`'s session test sends a session-signed
`MsgRun` and passes, but only because its master is the one address test genesis
seeds.

### The apphash moves

A new `Params` field moves the genesis root even with an empty default:
`encodeStructFields` writes one store key per field unconditionally and does not
skip zero values. Gas goldens do **not** move, and the mechanism is worth
recording because the intuition says otherwise — the ante's whole-struct
`GetParams` runs on the throwaway `NewInfiniteGasMeter` that `sdk.NewContext`
installs before `auth.SetGasMeter` replaces it, and baseapp uses a single cache
wrap across ante and messages, so every later `GetParams` is a cache hit.

### The DoS is narrowed, not closed

`MsgAddPackage` under `permissionless` still reaches the same type checker with
the same per-byte-only charge, and it is reachable by any funded address (the
namespace rule grants everyone `{r,p}/<own-address>/*`, and CLA signing is a
permissionless `MsgCall`). `MsgRun` is singled out for the *allowlist* because it is the only
code-bearing message with no other gate. For the SIMULATE path both are
covered — an earlier draft claimed `MsgRun` was "the only
unauthenticated-identity, immediate-execution path", which was false:
`MsgAddPackage` in simulate is exactly that too, and §2 now closes it.
Closing the `add_package` row for real transactions still means running under
`permissioned` or `inert`.

## Still open

1. **Run `init()` in the verifier.** The oracle certifies type-check and
   preprocess but never executes the package, because an unmetered off-chain
   `init()` from a stranger could hang the daemon — so `PreprocessFiles` is used
   rather than `RunMemPackage`. The chain DOES run `init()` at
   `MsgEnablePackage`, so its cost is currently uncertified. The subprocess
   makes this tractable for the first time (a hostile `init()` is now just a
   process to kill), which is a reason to revisit it, not a reason it is done.

2. **`DisablePackage`** remains unimplemented in #5888 (it needs object
   eviction), so a package can be enabled exactly once — and the storage deposit
   taken at enable is correspondingly never released.

3. **Move the policy decision into the `vm` module.** `codePolicyResult` lives in
   `gnoland` but encodes the semantics of three `vm.Params` fields, so a reader
   of `params.go` has no path to the rule. The better shape is an exported
   `vm.CheckCodePolicy(params, signers, ...)`, with the ante keeping only the
   *when* — mirroring how `RequireSigForSimulate` puts the generic hook in tm2
   and the gno.land-specific predicate at the call site. Enforcement stays in
   the ante either way; only ownership of the semantics moves.

4. **`pkg_approvers` is enforced in the keeper, not the ante.** The argument used
   here for the ante — it is the only layer that can refuse during CheckTx —
   applies to `MsgEnablePackage` just as well, and that is the message now
   carrying the expensive type-check. The codebase currently asserts both
   positions at once. Worth resolving deliberately.

5. **The replay carve-out is per-check, not per-delivery.** "This delivery is
   replayed history" is a property of the delivery, yet each check hooks
   `GenesisReplayKey` itself — and `checkSessionRestrictions` has no carve-out
   at all, so a replayed session tx *is* re-authorized against current params.
   Extracting `isGenesisReplay(ctx)` and deciding that asymmetry belongs in a
   follow-up. Note also that the auth ante's own use of this key requires a
   second condition (the operator's `--skip-genesis-sig-verification`); this
   carve-out fires on the key alone.

6. **Split the policy enum.** `permissioned` is an authorization rule and
   `inert` is a processing mode, so the enum cannot express "only trusted
   submitters, and their packages still need approval" — the obvious posture for
   a cautious chain. `run_submitters` having to escape the enum entirely is the
   diagnostic. Splitting it into an allowlist set plus a `defer_typecheck` flag
   is a proto change and belongs in its own PR.

7. **`gnokey --simulate only` now transmits a validly signed transaction.**
   Signing the simulation tx was necessary (the `GasWanted` rewrite invalidates
   the original signature), but it means a dry run hands the RPC endpoint
   something broadcastable. A hostile endpoint could submit the messages the
   user asked not to execute. Partially mitigated by the fee/gas ratio check in
   CheckTx, which does not apply in DeliverTx. This widens a hole that already
   existed for `maxGas == 0`, rather than opening a new one.

8. **Fixed, and it was understated.** gpao now consults `/block_results` and
   ignores transactions that failed, refusing to act on a block whose outcomes
   it cannot read. The original entry called this "cheap denial-of-approval". It
   was also a **budget bypass**, which is worse, because the budget is the only
   thing this daemon adds: park slow but valid bytes at a path, then send a
   second `MsgAddPackage` at the same path with gas too low to survive the
   preprocess charge. It fails and changes nothing, yet it sits in the block — so
   gpao timed the trivial bytes and approved a path whose stored contents were
   the slow ones, making the enable it signs the vehicle for exactly the
   unmetered compile it exists to prevent.

   The other half is still open: the submitter may replace parked bytes between
   verification and enable, because `MsgEnablePackage` names a path rather than a
   content hash. See item 12. Blocks that fail to fetch are also no longer
   skipped — the height is retried rather than advanced past.

9. **The txtar genesis merge is a hand-maintained field list, now guarded.**
   `gnoland start` copies named fields, so a field set in
   `gno.land/genesis/genesis_params.toml` but absent from the merge is silently
   replaced by the default — a txtar test passes while a real chain, which reads
   the file directly, uses a different value.

   That was already true of `chain_domain` and `sysnames_pkgpath`, and harmless
   only by luck: the file sets them to exactly what the defaults are. Both are
   now carried, and `TestGenesisParamsReachTheHarness` diffs a state loaded from
   a deliberately non-default fixture against a pristine default, so a param the
   loader starts handling shows up on its own instead of having to be
   remembered. Verified by adding a case to the loader and watching the test
   name the uncarried field.

   The structural fix — one merge function, or one genesis state instead of two —
   is still worth doing; this only makes the drift loud.

10. **`MsgEnablePackage` has no `gnokey` route.** It can only be sent by a
    programmatic client or a hand-built tx, which is why the inert lifecycle
    test is Go rather than txtar. An approver operating by hand has no
    supported path today; `gnokey maketx enablepkg` belongs with the message
    itself (#5888) rather than here, but the gap should not go unrecorded.

11. **Calling an absent package panics instead of erroring.** `MsgCall` to a
    path with no enabled package dies on `unexpected node with location
    <path>:0:0` — an internal VM message, not a package-not-found error, and
    identical whether the package is parked or was never submitted. Not caused
    by this PR and not specific to `inert` (verified against a never-submitted
    path), but `inert` makes the parked case ordinary rather than exotic: a
    caller who arrives before the approver does now gets this. Fixing it means
    touching the generic call path, so it is called out rather than folded in.

12. **A content hash in `MsgEnablePackage` is still needed.** The creator-bound
    guard in §5b stops a stranger from swapping bytes under a reviewed path, but
    not the submitter themselves: the same creator may replace GOOD with EVIL
    after an approver has read GOOD, because that replacement is also the
    legitimate retry after a failed enable. Approval names a path, not bytes.
    This is the single largest remaining hole in the inert flow.

13. **Nothing can delete a parked package.** `DelInertPackage` runs only after a
    successful enable, `DisablePackage` is unimplemented, and there is no reject
    or expiry message. A submission an approver declines occupies IAVL forever.
    It is paid for — roughly 1.27ugnot/byte in non-refundable gas at submit, so
    it is not free state — but it is unreclaimable. Nor is an ordinary live
    package's blob reclaimable: `DeleteMemPackage` is reachable only on a
    private redeploy and no message deletes one. What is specific to `inert` is
    unreclaimable state that was never even usable.

14. **Fixed: `msg.Send` is refused on an inert submission.** The coins used to
    move to the package address at submit, but `EnablePackage` builds its
    `ExecContext` with an empty `OriginSend` and no `OriginSendRecipient` — the
    only two fields that differed from the ordinary deploy path — so a payable
    `init()` did not merely see an empty envelope, it panicked on the recipient
    mismatch. Identical source deployed under `permissionless` and failed under
    `inert`, which made a governance parameter change program semantics.

    `AddPackage` now refuses a non-zero `Send` on the inert branch, and sends no
    coins there at all. Carrying the envelope through to enable was the
    alternative and was rejected: it means stamping the amount into
    `gnomod.toml` beside the deposit ceiling and reconstructing an origin-send
    context for coins that moved in a different transaction, at a different
    height, possibly under a different account state — a lot of machinery to
    make a two-phase deploy impersonate a one-phase one. Refusing costs the
    submitter one transfer after activation, and it also stops coins being
    stranded at the address of a package that is never approved.
    (`TestVMKeeperInertRefusesPayableSubmission`.)

    The related type-check-mode asymmetry is now moot: `AddPackage` drops to
    `TCGenesisStrict` at height 0 while `EnablePackage` is unconditionally
    `TCLatestStrict`, but nothing can be parked at genesis (item 21) and enable
    now requires the `inert` policy (item 16), so a genesis-time enable has
    nothing to act on.

15. **Fixed: genesis replay ignores the policy.** `InitGenesis` runs before the
    replay loop, so a fork that turned `code_submission_policy` to `inert`
    re-executed every historical `MsgAddPackage` down the inert branch: packages
    that deployed live on the source chain parked on the fork, and the chain
    came up with its own realms missing — silently, until something called one.

    `AddPackage`'s inert branch now also tests `!isGenesisReplay(ctx)`, keyed on
    the same `auth.GenesisReplayKey{}` context value the ante's carve-out uses.
    Two carve-outs, because they guard two different decisions: the ante decides
    whether the tx is admitted, this decides what the message does. Replay must
    reproduce what the source chain did, whatever policy the fork adopts going
    forward. (`TestVMKeeperGenesisReplayIgnoresInertPolicy`.)

    That carve-out has a consequence on the other side of the split, now also
    fixed. Because a replayed submission goes live instead of parking, the
    matching `MsgEnablePackage` -- also in the replayed history -- finds nothing
    parked. It has to succeed anyway: a failed genesis transaction under
    `StrictReplay` is a node that will not boot, so a fork of a chain that
    genuinely ran `inert` could not start. `EnablePackage` now returns early
    during replay, which also exempts the policy and approver gates, for the
    same reason the ante exempts its own: replay runs after `InitGenesis` has
    installed the fork's params, so a fork that moves off `inert` or rotates
    `pkg_approvers` would otherwise refuse its own history. The mandate those
    gates protect was exercised on the source chain; replay reproduces that
    record rather than granting it again.
    (`TestVMKeeperGenesisReplayEnableIsANoOp`.)

15b. **Still open: replay reproduces the fork's policy, not the source chain's.
    Planned fix below.**

    The carve-out in item 15 is right for a fork that turns `inert` ON, and
    wrong for a fork OF a chain that already ran it. On such a chain
    `MsgAddPackage` filed the code away without compiling it -- deferring the
    compile is the whole point of the policy, so uncompilable parked code is
    ordinary, not exotic. Replay now takes the ordinary path for those same
    transactions, compiles them for the first time, and they fail. A failed
    genesis transaction under `StrictReplay` is a node that will not boot, so
    such a chain cannot fork itself. Reproduced: the same package parks cleanly
    on the source chain and fails replay with "invalid gno package; type check
    failed".

    Two more in the same family. The documented same-creator re-parking retry
    path replays as v1 going live and v2 hitting "package already exists". And a
    submission the source chain's approver reviewed and turned down deploys live
    on the fork -- the takeover §5f exists to prevent, reintroduced through
    replay.

    Neither keeping nor removing the carve-out fixes this; each is right for one
    fork direction and wrong for the other. The branch has to follow the policy
    the source chain had **at that transaction's height**, and nothing records
    that today.

    **Carry it per transaction, not per chain.** `GnoTxMetadata` already carries
    what the source chain did for each replayed transaction -- `GasUsed`,
    `GasWanted`, `Source`, `Note` -- populated by `gnogenesis fork generate` at
    assembly time and unused in normal operation. A source policy belongs in
    exactly that set. Per-transaction also handles a history whose policy changed
    partway, which a single pinned value cannot.

    The alternative considered and rejected was a policy-epoch list on
    `vm.GenesisState`. It needs a new protobuf message type rather than one
    scalar field, it cannot express a policy that changed mid-history without
    interval logic, and it puts chain-shape data in a keeper's genesis rather
    than with the transactions it describes.

    Touch points:

    - `gnoland.proto` / `types.go` / `pb3_gen.go`: add
      `CodeSubmissionPolicy string` to `GnoTxMetadata`. One scalar, no new
      message. `pb3_gen.go` is hand-maintained in this tree, so this is an edit
      rather than a regeneration.
    - `contribs/gnogenesis/internal/fork/generate.go`: populate it, beside the
      existing `GasUsed`/`GasWanted` provenance.
    - `deliverGenesisTx`: put it in the context, alongside the
      `GenesisReplayKey` it already sets.
    - `AddPackage`: during replay, take the policy from the context instead of
      `params`, and drop `!auth.IsGenesisReplay(ctx)` from the branch condition.
      The branch then parks exactly when the source chain parked.
    - `EnablePackage`: same source-policy read for its policy gate, and keep the
      approver exemption -- a fork may legitimately have rotated
      `pkg_approvers`, and replay is reproducing a record rather than granting
      it again. The early return added for item 15's knock-on can go: once
      replay parks correctly, the replayed enable finds its package and does its
      ordinary work.

    Absent metadata -- an older export, or a fresh launch -- keeps today's
    behaviour, so nothing that boots now stops booting.

    Worth stating: this is only reachable once a chain has actually run `inert`
    for a while and then forks. Since this PR is what introduces the policy, the
    near-term case is the one item 15 already handles. That is why this is
    planned rather than done.

16. **Fixed: `EnablePackage` checks the policy and re-applies the gnomod rules.**
    Enable used to run under any later policy, so a package parked during an
    `inert` era stayed activatable forever — governance moves to `permissioned`
    precisely to stop strangers getting code onto the chain, and every parked
    package remained a stranger's pending deploy one approver could still land.
    `PkgApprovers` was no substitute: it is not cleared when the policy changes.

    Enable now refuses unless `code_submission_policy == "inert"`, before the
    approver check so the refusal names the real reason. Returning to `inert`
    makes parked packages activatable again; the check is about the policy in
    force, not a permanent disqualification.
    (`TestVMKeeperEnableRequiresInertPolicy`.)

    It also now calls the full `checkGnomodConstraints`, rather than the
    hand-copied private-override rule it carried before. That rule is the one
    whose answer actually changes between submit and enable — for a package
    parked before anything existed at the path, submit evaluated it against
    nothing at all. The rest are stable across the split and are re-checked
    anyway rather than hand-picked: enumerating which of a deploy's
    preconditions enable may skip is how the override rule went missing in the
    first place. `checkGnomodConstraints` takes a `priorPrivate bool` instead of
    a `*gno.PackageValue` to make this possible — enable cannot load the live
    package value without poisoning the object cache, and the bool is the only
    thing the function ever asked of it.

    `checkNamespacePermission` and `checkCLASignature` were already re-run at
    enable, above the type check; an earlier draft of this list said otherwise
    and was wrong.

17. **The creator-bound guard makes a parked path unclaimable.** §5b's guard
    has a cost worth stating plainly rather than burying. Under `inert`,
    `checkNamespacePermission` is a no-op while `r/sys/names` is undeployed —
    the state a chain boots in — so a first mover can park a one-file package at
    any name, `gno.land/r/gnoland/home` included, for one amino write plus
    preprocess gas. Nobody else can then park there: not the namespace owner,
    not governance. With `DelInertPackage` reachable only after a successful
    enable and `DisablePackage` unimplemented, the legitimate owner's only
    routes are to have an approver activate the squatter's code or to wait for
    the policy to leave `inert` — which, with item 16's policy check, now
    neutralises the squat permanently rather than merely deferring it, but also
    leaves the blob parked forever with no way to deploy at the path until the
    chain leaves `inert` for good. The guard is still right — the front-running
    it prevents is worse — but it needs a companion: let a namespace-permitted
    address replace, or let an approver reject and evict, or adopt the content
    hash of item 12, which removes the need to bind a path to a creator at all.

18a. **Parked bytes are priced well below live bytes, and only `.gno` bytes are
    priced at all.** `chargePreprocessGas` sums `len(f.Body)` only for files
    ending in `.gno`, so a submission's README, extra `.toml`, or arbitrary data
    files are charged nothing at the per-byte preprocess rate. They are still
    written -- `AddInertPackage` sets into `iavlStore`, the merkleized store --
    at amino-encode plus one `WriteCostFlat` 24,000 + 14/byte, against
    `storagePriceDefault` of 100ugnot per byte for live realm state. And no
    storage deposit is taken at submit at all: `processStorageDeposit` runs only
    at enable. Combined with items 13 and 17 (nothing can delete a parked
    package, and the path is bound to its original submitter), a submitter can
    put mostly-unpriced bytes into the app hash permanently. §3's claim that
    inert submissions are "charged up front" is true of the type-check work it
    defers and not of the bytes it stores. Fixing it means either charging the
    deposit for the parked blob at submit or counting all bytes rather than only
    `.gno` -- the latter changes the ordinary path too, so it is a consensus
    decision rather than a patch.

18. **There is no way to enumerate parked packages.** `FindPathsByPrefix` ranges
    only over `pkg:`, so `inert_pkg:` is invisible to it. A node operator cannot
    answer "what is awaiting approval" without external bookkeeping.

## Determinism audit

This branch was audited for the one failure mode that matters most here: two
honest validators processing the same block reaching different state or
different gas. The motivating instance was real — §5d's deleted source blob
diverged on restart history — so the question was whether it had siblings.

**No further instances were found, and each hypothesis fails for a structural
reason rather than by luck:**

- **Gas that depends on cache warmth.** `cacheStore.Get` genuinely charges
  nothing on a hit, and a type-check cache hit genuinely skips a metered read.
  But every cache that gates gas is per-transaction or per-block: the tm2 cache
  store is re-wrapped at each `BeginBlock`, and `MakeGnoTransactionStore` hands
  each transaction a `maps.Clone` of the type-check cache that is never
  committed back. The process-lifetime cache is written in three places, all at
  boot, all looping stdlibs — so no user package ever enters it, and every node
  starts every transaction from an identical stdlib-only set. Safe, but safe by
  a pinned invariant rather than by construction.
- **Other in-memory-only writes.** Enumerated exhaustively: `SetBlockNode` is
  the only one in the store whose backend write is commented out. Everything
  else either writes through or is per-transaction.
- **Parked packages at boot.** `AddInertPackage` never touches the package
  index, and `IterMemPackage` walks the index — so boot preprocessing cannot see
  a parked package. `inert_pkg:` also sorts outside `FindPathsByPrefix`'s range.
- **Ordering.** No map iteration in either path. File sorting is enforced twice
  over: `ValidateMemPackageAny` runs above the policy branch, and
  `runMemPackage` sorts again before storing.
- **Atomicity.** Every mutation sits inside one of two boundaries, both gated on
  `result.IsOK()` — the sdk cache-multistore and the gno transaction store's
  `txlog` wrapper.
- **Gnomod round-trip.** All fixed struct fields, no maps, `OrderPreserve`.
  Deterministic, including the new `max_deposit`.
- **Wall clock, filesystem, environment, randomness.** None on a consensus path,
  and `ProdOnly` with no test getter holds at all three type-check sites —
  including `EnablePackage`, which is the newest and was the one worth checking.

One structural finding deserves recording even though it is not a bug:
**`baseKey` is not merkleized.** It is mounted with `dbadapter`, whose `Commit`
returns a zero `CommitID`, so the VM's entire object graph — `oid:`, `tid:`,
`rlm:`, `pkgidx:` — sits outside the app hash. Only mempackage blobs (including
`inert_pkg:`) and escaped hashes are committed to the merkle store. Two
validators can therefore diverge in VM object state and still agree on the app
hash, until the divergence surfaces through a hashed channel. That is an
argument for guards like §5f rather than for trusting hash mismatch to catch
this class.

19. **`MsgEnablePackage`/`MsgDisablePackage` are absent from the codec parity
    test.** `parity_test.go` was extended for `Params` in this branch but the two
    new messages, and the ~400 lines of hand-written `pb3_gen.go` behind them,
    are not covered — and their `ValidateBasic`/`GetSignBytes`/`GetSigners` have
    no direct test either.

20. **`chargePreprocessGas` still has no charge at enable.** Submit prices the
    work by source length, but the type-check and `init()` happen in a later
    block, paid for by the approver's gas. So the block that does the work does
    not budget it, and an automated approver's hot key funds it. Whether the
    submit-time charge should be re-levied at enable, or the enable metered
    directly, is a design question this branch does not answer.

21. **Fixed: a genesis under `inert` no longer parks its own packages.** The
    inert branch had no height-0 carve-out, so a chain launching under `inert`
    stored its entire genesis set unexecuted — booting with no `r/sys/params`
    and no govdao, hence no way to propose a change and no approver able to act,
    since the realms it needs would not exist. Genesis content is the chain's
    own and there is nobody for an approver to protect it from.

    Worth noting the asymmetry that hid it: the ante already carved genesis out
    (`checkCodePolicy` returns early on `auth.GenesisReplayKey`, which
    `deliverGenesisTx` sets for every InitChain tx), so `permissioned` at genesis
    was always fine. Only the keeper's branch was missing the same treatment,
    even though the two other height-sensitive rules beside it — the type-check
    mode and the draft-package rule — both have it.

22. **gpao's remaining fee exposure is bounded, not eliminated.** It now skips
    a path that is already deployed, ignores transactions that failed on chain,
    and stops approving once a run has spent `--max-spend`. What it still cannot
    do is estimate an enable before sending it, so an approval that exceeds
    `--gas-wanted` still costs a fee to discover. Fixing that needs either gas
    estimation for the message or a query that exposes the parked key space;
    neither exists today.

23. **Fixed: gpao resolved user imports from local disk before the chain.** A
    package importing something present in the operator's `examples/` but absent
    from the chain verified clean and was approved, then failed its own
    type-check at enable — burning a fee and marking the path rejected for a
    fault that was the operator's tree, not the code. It now reads `/p/` and
    `/r/` imports from the chain only, and stdlibs from disk, because stdlibs
    ship with the binary and are not chain state. Where disk and chain agree the
    answer is unchanged; where they disagree the chain is the one that decides.

Written with AI assistance (Claude Code). Every claim about existing behavior
here was checked against the code at the commit that introduced it. Numbers
that came from a one-off measurement rather than a repeatable test are marked as
such.
