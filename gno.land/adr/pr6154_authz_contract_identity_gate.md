# ADR: bind `ContractAuthority` to a real principal, and stop consumers handing out live capabilities

## Context

`p/moul/authz`'s `ContractAuthority` is documented as meaning "this contract
is the authority". In practice it compared the `caller` argument against
nothing:

- `NewContractAuthority` defaulted `proposer` to `AutoAcceptAuthority`, whose
  `Authorize` runs the action for any caller.
- The consumer's `PrivilegedActionHandler` is a caller-supplied closure. The
  idiomatic registration — `func(_ string, action PrivilegedAction) error {
  return action() }` — also runs unconditionally.

So `caller` travelled the whole chain (`Authorizer.Transfer` /
`DoByCurrent` / `DoByPrevious` → `ContractAuthority.Authorize` → proposer →
handler) and was never compared to anything. `MemberAuthority` is the only
`Authority` implementation in the package that consults `caller` at all.

The previous gate, `unsafe.CurrentRealm() == contractAddr`, lived inside the
consumer's handler and was removed in the interrealm v2 migration for being
`.Title()`-bypassable (`runtime.CurrentRealm` walks past non-crossing frames
to the most recent crossing ancestor). Nothing replaced it, so plain
`NewContractAuthority` consumers were left with no caller check.

`r/gnops/valopers` is the only shipped consumer of plain
`NewContractAuthority`. It compounded the gap by exporting `Auth() *authz.Authorizer`,
handing every caller a live, mutable handle to the authority — including
`Transfer`, which replaces the authority outright. Because `Transfer` is
gated by the authority it replaces, installing a `DroppedAuthority` is
irreversible: no authority remains that could restore governance, and the
legitimate GovDAO `updateInstructions` path routes through the same object.

The same realm leaked a **second** live capability of the same class, on a
different axis: `Valoper.Auth() *authorizable.Authorizable` (per-operator
auth lists). Found in review, pre-existing, and fixed here — see
"What review found: a sibling capability leak".

### What review found: the identity gate alone is not sufficient

An identity gate binds *who the frame belongs to*, not *what the frame is
doing*. The VM mints a crossing frame's `cur` from the **callee's declaring
package**, so inside any crossing frame of realm X, `rlm.Address()` is
unconditionally `chain.PackageAddress(X)`. Two consequences:

- `NewContractAuthority(ownPath)` + `DoByCurrent` is a **tautology**. Only
  `DoByPrevious` and `Transfer` gain a real check.
- **Any exported function of the authority contract that returns a crossing
  closure hands out that contract's identity**, and with it the authority.

`r/gnops/valopers` did exactly that, via
`NewInstructionsProposalCallback(string) func(realm) error`. Any realm could
call the returned closure and rewrite `instructions` with no proposal and no
vote. Reproduced from an unrelated `PKGPATH` both with layers 1–2 applied and
on plain `master` — a pre-existing sibling hole, not a regression, but one
that leaves the same impact class open. A tree-wide search for exported
functions returning `func(realm) error` finds exactly one such leak, this one;
every other proposal-callback consumer builds its closure inline and returns a
`dao.ProposalRequest` whose executor never escapes.

That search was scoped to *closure*-returning functions, which is why it
missed the sibling leak below: an exported getter returning a mutable
capability **struct** is the same hazard and matches no `func(realm) error`
pattern. Audit by returned type, not only by returned closure.

This was also **not patchable after the fact** for most of this PR's life:
nothing in valopers called `auth.Transfer` and layer 2 removes the external
handle, so the authority was unrotatable once deployed. Decision 7 closes
that — a GovDAO-gated rotation entrypoint now ships with the fix, which it
had to, since code cannot be added to a deployed realm.

### What review found: the tautology is the wrong principal, not a dead check

The identity gate is sound in the package, but `r/gnops/valopers` pointed it
at **itself** and drove it with `DoByCurrent`, which is precisely the
tautology above: the check compared valopers' address to valopers' address
and could never reject. Measured by instrumenting the live governance path
(propose → vote → execute):

```
rlm.PkgPath()            = gno.land/r/gnops/valopers
rlm.Address()==valopers  = true    <- what DoByCurrent compared
rlm.Previous().PkgPath() = gno.land/r/gov/dao
prev.Address()==govdao   = true    <- available, and unused
```

The principal the realm actually wants to assert is in `Previous()`. Decision
4 repoints the authority at `gno.land/r/gov/dao` and switches to
`DoByPrevious`, which turns the check into a real one and makes it the only
**code-level** guard against the regression this realm most fears — a
maintainer later adding an exported crossing entrypoint that reaches
`updateInstructions`. Measured, same attack under both configurations:

| | own path + `DoByCurrent` | `r/gov/dao` + `DoByPrevious` |
|---|---|---|
| foreign realm calls a re-exported plain entrypoint | write lands | `panic: unauthorized` |
| foreign realm wraps a re-exported *closure* in its own executor | write lands | write lands |
| legitimate propose → vote → execute | ok | ok |

The second row is an honest limit, verified rather than assumed: an
attacker holding a crossing closure chooses its `Previous()` by wrapping it
in their own `dao.NewSimpleExecutor`, so the principal check cannot see that
shape. It stays guarded by the seal in decision 3.

### What review found: a sibling capability leak

`valopers.gno` exported `func (v Valoper) Auth() *authorizable.Authorizable`,
and `GetByAddr` is exported and non-crossing. `Valoper` is returned by value
but `auth` is a pointer field, so the copy shared the realm's live
`Authorizable`. `Authorizable`'s gates read `rlm.Previous().Address()`, so
inside a hostile realm's frame `Previous()` is whoever called it. A valoper
**operator** merely calling any function of a hostile realm (faucet, airdrop,
mint) let that realm run

```gno
valopers.GetByAddr(operator).Auth().AddToAuthList(0, cur, attacker)
```

with `Previous()` == the operator, i.e. the `Authorizable`'s own owner. The
write was accepted and persisted; from the next transaction the attacker
acted alone — `UpdateKeepRunning` to drain the validator,
`UpdateSigningKey` to rotate its consensus signing key.

Pre-existing (byte-identical at the merge-base) and **critical**. It is
bundled here rather than deferred because it is the same class as layer 2 —
an exported live capability handle — in the same realm, and the audit that
produced this ADR should have caught it. That audit searched for exported
functions returning `func(realm) error`, so it structurally could not see an
exported getter returning a mutable capability struct. The lesson is
recorded in decision 2.

Note the exported wrappers `AddToAuthList`/`DeleteFromAuthList` were never
the hole: there `cur.Previous()` is the hostile realm rather than the
operator, so the owner check rejects. The raw handle was a strict privilege
escalation over the intended surface.

### What review round 4 found: the asserted principal was itself forgeable

Layer 4 points the valopers authority at `gno.land/r/gov/dao` and compares
`rlm.Previous().Address()` against it. That is only a gate if the caller cannot
choose to be `r/gov/dao` — and it could.

`dao.NewSimpleExecutor` and `SimpleExecutor.Execute` are both exported, and
`Execute` is a *crossing* method declared in `r/gov/dao`, so invoking it mints a
`gno.land/r/gov/dao` frame for whatever callback it wraps. Any realm could
therefore satisfy any gate keyed on that path or its address, with no proposal,
no vote and no DAO membership — and with no proposal id minted, so nothing to
render, audit or deny afterwards. Characterized during review, which left the
remediation open; this ADR takes it.

Two consumer shapes were reachable, both measured on a victim realm written
exactly as this package's own godoc instructs:

| victim exports | before | after |
|---|---|---|
| an ordinary privileged entrypoint `func(cur realm) error` | attacker wraps it in their own executor; the write lands | `execution denied` |
| its `*authz.Authorizer` (as two `examples/quarantined` consumers do) | attacker rotates the authority to `AutoAccept`, then writes freely | authority unchanged, write `unauthorized` |

The first is the sharper one: it needs nothing unusual from the victim. A Gno
function value `func(cur realm) error` is assignable to `dao`'s callback type, so
the governance realm will invoke the victim's own entrypoint on an attacker's
behalf. `r/gnops/valopers` was never vulnerable — it exports no such symbol, its
`auth` is unexported, and it wraps no `Transfer` — but the *guidance* this PR adds
recommended the pattern package-wide, which is the durable artifact.

The documented containment in `r/gov/dao/types.gno` was never the one holding the
line. It said invocation was "contained by the executor being unexported inside
`ProposalRequest`/`Proposal` with no accessor" — that protects the executor
*object*, but an attacker never needs someone else's executor. A reachable
*callback* suffices, and `NewSimpleExecutor` is public.

### What review round 4 found: three further defects in the package

- **An explicit proposer *replaces* the identity gate, it does not widen it.**
  `NewRestrictedContractAuthority(govdaoPath, h, NewMemberAuthority(alice))`
  means "alice, and not GovDAO": `contractAddr` is never consulted on that
  branch. The godoc said "widen who may propose" and `String()` renders
  `contract=…,proposer=…`, which reads as a conjunction, so a consumer adding a
  proposer could reasonably believe it had added a second lock rather than
  removed the first.
- **`String()` was spoofable.** It called `a.proposer.String()` directly on an
  open interface, so a foreign impl returning `"contract-identity"` made a fully
  permissive authority render byte-identical to the gated default. Measured at
  the consumer: with such a proposer installed in `init.gno`, four of the five
  valopers configuration artifacts stayed green and only
  `TestUpdateInstructionsRejectsNonGovDAO` went red. Since layer 2's entire
  product is a description string for on-chain inspection, that made the
  inspection surface untrustworthy.
- **A malformed `path` was accepted and permanently bricked the authority.**
  `chain.PackageAddress` hashes any string, so `"gno.land/r/gov/dao "` binds an
  address no realm can present; `Transfer` routes through the same gate, so it
  cannot be rotated out either — the same "permanent brick with no on-chain
  recovery" the nil-handler panic exists to prevent. Validation catches typos but
  *cannot* catch a well-formed wrong path: `"gno.land/r/gov/dao/impl/v0"` (the
  path `allowedDAOs` holds, hence the natural thing to reach for) is valid and
  permanently dead, because an executor's `Previous()` is the proxy, never the
  impl. That is what makes a rotation entrypoint load-bearing rather than
  optional.

### What review round 4 found: a refusal could strand a proposal

`updateInstructions` panicked on the authz error. `impl.ExecuteProposal` turns an
executor *error* into status `Denied` plus a `DeniedReason`, but it cannot see a
panic: a panic aborts the whole transaction, rolling back the `Denied` write, so
the proposal stays `Accepted` and retryable forever — precisely the state
`ExecuteOrRejectProposal`'s godoc exists to avoid. Unreachable on the approved
route today, but `init.gno` offers its handler as the seam for an intent check
(timelock, quorum, audit trail), and the first such check to return an error
would have hit it.

## Decision

Four layers: in the package, in the consumer's capability getters, in the
consumer's capability surface, and in which principal the consumer asserts.

1. **`NewContractAuthority` gates on the contract's own identity** instead of
   defaulting `proposer` to `NewAutoAcceptAuthority()`. Only actions whose
   caller is the contract's own package address proceed.

   (Round 4 changed *how*: the default started as `proposer == nil` read by
   `Authorize` as "the contract itself", and is now an explicit
   `contractIdentityAuthority` bound to `contractAddr`. See decision 6 for
   why the absence-as-policy encoding was replaced.)

   A `NewMemberAuthority(contractAddr)` default would behave identically, and
   was the first shape tried, but it is wrong for a fixed identity: it
   allocates an `addrset`/`avl.Tree` in realm storage for one address that
   never changes, leaves `contractAddr` with no reader, and puts
   `AddMember`/`RemoveMember` — mutators meaningless for a contract identity —
   on the security-critical default path.

   The `caller` reaching `Authorize` is established upstream by
   `Authorizer.DoByCurrent` (`rlm.Address()`), `DoByPrevious` and `Transfer`
   (`rlm.Previous().Address()`), all under `rlm.IsCurrent()`. An external
   realm cannot present the contract's address through those entry points.

   `NewContractAuthority` now also rejects an empty path and a `nil` handler.
   The empty path would bind the gate to `chain.PackageAddress("")`, a
   meaningless identity. The `nil` handler is worse: `Authorize` checks
   `contractHandler == nil` *before* consulting the proposer, and `Transfer`
   routes through `Authorize`, so a nil-handler authority can never be
   rotated out — a permanent brick with no on-chain recovery, the same
   failure mode reached by an ordinary deployer typo. Both match
   `NewRestrictedContractAuthority`, which already panicked on each.

2. **`r/gnops/valopers` exports descriptions, never live capabilities.** Two
   getters, one rule.

   `Auth()` returns `string`, not `*authz.Authorizer`. Rendering and
   inspection only need a description. Not exporting the handle removes the
   mutator surface rather than relying solely on layer 1 to reject the
   caller. Zero in-tree callers used the pointer.

   `Valoper.Auth() *authorizable.Authorizable` is **removed** and replaced by
   `Valoper.AuthOwner() address`, which copies out a value. This closes the
   critical sibling leak described in Context. Zero in-tree callers outside
   the package used it; in-package call sites use the `auth` field directly.

   The generalisation, since this is the second instance of the same bug in
   one realm: **an exported getter returning a pointer to an
   authority/capability struct is equivalent to exporting its mutators**,
   even when the containing value is returned by value. Audit for the
   returned *type*, not only for returned closures.

   `ContractAuthority.String()` now also renders the **proposer**, not just
   the contract path. This is security-relevant rather than cosmetic: it was
   the only thing that could distinguish

   ```gno
   NewContractAuthority(path, handler)                           // gated
   NewRestrictedContractAuthority(path, handler, AutoAccept{})   // open to all
   ```

   and without it the two rendered byte-identically. Review demonstrated the
   consequence: swapping valopers' `init.gno` for the open form — i.e.
   reopening the hole for that realm — left every consumer assertion and
   the whole valopers suite green. It also means an on-chain reader of
   `Auth()` can now tell a gated authority from an open one.

3. **`r/gnops/valopers` no longer exports a privileged closure.**
   `NewInstructionsProposalCallback` is removed; the closure and the proposal
   request are both built inside valopers, which returns a
   `dao.ProposalRequest`. `proposal.ProposeNewInstructionsProposalRequest`
   keeps its signature and delegates.

   Returning a request rather than an executor is deliberate.
   `dao.ProposalRequest` exposes only `Title`/`Description`/`Filter` — the
   executor is an unexported field with no accessor, so the capability is
   sealed. A `dao.Executor` would not be: `Execute` is an exported method, so
   a returned executor stays directly invocable by its holder, and
   `dao.NewSafeExecutor`'s only gate (`InAllowedDAOs`) **fails open** while
   `allowedDAOs` is empty — the documented bootstrap state, see
   `r/gov/dao/loader/v0`.

   The symbol is removed rather than deprecated: an inert shim is pointless
   and a working one reopens the hole. Replay-checked against live
   `gnoland1` — the realm's instructions are byte-identical to the genesis
   default and none of its 26 GovDAO proposals is an "Update instructions"
   one, so the symbol has no on-chain trace. Precedent: the min-fee callback
   was already removed the same way despite having been used (Prop #19).

4. **`r/gnops/valopers` asserts GovDAO as the principal, not itself.**
   `init.gno` builds `NewContractAuthority("gno.land/r/gov/dao", handler)` and
   `updateInstructions` uses `DoByPrevious`, so the gate compares the realm
   that crossed in against GovDAO's package address.

   This replaces `NewContractAuthority(ownPath)` + `DoByCurrent`, which was
   the tautology measured in Context. The repoint is what makes the check
   able to reject at all, and it is what closes the plain-entrypoint
   regression shape in the table above. It also makes `Auth()`'s description
   truthful — `contract_authority[contract=gno.land/r/gov/dao,proposer=contract-identity]`
   states the governance model rather than restating the tautology.

   The pass-through handler in `init.gno` is retained as the seam for a
   future *intent* check (timelock, quorum, audit trail). It is explicitly
   not where the caller check lives.

Legitimate governance is unaffected. `SimpleExecutor.Execute`
(`examples/gno.land/r/gov/dao/types.gno`, package `dao` — **not**
`r/gov/dao/impl/v0`) invokes the proposal callback via `e.callback(cross(cur))`,
crossing *into* the valopers-defined callback. Inside it `cur` is valopers
and `cur.Previous()` is `r/gov/dao`, so `DoByPrevious` authorizes as
`contractAddr`. Verified end to end by
`r/gnops/valopers/proposal/filetests/z_governed_instructions_filetest.gno`
(propose → vote → execute → `instructions` updated) and by
`gno.land/pkg/integration` `TestTestdata/valopers`, both under the repoint.

## Decision (round 4 additions)

5. **`SimpleExecutor.Execute` is invocable only from the `r/gov/dao` namespace**
   (`r/gov/dao/types.gno`). This is option 1 of the two considered, chosen as the
   smallest change matching the documented intent. The principal is the proxy
   itself — deliberately **not** `InAllowedDAOs`, which is `SafeExecutor`'s bug:
   that list holds the *impl* path while an executor's caller is the proxy, so it
   rejects the approved route and fails open while empty. Accepted set is the
   proxy path exactly plus its subpackages: on the approved route
   `impl.ExecuteProposal` is non-crossing, so `Previous()` is
   `gno.land/r/gov/dao` exactly, while a bare prefix check would reject every
   real proposal.

   This gates the executor object's *invocation*. Keeping a privileged closure —
   or an exported entrypoint shaped like `func(realm) error` — out of reach
   remains the consumer's job, and is now stated as such in both godocs.

6. **The contract-identity default is a real `Authority`, not a `nil` field.**
   `contractIdentityAuthority` is a two-line unexported impl bound to
   `contractAddr`. Encoding the strictest policy as an *absence* meant the secure
   default was what you got by leaving something out — so a struct literal that
   skipped the constructor, a value persisted by an older build, or re-adding the
   one-line `if proposer == nil { proposer = NewAutoAcceptAuthority() }` would all
   silently be wide open again. `Authorize` now fails closed on `nil`, both
   constructors funnel through one private constructor, and `String()` renders any
   non-canonical proposer wrapped as `custom_authority[…]` (mirroring what
   `Authorizer.String` already did at the outer level), so a foreign impl cannot
   name itself `contract-identity`.

7. **The path is validated, and the authority is rotatable.**
   `assertValidContractPath` mirrors `gnolang.ReGnoUserPkgPath`'s "lowercase ascii
   alphanumeric" rule as a deliberately permissive superset, since no path
   validator is exported to Gno code. Because validation cannot catch a
   well-formed wrong path, `r/gnops/valopers` gains
   `NewAuthorityRotationProposalRequest` — sealed in a `dao.ProposalRequest` the
   same way the instructions write is, reaching an unexported `rotateAuthority`.

   It takes a **path**, not an `authz.Authority`: an interface parameter would be
   an open-interface input, letting a proposal install an always-approve or
   always-deny authority that readers of `Auth()` could not distinguish. A path
   keeps the installed authority canonical by construction, so the only thing
   governance can change is *which* principal is asserted. It had to land before
   deploy, since code cannot be added to a deployed realm, and redeploying is
   expensive: `r/sys/validators/v0/cache.gno` hardcodes `const valopersRealmPath`.

8. **`updateInstructions` returns its error** instead of panicking, so a
   passed-but-refused proposal is deniable rather than stranded.

## Alternatives considered

- **Restore a `CurrentRealm()`-style check inside the handler.** Rejected:
  that is the mechanism just removed as bypassable, and it puts the gate in
  consumer code, so every consumer must get it right independently.
- **Only unexport `valopers.Auth()`.** Rejected: it fixes one consumer and
  leaves the package default unsafe for the next one. The leak-shaped
  hazard is the package default, not the getter.
- **Guard `updateInstructions` on the GovDAO principal.** Initially rejected
  as insufficient, now **adopted** as decision 4, because the objection was
  conditional on a leak that decision 3 removes. The objection: an attacker
  builds their own `dao.NewSimpleExecutor` around the privileged closure and
  calls the exported `Execute`, making `Previous()` be `gno.land/r/gov/dao`
  on demand. That requires *obtaining the closure*, which was possible only
  while `NewInstructionsProposalCallback` was exported. With the capability
  sealed, the principal check is no longer trivially forgeable and becomes
  meaningful defence-in-depth.

  Adopted as an authority repoint rather than as a hand-rolled
  `Previous().PkgPath()` comparison inside `updateInstructions`, so the rule
  lives in the authority object (and shows up in `Auth()`) instead of in an
  ad-hoc string compare.

  Its limit is stated in Context and in the `updateInstructions` godoc: it
  does not cover a future re-export of the closure itself, only of a plain
  crossing entrypoint. Partial, verified, and strictly better than the
  zero code-level guards the previous configuration provided.
- **Remove `authz` from `r/gnops/valopers` entirely.** Raised in review on the
  correct observation that the check compared valopers' address to valopers'
  address and could therefore never fire. Rejected in favour of the repoint:
  the diagnosis was right but the conclusion loses something the realm has no
  other mechanism for. Deleting the authority leaves `updateInstructions`
  protected only by being unexported and by the seal — adequate today, and
  with no guard at all against the plain-entrypoint regression. Repointing
  costs the same amount of code, keeps a machine-readable statement of the
  governance model in `Auth()`, keeps `Transfer` available as the
  delegation path, and turns the dead check into a live one.
- **Return a `dao.NewSafeExecutor`-wrapped executor from valopers.** Rejected
  in favour of the sealed request, for the fail-open reason in decision 3.
- **Seal the `Authority` interface to canonical implementations.** Rejected:
  unexported-marker sealing is bypassable via embedding in Gno (see
  `p/test/seal/filetests/z_seal_*_filetest.gno`), and the package is
  intentionally extensible — third-party implementations are the design
  intent (documented as the Class-3 residual on `NewWithAuthority`).

## Consequences

- **Behavior change for existing consumers.** A plain `NewContractAuthority`
  no longer accepts arbitrary proposers. Consumers that relied on the
  open-proposal default must migrate to
  `NewRestrictedContractAuthority(path, handler, NewAutoAcceptAuthority())`,
  which is unchanged and remains the documented escape hatch.

  Audited across the tree. Three **consumer** call sites construct a plain
  `NewContractAuthority`: `r/gnops/valopers` (migrated here — it is the
  contract's own authority, so the new default is what it wanted), the
  `r/moul/config` test under `examples/quarantined/`, and the `moul_authz`
  gnoland integration txtar. The txtar exercises the open-proposal DAO flow
  deliberately, so it moves to `NewRestrictedContractAuthority` +
  `AutoAccept`. Every other `authz` consumer uses
  `MemberAuthority`/`AutoAccept`, which this does not touch.

  That count deliberately excluded in-package examples and tests, and the
  first version of this change paid for it: `example_test.gno` also
  constructs plain `ContractAuthority`s, and its `Example_contractAuthority`
  paired one with a **foreign** path, producing a headline example whose
  privileged action can never run. Fixed here, and now executed by
  `p/moul/authz/filetests/z_contract_authority_shape_filetest.gno` — see
  Coverage.

- **`Auth()`'s return type changes** from `*authz.Authorizer` to `string`,
  **`Valoper.Auth()` is removed** in favour of `Valoper.AuthOwner() address`,
  and **`NewInstructionsProposalCallback` is removed**. All three are
  breaking for out-of-tree callers; none has in-tree callers beyond the ones
  migrated here, and none has an on-chain caller on `gnoland1`.

- **`ContractAuthority.String()`'s output changes** for every consumer, from
  `contract_authority[contract=P]` to
  `contract_authority[contract=P,proposer=Q]`. Breaking for anything
  asserting on or parsing that string. In-tree that is
  `p/moul/authz`'s `TestAuthorityString`, `valopers`' `admin_test.gno` and
  `z_auth_readonly_filetest.gno`, all updated here.

- **`r/gnops/valopers` imports `r/gov/dao`.** No cycle, but it is a
  dependency-order change in a genesis package, and the repoint in decision 4
  makes the dependency semantic rather than incidental: the realm now names
  GovDAO's package path in its authority. Moving governance to a different
  realm path would require a rotation (below) or a redeploy.

- **Residual, documented but not closed here.** `Transfer` and
  `DoByPrevious` derive the principal as `rlm.Previous().Address()`. If a
  contract that *is* the authority crosses into an untrusted realm while its
  `Authorizer` is reachable, then inside that realm `Previous()` is the
  contract and `caller == contractAddr`. The gate holds for every external
  caller that has not been called into by the authority contract itself.
  This does not affect valopers (layer 2 removes the handle), but consumers
  exporting an `Authorizer` should know the boundary.

  **Nothing pins this residual.** `TestForgedCallerCannotTransfer`, which an
  earlier draft of this ADR cited, pins a different property: that raw
  `Authorize` with a forged caller cannot mutate the *installed* authority.

- **A foreign-path `ContractAuthority` is not a one-way door**, contrary to
  what an earlier draft implied. `DoByCurrent` against it is dead, but the
  foreign realm can still drive and rotate it via `Previous()` once it
  crosses in — which is the intended async-DAO shape. What is lost is the
  original owner's ability to take it back:

  ```
  1. Transfer INTO foreign-path ContractAuthority: err = <nil>
  2. DoByCurrent after switch:                     err = unauthorized (action ran: false)
  3. owner Transfer back out:                      err = unauthorized
  4. foreign realm Transfer back out:              err = <nil>
  ```

- **The valopers authority is rotatable by GovDAO (decision 7).** This
  residual was open for three review rounds and is now closed: leaving it
  open sat badly next to decision 1's rationale for panicking on a nil
  handler ("a permanent brick with no on-chain recovery"), and decision 7's
  discovery that a well-formed but wrong path is silently dead made a
  recovery path load-bearing rather than a nicety.

  What decision 4 changes: under own-path + `DoByCurrent`, a rotation
  entrypoint needed a same-package `cross()` call to make
  `Previous() == PackageAddress(ownPath)`. Repointed at `r/gov/dao`, an
  unexported `rotate(cur realm, newAuth authz.Authority)` sealed in a
  `dao.ProposalRequest` authorizes naturally, because inside a
  GovDAO-executed closure `Previous()` **is** `r/gov/dao` — the same frame
  identity `updateInstructions` already relies on. So the mechanism is now
  ordinary rather than a trick.

  It is still not added here, deliberately: it widens an embargoed fix with a
  new privileged entrypoint that would itself need review. It must be added
  *before* deploy if wanted, since code cannot be added to a deployed realm.
  The redeploy alternative is more expensive than it looks —
  `r/sys/validators/v0/cache.gno` hardcodes
  `const valopersRealmPath = "gno.land/r/gnops/valopers"`, so a new pkgpath
  means editing a second genesis realm and re-registering every valoper.

- **A further residual is out of scope and not addressed here.** Every check
  above reads identity off a `realm` interface value through virtual
  dispatch (`rlm.Address()`, `rlm.Previous()`), under an `rlm.IsCurrent()`
  liveness check. That pairing is only sound while liveness and identity are
  guaranteed to come from the same concrete runtime realm value. Hardening
  that guarantee belongs in the VM, not in this package.

- **Coverage.** `p/moul/authz` gains a regression suite for the
  contract-identity gate: default-proposer accept/reject, external
  `Transfer` rejected with the authority left intact, external
  `DoByPrevious` rejected, contract self-rotation still allowed,
  forged-caller boundary, contract-path isolation, the
  `NewRestrictedContractAuthority` + `AutoAccept` escape hatch, and the
  nil-handler brick.

  Reverting the gate failed **6** of them when the gate first landed (see
  Round 4 verification below, where it now fails 10):
  `TestDefaultProposerIsContractOnly`, `TestExternalTransferRejected`,
  `TestExternalDoByPreviousRejected`, `TestContractPathIsolation`,
  `TestForeignFrameCannotDriveContractAuthority` and
  `TestContractAuthorityUnauthorizedCaller`. (An earlier draft said 5, under
  names that do not exist.)

  Consumer-level coverage for the repoint and the bundled leak:

  - `r/gnops/valopers/admin_test.gno` — `TestAuthDescribesGovDAOGate` pins
    `init.gno`'s configuration *including the proposer*, and
    `TestUpdateInstructionsRejectsNonGovDAO` pins that a non-GovDAO principal
    is refused. The latter is the assertion the previous configuration could
    not express: under own-path + `DoByCurrent` that exact call **succeeded**,
    which is why the old test asserted `NotPanics` and pinned nothing.
  - `filetests/z_govdao_only_principal_filetest.gno` — the same property from
    a foreign realm, plus the truthful `Auth()` description.
  - `filetests/z_operator_authlist_capability_filetest.gno` — the sibling
    leak: an operator calling a hostile realm once no longer lets it join the
    operator's auth list.

  Both reverts were checked rather than assumed. Reverting the repoint fails
  `TestAuthDescribesGovDAOGate`, `TestUpdateInstructionsRejectsNonGovDAO`,
  `z_auth_readonly_filetest.gno` and `z_govdao_only_principal_filetest.gno`.
  Reverting the `String()` proposer rendering fails `TestAuthorityString`,
  `TestAuthDescribesGovDAOGate`, `z_auth_readonly_filetest.gno` and
  `z_govdao_only_principal_filetest.gno`.

  Two coverage gaps found in review are closed:

  - `r/gnops/valopers/filetests/z_auth_readonly_filetest.gno` used to pass
    with the fix fully reverted, because `ContractAuthority.String()` was
    byte-identical either way — so it pinned `Auth()`'s return type and
    nothing about reachability. Now that `String()` renders the proposer it
    also pins the gate's shape, and it goes red on either revert above. The
    consumer-level regression this was actually about lives in
    `z_foreign_realm_capability_filetest.gno`: a foreign realm cannot reach
    the executor, cannot get it adopted, and leaves `instructions` untouched.

    Honest caveat, unchanged: for both that file and
    `z_operator_authlist_capability_filetest.gno`, the removed capability
    means the *attack itself* no longer compiles, so a revert fails them on a
    build error rather than an assertion. Each therefore also drives a
    runtime path — `MustCreateProposal` in the first, the exported
    `AddToAuthList` wrapper in the second — and spells the invariant out in
    its header for whoever edits the realm next.
  - `Example_*` functions that take **any parameter** are never executed by
    `gno test` — a sentinel `panic()` inside one still reports `ok`. The
    predicate is `isExampleFunc` in `gnovm/pkg/test/test.go`, which rejects on
    method receiver, parameters or results; `// Output:` plays no part in it,
    so adding one does not make the body run. (An earlier draft stated the
    condition as a conjunction of both, which would send a maintainer down a
    dead end.) The package's headline example was therefore invisible to CI.
    It cannot be covered by a unit test either: a closure declared inside
    `p/moul/authz` always presents that package's address regardless of
    `testing.SetRealm`, for the same declaring-package reason described in
    Context. `filetests/z_contract_authority_shape_filetest.gno` runs the
    shape from a real realm.

  - `Example_switchingAuthority` seeded its member set with `cur.Address()`
    while `Transfer` derives its principal as `cur.Previous().Address()`, so
    the `panic(err)` added in review sat on a call that could never
    authorize: the documented switching recipe aborted for every external
    caller, and a realm copying it into `init(cur realm)` would fail its
    deploy transaction. This is the same defect class as the
    `Example_contractAuthority` one above, in the sibling example, and it
    shipped green for exactly the `isExampleFunc` reason. Fixed to
    `cur.Previous().Address()`.

  Pre-existing tests that asserted the removed mechanism were updated to
  assert the new upstream rejection instead — in `p/moul/authz` itself and in
  the `r/moul/config` consumer test. `r/gnops/valopers/admin_test.gno` no
  longer reassigns `auth` before asserting, so it actually exercises
  `init.gno`'s configuration.

### Round 4 verification

**The identity-gate revert now fails 10 artifacts, not 6.** Replacing the
default proposer with `NewAutoAcceptAuthority()` fails the original six
(`TestDefaultProposerIsContractOnly`, `TestExternalTransferRejected`,
`TestExternalDoByPreviousRejected`, `TestContractPathIsolation`,
`TestForeignFrameCannotDriveContractAuthority`,
`TestContractAuthorityUnauthorizedCaller`) plus `TestAuthorityString`,
`TestExplicitProposerReplacesIdentityGate`,
`TestSpoofedProposerCannotImpersonateContractIdentity` and
`filetests/z_contract_authority_shape_filetest.gno` — the last four because the
rendered proposer is now part of what is pinned.

**The spoofing proposer is caught by 6 consumer artifacts, not 1.** Installing a
wide-open proposer in `r/gnops/valopers/init.gno` whose `String()` returns
`"contract-identity"` previously left four of five artifacts green, with only
`TestUpdateInstructionsRejectsNonGovDAO` red. It now fails
`TestAuthDescribesGovDAOGate`, `TestUpdateInstructionsRejectsNonGovDAO`,
`TestRotateAuthorityRejectsNonGovDAO`, `TestRotationProposalRequestIsSealed`,
`z_auth_readonly_filetest.gno` and `z_govdao_only_principal_filetest.gno`.

**Both executor-forgery shapes were reproduced before the gate and are refused
after it,** against victim realms written exactly as the `NewContractAuthority`
godoc instructs:

```
# victim exports only `func Reset(cur realm) error`
control (direct call): unauthorized
exploit Execute:  before -> after exploit: RESET
                  after  -> execution denied: executors are only invocable by gno.land/r/gov/dao
                            after exploit: genesis

# victim exports its *authz.Authorizer
exploit Execute:  before -> authority after: auto_accept_authority; Value: PWNED
                  after  -> authority after: contract_authority[...,proposer=contract-identity]
                            outsider write: unauthorized; Value: genesis
```

Pinned in-tree by
`r/gov/dao/impl/v0/filetests/executor_invocation_authority_filetest.gno`, which
adapts the characterization test into its inverted, regression form
(same file name and section numbering, so the two are diffable). Its section 1
keeps the approved propose → vote → execute path green, which is the thing a
gate on a shared primitive most risks breaking.

**`z_govdao_only_principal_filetest.gno` now attacks.** Its previous body
asserted a rendered string plus "instructions were not defaced" without ever
attempting a write, so `instructions touched: false` was vacuously true and the
file stayed green with the authority swapped for a fully permissive one. It now
drives the two routes an attacker has — the forged-principal executor and
`MustCreateProposal` — and asserts the refusal of each.

**Cross-transaction coverage.** Filetests run one VM session and so cannot
observe a realm reload, which matters because the authority's proposer is an
object in realm storage. `gno.land/pkg/integration/testdata/valopers_governance.txtar`
runs on a real chain, one block per step: `Auth()` after a reload, propose →
vote → execute rewriting `instructions`, propose → vote → execute rotating the
authority, and then the old principal being refused *and the proposal retired*
via `ExecuteOrRejectProposal` — which is the observable form of decision 8. A
panic there would abort the transaction and roll the `Denied` write back.

**Blast radius of the executor gate, measured.** Full `gno test ./...` over
`examples/` passes. Eleven call sites needed a one-line
`testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao"))` to stand in for
the proxy — ten in `r/sys/validators/v0/proposal_test.gno` and two
`r/gov/dao/impl/v0` filetests that call `impl.ExecuteProposal` directly. No
production call site changed, and one pre-existing site already stood in as
`gno.land/r/gov/dao/impl/v0`, which is why the accepted set includes
subpackages.
