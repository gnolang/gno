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

### 8a. Rejected: charge the creator a block's gas at enable, one enable per block

Designed, reviewed, and refused. Recorded in full because the shape is the one
most people reach for, and the reason it fails is not obvious from the outside.

**The shape.**

- `MsgEnablePackage` succeeds at most once per block. A second one in the same
  block is refused.
- Before `init()` runs, the CREATOR is charged up front for a whole block's gas
  -- `Block.MaxGas` priced at the chain's current gas price -- and refunded
  whatever `init()` did not use.
- `init()` runs against a budget of `Block.MaxGas`, so the worst case a block
  can absorb is roughly double its normal execution time.
- The approver's transaction stays small: it authorises and pays its own flat
  fee, nothing more.

**Why the one-per-block rule needs no ante change.** A failed message commits
nothing -- baseapp writes the message cache only when the result is OK -- so an
unauthorised enable fails at the approver check and never marks the slot as
used. Only a *successful* enable takes it, and only an approver can produce one.
A spammer cannot starve the oracle by racing for the slot.

**Where the slot marker lives.** Not the context: `getContextForTx` rebuilds each
transaction's context from block state, and the block context is only reassigned
at InitChain and BeginBlock, so a value set inside one transaction does not reach
the next. A keeper field reset on the height changing is deterministic -- every
node replays the same block in the same order -- and keeping it out of the store
leaves no app-hash surface and nothing to migrate.

**Why the up-front charge is not the escrow this ADR already rejected.** That one
held a ceiling across time, typically thousands of times the real charge, from
submit until an approval that might never come. This is charged and refunded
inside a single message. The price is available: the chain's gas price is
consensus state, set per block by the auth keeper, so the conversion is
deterministic.

**Why it fails: the charge is reverted by the failure it exists to punish.**

The charge and refund are bank writes inside the message. `runTx` takes its
checkpoint after the ante and before the messages, and on failure
`WriteCheckpoint` flushes only the ante writes -- baseapp says so in a comment
at that line. So when `init()` exhausts the budget, the out-of-gas panic fails
the message and the charge is rolled back with everything else. The creator pays
nothing; the approver still pays its flat fee. That is precisely the drain §8
set out to close, arriving back intact.

It gets worse under an adversary. A package whose `init()` loops and then panics
is approved by the oracle -- which type-checks and preprocesses but never runs
`init()` -- costs the chain a full block's work, and costs its author only the
submit fee, because the panic reverts the charge. Ending `init()` with a panic
rather than exhausting gas is a large discount on the same attack.

There is no fix inside this message shape. For a charge to survive a failed
transaction it must be an ante-phase write, and the ante cannot know the
creator: the creator is read from the stored `gnomod.toml` inside the keeper,
long after the ante has run. Putting the creator on the transaction is the only
way, and that is a different design -- see below.

**Three more, each independently disqualifying.**

- The `Block.MaxGas` budget does not bound what it claims. The store's gas meter
  is frozen at `BeginTransaction`, before any handler runs, so a swapped meter
  reaches only the VM half and storage keeps billing the approver -- the same
  objection §8 already records. The enable-time type check is outside every
  budget, metered only by the flat per-byte charge levied at submit.
- Moving `init()` off the transaction's meter breaks block accounting either
  way. Left off the block meter, a saturated block reads as idle and the dynamic
  gas price *falls*, so attacking gets cheaper the more you do it. Charged to
  the block meter, one enable consumes the whole block allowance and evicts
  every other transaction. Today, with `init()` on the approver's ordinary
  meter, neither happens.
- The per-block marker cannot live in memory. `runMsgs` executes for simulate as
  well as delivery, `.app/simulate` is a public query that runs outside the
  consensus mutex against the last committed height, and only *store* writes are
  rolled back by the surrounding cache wrap. A field on the keeper would be both
  a data race with consensus and a fork: a simulate at height H can clear a mark
  set by a delivery at H+1, letting a second enable through on that node alone.
  A store key would be rolled back correctly, which is the pattern the auth
  keeper already uses for per-block gas price.

**Two further problems found in review.**

0. On any chain that did not configure an initial gas price the charge is zero,
   because the price accessor returns an empty value and the update
   short-circuits on it -- so the mechanism is silently inert, including on
   every txtar chain, which is why no integration test would have caught any of
   this. And the charge is unconsented in the sense §5c exists to prevent: the
   creator never declares it, it is set by a consensus param and a moving price,
   and it lands in a transaction they neither sign nor can refuse.

**Open questions, both real.**

1. Genesis replay. `deliverGenesisTx` delivers historical transactions carrying
   overridden heights. If a fork's history contains two enables at the same
   source height, a naive per-height rule refuses the second and StrictReplay
   refuses to boot. It needs the same exemption the other carve-outs have.
2. Liquidity. The creator must hold a block's gas -- on the order of several
   GNOT at present prices -- from submit until whenever an approver acts, which
   is unbounded. And exhausting the budget costs them all of it while leaving
   the package parked, which is correct as a deterrent but is the most expensive
   single mistake a submitter can make. It should be named in the error, not
   discovered from a balance.

**What to do instead.** Two things, in order.

A chain-wide `max_init_gas` vm param, applied as a passthrough meter at the
machine construction, is worth shipping on its own merits and does not depend on
any of the above. It turns the approver's required `GasWanted` from an unbounded
guess into a known constant, keeps every unit on the block meter and in the fee
market, and adds no state, no coin flow and no message change. It needs the
allocator's meter restored afterwards, which is a live bug today regardless.

For the cost question itself, put the creator on the transaction. Adding a
`Creator` signer to `MsgEnablePackage` makes them the fee payer, runs `init()`
on their own meter under their own limit, and drops the approver's cost to
nothing. It also bounds retries in a way this design could not: the ante
increments the sequence for every signer and sequence writes survive a failed
transaction, so a failed enable burns the approval and the creator must return
for a fresh signature. It closes the content-hash gap in the same stroke if the
message also carries the hash. The cost is one off-chain round trip -- the
oracle co-signs rather than broadcasts.

### 8b. Done: a flat charge on inert submission

Two vm params, `inert_submission_charge` (a literal ugnot amount, empty by
default) and `inert_charge_collector` (where it goes). `AddPackage`'s inert
branch moves the charge from the creator to the collector, last, after every
refusal. `EnablePackage` is untouched.

**Why flat rather than metered.** §8 and §8a, and two later attempts, all tried
to bill the creator for the work actually done, which means reading a gas meter
and refunding the remainder. Every one of them failed on the same property: once
money is derived from a gas reading, gas becomes a consensus input. A fork
replaying history recomputes a different refund than the source chain paid, so
balances drift; the store's per-transaction write-gas dedup can refund a charge
made on one meter into another, inflating the refund; and two node-local caches
that gate I/O gas (`stdlibKeyBytes`, the type-check cache) stop being liveness
hazards and become app-hash inputs. A flat amount reads no meter and refunds
nothing, so none of that is reachable.

**Why a literal amount rather than a price-derived one.** `LastGasPrice` is
recomputed only in `auth.EndBlocker`, and `InitChain` runs no EndBlock, so
during a fork's replay it stays pinned at `InitialGasPrice` for the whole
history. A charge computed as `Block.MaxGas × price` would therefore collect
something different on the fork than the source chain collected. A literal
amount replays correctly because params are store-backed: a governance
transaction that changed it re-applies at its own point in the replayed history.

**The economics.** The exposure being priced is not `init()` gas. Fees are flat
— the ante deducts `tx.Fee.GasFee` in full regardless of consumption — so the
approver pays the same whether the package burns 1M gas or 40M. Its exposure is
that flat fee times the number of approvals it can be induced to make, and gpao
stops approving for everyone once `--max-spend` is reached. At shipped defaults
that is 100 approvals at 1 GNOT each.

A *small* inert submission is nearly free today, and small is what the attack
wants. The chain's initial price is `1ugnot/1000gas`, and the branch does not
type-check, run `init()`, or take a storage deposit. What it does charge is
`chargePreprocessGas`, at 1250 gas per `.gno` byte, for the compile it is
deferring. That is proportional to source size, so it prices a large package
seriously — roughly 1 GNOT for 900 KB — and barely at all for the few hundred
bytes an attacker needs. A minimal package costs a few hundred ugnot, so the
deploy pipeline can be halted for well under 1 GNOT against 100 GNOT of oracle
budget.

That asymmetry is the point: the byte charge already prices bulk, and what it
cannot price is a tiny package that is nonetheless expensive to activate. The
flat charge is what covers that.

The charge inverts that ratio: each induced approval costs the attacker the
charge and the oracle its flat fee. Note the value is a governance decision and
should be set against the real floor above — an earlier draft of this design
calibrated it against a figure taken from a txtar test convention, which is 100×
the chain's actual price.

**What it does not do.** It prices the drain rather than eliminating it, and it
does not bound the approver's per-enable gas: a parked blob is bounded by
`MaxTxBytes` at 1 MB, and reading it plus the namespace and CLA realm calls puts
the approver's worst case near 40M gas. gpao now sizes each approval by
simulating it rather than by a fixed flag (item 22), so that number no longer has
to be guessed at — but it is still gas the approver pays. The approve/activate
split above remains the better long-term answer, because it deletes the drain
instead of pricing it.

**Ceiling.** `Validate` caps the charge, because a charge governance can raise
without limit is a deploy freeze — the outcome the charge exists to prevent. It
is otherwise an ordinary `vm:p:` param and `r/sys/params`' generic factories can
set it; if the collector is an approver DAO, that DAO can propose its own
revenue, which is a reason to consider a gated setter later.

**Empty means off, and the collector is not validated.** Empty-by-default is
what makes replay correct without a carve-out — a chain whose history predates
the field replays charge-free rather than being charged for submissions that
never paid. So `applyLegacyDefaults` fills the collector and must never fill the
charge, and neither must the legacy fill in `gnogenesis fork generate`. The
collector is deliberately unvalidated: an unconditional non-zero rule breaks
`fork generate`, which builds `Params` without `applyLegacyDefaults`, and a
cross-field rule would abort a governance proposal mid-execution, since
`WillSetParam` re-validates the whole struct and panics while `r/sys/params` sets
one key per proposal. The guard sits at the charge instead, where a
misconfiguration skips it rather than burning it at the zero address.

### 9. `vm/qpkgmeta_json`: a parked package is visible to its creator

Under "inert" a submission is stored without being activated, and every other
query reads the live key space only. `GetInertPackage` had two callers, both
inside the keeper's write path. So between paying to submit and somebody
approving, a creator could see nothing: `vm/qpkg_json` and `vm/qfile` both
answer "package not found", exactly as they do for a path nobody ever used.
With a submission charge, that is a real cost followed by silence.

The query reports `live`, `inert` or `absent`, plus the creator, submit height
and declared deposit that `stampGnomod` wrote. Both key spaces store that file,
so one read path serves both states and nothing new is persisted.

**Absent is a successful response carrying a status, not an error.** A caller
has to tell "never submitted" from "the node could not answer", and an error
collapses the two — which is the failure being fixed, not a detail of it.

**`Pending` is separate from the status.** `AddPackage` refuses to park over a
live package, but only a public one — its liveness check is
`pv != nil && !pv.Private`, and that exemption is what makes a private redeploy
possible. Both key spaces then hold the path at once. Reporting only the live
package there would hide the parked submission, so the status keeps describing
what is callable and `Pending` reports that something awaits approval.

**Named apart from `qpkg_json` deliberately.** That one dumps a live package's
variables and cannot answer for a package that is not live yet. Sharing a prefix
would imply they are two encodings of one question.

The query is metered like the others: `GetInertPackage` charges amino-decode gas
by blob length, under `maxGasQuery`. `ParseMemPackage` reads only `gnomod.toml`,
so the parse does not scale with package size.

Not done: gnoweb does not call it, so nothing renders a status yet.

### 10. Listing the queue, and the oracle's own verdicts

§9 answers "is *my* package parked". Two other questions had no answer at all.

**What is awaiting approval.** Parked packages sort outside
`FindPathsByPrefix`'s range, so they appeared in no listing. An operator asking
what was waiting had to keep its own books, and an oracle restarting could not
catch up on what it missed while down — it learns about packages by watching
blocks go past, so anything submitted during an outage was simply never seen.
`vm/qinertpaths` is a plain prefix match with the same `?limit=` cap as
`vm/qpaths`, which the two now share rather than copy. Deliberately without
`qpaths`' `@user` handling: this answers an operational question about a queue,
not a browsing one.

**Why *this* package was refused.** That is the oracle's knowledge and not the
chain's. The chain can report that a package is parked and that no enable could
succeed right now — no approvers configured, policy moved off `inert`. It
cannot report that the code failed to type-check, or that the enable was
simulated and the chain would reject it. Until now the only record was the
operator's stderr, so a submitter paid the charge and heard nothing.

gpao records a verdict at each point it reaches one and serves them read-only
under `--status-listen`, off by default. `blocked` is kept distinct from the
failure states: it means the oracle has hit `--max-spend` and nothing is wrong
with the package, which is the difference between "fix your code" and "go ask
the operator" — indistinguishable from silence otherwise.

The board is a separate structure rather than an export of `seen`,
`overBudget` and `failedEnable`. Those are owned by the verifier goroutine and
carry no lock, which is what keeps the hot path cheap; serving them would read
them from the HTTP goroutine. This is the only oracle state that crosses that
boundary, so it is the only state that takes a mutex. It is keyed by path,
where the retry counters are keyed by content hash — a retry is about specific
bytes, but somebody asking after a package wants the latest word on the path
they submitted.

Related: gpao no longer broadcasts an enable the node has already rejected.
`gnoclient.Simulate` reported a transport failure, a rejected query and a
failed message as one error; `SimulateResult` returns the response as given so
the three can be told apart. The right answer to an unreachable node is the
opposite of the right answer to a rejected message — a node that cannot be
reached must not stop approvals, or anyone able to disturb the query path
could stall the chain's deploys.

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
either list omits a historical signer, and would come up missing every package
those signers deployed. `StrictReplay` does not prevent that — see item 15 —
so the failure would be reported and then ignored. Keyed on
`auth.GenesisReplayKey`, not `BlockHeight() == 0`, because forked txs carry
their original heights.

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
skip zero values, and `getStructFieldsFromStore` reads them back the same way,
one `Get` per field.

Gas goldens do **not** move, and the mechanism is worth recording because the
intuition says otherwise. The whole-struct `GetParams` happens in gno.land's own
ante closure in `app.go`, not in tm2's auth ante — it runs before
`authAnteHandler`, so the ctx still carries the throwaway `NewInfiniteGasMeter`
that `sdk.NewContext` installed and `auth.SetGasMeter` has not yet replaced.
baseapp keeps one cache wrap across the ante and the message handlers, so every
later `GetParams` is a cache hit, and `cacheStore.Get` charges nothing for those.

Look in the right ante when checking this. `AccountKeeper.GetParams` passes
`ctx.WithGasMeter(nil)`, which is a different mechanism for a different struct;
`VMKeeper.GetParams` does not, and tm2's auth ante never reads vm params at all.

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

   Gas estimation (item 22) changes the picture without closing this. Simulating
   an enable runs the message for real on the node, `init()` included — the
   simulate path executes handlers, it only discards the result. So the oracle
   now has an answer about `init()` in hand before it broadcasts; it just does
   not act on it. A package whose `init()` panics fails the simulate, gpao logs
   that the estimate failed, and broadcasts anyway — paying a full fee to be
   told what it already knew.

   Not changed here because the two failure kinds need telling apart first. "The
   node says this message fails" is grounds to decline; "the node would not
   answer" is not, and treating them alike hands anyone who can disturb the query
   path a way to stall approvals chain-wide. `Simulate` reports both as an error
   and distinguishes them only by message text. **Needs a decision:** either
   surface the ABCI error typed so the two can be separated, or leave the fee as
   the cost of not knowing.

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

5. **Partly fixed: the replay carve-out is per-check, not per-delivery.** "This
   delivery is replayed history" is a property of the delivery, yet each check
   hooked `GenesisReplayKey` itself.

   The extraction this item asked for has landed: `auth.IsGenesisReplay(ctx)`
   now sits beside the key and all four branch sites call it, so the comma-ok
   and the key type are written once.

   What is still open is the asymmetry it exposed. `checkSessionRestrictions`
   has no carve-out at all, so a replayed session tx *is* re-authorized against
   current params — a fork that tightens session rules would refuse its own
   history. And the auth ante's own use requires a second condition (the
   operator's `--skip-genesis-sig-verification`) while the code-policy and vm
   carve-outs fire on the predicate alone, so the same phrase means two
   different things depending on where it is read.

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

15. **Fixed: genesis replay follows the policy that governed the replayed
    history.** Replay executes a previous chain's transactions through this
    keeper at `BlockHeight > 0`, after `InitGenesis` has installed the fork's
    params. The question is which policy those transactions run under.

    The answer the params already give is the right one. `code_submission_policy`
    is store-backed, `gnogenesis fork generate` copies the source chain's vm
    params into the fork's genesis untouched — it rewrites only the depth params
    and `preprocess_gas_per_byte` — and every historical governance transaction
    that moved the policy re-applies as it replays. So `GetParams` during replay
    already returns what the source chain read at that point in its history,
    including a history whose policy changed partway.

    **That rests on one property of the fork tool worth naming, because nothing
    in the keeper can see it.** `buildHardforkGenesis` starts from the source
    chain's **genesis document**, not from a state export, and appends history
    after it. So the fork begins at the source's height-0 params and every
    governance transaction that moved them replays in order.

    Snapshot the source's FINAL state instead and this inverts: every historical
    transaction would run under the last policy the chain ever had, so a chain
    that adopted `inert` late would park its entire history. The keeper would be
    reading the params correctly and still get every answer wrong. Both halves
    are covered — the source-genesis start by the params assertions in
    `TestBuildHardforkGenesis_DefaultsGasParams`, the ordering by the
    `SourceBase`-then-`SourceHistorical` assertions in
    `TestBuildHardforkGenesis_AnnotatesSource`.

    `AddPackage`'s inert branch therefore reads the policy like any other
    delivery. A chain that parked a package parks it again; one that deployed
    live, deploys live.
    (`TestVMKeeperGenesisReplayFollowsTheReplayedPolicy`.)

    `EnablePackage` exempts its two **authorization** gates during replay, and
    runs everything else. The exemption is for the reason the ante exempts its
    own: a fork may have moved off `inert` or rotated `pkg_approvers`, and must
    not refuse its own history — the mandate those gates protect was exercised
    on the source chain, and replay reproduces that record rather than granting
    it again. Everything after them runs, so a package the source chain parked is
    activated here as it was there.
    (`TestVMKeeperGenesisReplayEnableActivatesAParkedPackage`.)

    One exemption remains beyond the gates: a replayed enable that finds nothing
    parked **and finds the package live** returns nil. That is the case where the
    replayed policy was not `inert`, so the submission went live on the ordinary
    path and the enable is genuinely spare work — which is what every genesis
    exported before this branch existed looks like.
    (`TestVMKeeperGenesisReplayEnableWithNothingParkedIsANoOp`.)

    The liveness condition is load-bearing, and an earlier version of this branch
    omitted it. "Nothing parked" has a second cause: the replayed `MsgAddPackage`
    **failed**, which it can do for reasons its own history did not have — a
    creator whose account an earlier diverging transaction never created, a
    namespace that changed hands, a prior park by someone else. Returning nil
    there records success for a package that is not on the chain, and the fork
    boots with the realm missing while the replay report's last word on that path
    is "enabled OK". Blob presence is the exact liveness test, probed with
    `GetMemPackage` rather than `GetPackage` for the reason documented at the
    `liveBlob` probe.
    (`TestVMKeeperGenesisReplayEnableRefusesWhenNothingIsLive`.)

    **`StrictReplay` is advisory, not a boot guard.** Several comments on this
    branch justified themselves with "a failed genesis transaction stops the node
    booting". It does not. `InitChain` puts the failure count in
    `ResponseInitChain.Error`, `localClient.InitChainSync` returns a nil Go error
    regardless, and the handshake inspects only that error — so the response
    field is never read. The repo already knew: the valoper coverage check
    `panic`s precisely because that is "the only way to abort handshake". Nothing
    on this branch may rest on StrictReplay stopping a node until that is fixed,
    which is its own change.

    **Rejected: carry the source policy per transaction in `GnoTxMetadata`.**
    This was the planned fix, and it is unnecessary and unimplementable as
    designed. Unnecessary because the params already carry the answer, as above.
    Unimplementable because nothing can populate the field: every metadata field
    `gnogenesis fork generate` sets is an observation of block data, and its only
    chain-state query is one account lookup at the halt height. A policy at
    height H is keeper state, which would need an archive node retaining every
    version and a query per transaction. `tx-archive backup` is further still —
    it populates only `Timestamp`. The field would also have been the first
    `GnoTxMetadata` entry to change a message's state transition rather than its
    delivery context, and it would have needed a `pb3_gen.go` regeneration to
    avoid being silently dropped on the binary codec path.

    **Known limit.** An operator who sets a *different* policy in the fork's own
    genesis params gets that policy applied to the replayed history, because
    `InitGenesis` runs before the replay loop. To adopt `inert` at a fork, turn
    it on with a migration tx appended after the history rather than in the
    genesis params. The alternative — staging vm params across the replay
    boundary so replay always runs on source params — moves every vm param, not
    just this one, and needs its own analysis.

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

18d. **A capability the realm-param key rule removes, stated here because the
    commit that added it did not.** A genesis section named `[vm:p]` used to
    work as a way to set any vm parameter. The loader read `p` as the realm
    part, so the keys became `p:<name>`, `InitGenesis` wrote them as
    `vm:p:<name>` -- the same store keys the real fields use -- and
    `WillSetParam` validated and applied them. That route went around the `[vm]`
    section, which accepts only `chain_domain` and `sysnames_pkgpath`.

    It was an accident of the loader rather than a designed path, and it is what
    let a genesis file quietly overwrite a validated parameter, so refusing it is
    the point. But `genesis_params.toml` carries three commented-out vm params
    (`max_gas`, `chain_tz`, `default_storage_allowance`) waiting on the
    unmarshaler work its own TODO describes, and `[vm:p]` was the workaround an
    operator could have reached for.

    The supported route is `gnogenesis params set vm.<field>`, which sets the
    typed struct field by its json tag and is validated before the file is
    written. Widening the `[vm]` section to the full struct is the real fix and
    is what that TODO is asking for.

18c. **Correction to another commit message.** The commit adding the
    second-colon rule justifies it by saying a realm writing a param at runtime
    goes through `sys/params`' `prmkey`. That names the wrong path. `sys/params`
    is governance-only, locked to `gno.land/r/sys/params`; an ordinary realm
    writes through `chain/params`' `pkey`, which panics on a colon in the key
    and builds `vm:<realm-path>:<key>` from a package path that cannot contain
    one.

    The rule is still right, and for a better reason than the one given: a
    runtime key has exactly one colon because `pkey` enforces it, so genesis is
    the only writer that can produce more. Worth noting `prmkey` checks the name
    but not the submodule, so `gno.land/r/sys/params` could in principle build a
    multi-colon key -- no code there does, and that realm cannot be redeployed.

18b. **Correction to an earlier commit message.** The commit that added the
    realm-param key rule says the `[vm:p]` genesis section "bypasses
    validation". That is wrong and overstates the hole. `SetAny` reaches
    `ParamsKeeper.validate`, which resolves the module prefix and calls
    `VMKeeper.WillSetParam`; that has a `p:run_submitters` case, parses the
    addresses, and finishes with `params.Validate()`. An unparseable or invalid
    value is refused.

    What the section actually bypasses is the loader's `[vm]` allowlist, which
    admits only `chain_domain` and `sysnames_pkgpath`. So the hole is "an
    operator can set vm params the allowlist was written to refuse", not
    "arbitrary unvalidated values reach the store". Recorded here because the
    commit has descendants and cannot be amended.

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

18. **Fixed: parked packages can be enumerated.** `FindPathsByPrefix` ranges
    only over `pkg:`, so `inert_pkg:` is invisible to it and a parked package
    appeared in no listing at all. `FindInertPathsByPrefix` ranges that space
    and `vm/qinertpaths` serves it; see §10.

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
  It is load-bearing rather than tidy: `AddInertPackage` stores the mempackage
  verbatim, test files included, so without it enable would type-check `_test.gno`
  and need a getter the consensus path does not have — making the answer depend
  on the node's disk. Pinned by
  `TestVMKeeperEnableTypeChecksProductionFilesOnly`, which exists because
  mutating the flag broke nothing else in the suite.

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

22. **gpao's remaining fee exposure is bounded, not eliminated.** It skips a path
    that is already deployed, ignores transactions that failed on chain, and
    stops approving once a run has spent `--max-spend`.

    It now also sizes each approval by simulating it: the measured gas plus 20%,
    bounded by the chain's `Block.MaxGas`. That replaces a fixed `--gas-wanted`
    flag that was shipped at 20,000,000 against a worst case near 40M, so
    approvals no longer fail for being under-funded by default. The flag remains
    as the fallback for a node that will not simulate.

    Two details the implementation turns on. The probe is signed at the chain's
    block ceiling rather than at the fallback, because simulate executes under
    the transaction's own limit — sizing the probe at the fallback would run out
    of gas on exactly the packages worth measuring. And the ceiling is read from
    the chain rather than assumed, because the ante refuses a `GasWanted` above
    `Block.MaxGas` instead of clamping it, so a chain configured below the tm2
    default would reject every probe.

    What remains: fees are flat, so a failed enable still costs a full fee to
    discover. `--max-spend` bounds that across a run, and a per-package attempt
    counter (item 24) bounds it per package.

23. **Fixed: gpao resolved user imports from local disk before the chain.** A
    package importing something present in the operator's `examples/` but absent
    from the chain verified clean and was approved, then failed its own
    type-check at enable — burning a fee and marking the path rejected for a
    fault that was the operator's tree, not the code. It now reads `/p/` and
    `/r/` imports from the chain only, and stdlibs from disk, because stdlibs
    ship with the binary and are not chain state. Where disk and chain agree the
    answer is unchanged; where they disagree the chain is the one that decides.

24. **Fixed: gpao retired a package on its first failed enable.** A candidate was
    marked seen before the approval was broadcast, so one failure removed those
    bytes from consideration for the rest of the run, with the reason visible
    only in the oracle's stderr.

    That matters more once submitting costs money: a creator pays the charge, the
    oracle approves, the enable fails for something that is not about their code
    — an unfunded storage deposit, a dependency not yet live, a namespace or
    governance param that moved, a block out of gas — and they see a charge
    followed by silence.

    Seen is now set only on terminal outcomes: verification rejected the bytes,
    the package was already active, or the enable succeeded. A failed enable goes
    through a bounded attempt counter instead, the same shape the verify-budget
    overrun already used. The spend-bound refusal is also left unrecorded, since
    it tells the operator to raise the bound or restart and both are no-ops on a
    path that has been retired.

    As with the overrun counter this does not re-offer the package by itself —
    heights only move forward. It is what makes a restart, or a resubmission of
    the same bytes, do something rather than nothing.

Written with AI assistance (Claude Code). Every claim about existing behavior
here was checked against the code at the commit that introduced it. Numbers
that came from a one-off measurement rather than a repeatable test are marked as
such.
