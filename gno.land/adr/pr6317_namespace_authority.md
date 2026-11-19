# ADR: Namespace authority in the keeper

Status: **proposed**, for discussion. No code yet.

This ADR is phase 2 of a three-phase plan agreed on #6317 and #6341:

| Phase | What ships | How | Where |
|---|---|---|---|
| 1 | `r/sys/teams` and a names realm that reads it | realm deploys and one `sysnames_pkgpath` vote, no release | #6341 |
| 2 | one ownership resolver in the keeper, a governance override, `AuthorityOf` | one halt-height binary upgrade | this ADR |
| 3 | moving a name to a new address, and an in-realm admin convention | separate ADR | later |

The earlier version of this ADR designed phase 3 first (id and rev on user
records, offer and accept, a T1 transfer, a replay activation). It is kept in
this PR's history. A short summary is under
[Deferred to phase 3](#deferred-to-phase-3).

## Context

Three things are fused today, and naming them separately is what makes the
problem tractable:

| | Question it answers | Lives in |
|---|---|---|
| **Identity** | who is `@samcrew`? | `r/sys/users` (`ResolveName`) |
| **Namespace authority** | who may act under `r/samcrew/...`? | `r/sys/names`, and `r/sys/teams` after phase 1 |
| **Package ownership** | who may replace a live private package? | the keeper |

Most of the cases raised so far only need namespace authority to change:
rotating a key, changing a multisig's quorum, publishing as a team, and
recovering a namespace whose key is lost. Phase 1 covers the first three for
**new** paths, using realms only. It cannot reach two cases, and this ADR
exists for them.

### Who decides today

| Operation | Rule | Location |
|---|---|---|
| New deploy | `names.IsAuthorizedAddressForNamespace(creator, ns)` | `checkNamespacePermission`, `keeper.go:480`, called at `:1015` (park) and `:1137` (normal) |
| Replace a live private package | the stamped creator, **then** the namespace check | `checkRedeployPermission`, `keeper.go:723`, called at `:827` before the policy branch |
| Activate a parked package | namespace check on the parked creator; over a live private package, also parked creator == live creator | `EnablePackage`, `keeper_inert.go:258`, `:296` |
| Park over another submitter's parked package | refused | park guard, `keeper.go` |
| Reject a parked package | parked submitter, an approver, or the live package's stamped creator | `RejectPackage`, `isLivePackageOwner`, `keeper_inert.go:481`, `:665` |

`checkRedeployPermission` is pure Go with no realm call. It was added by #6171
(`pr6171_private_redeploy_creator_binding.md`) to close a takeover: any funded
address could replace a live private realm, re-run its `init()` as itself and
become its `ownable` owner.

### What phase 1 cannot do

Measured on #6341's branch, with a txtar driving `gnokey` against a running
`gnoland`:

- **No realm can let anyone but the stamped creator replace a live private
  package.** The creator check never consults a realm. On mainnet, `samcrew`
  holds four private realms stamped to the multisig it needs to replace.
- **Governance cannot reach a namespace.** `r/sys/teams` accepts writes only
  from `r/<ns>/home`, so a T1 member is refused from `maketx run` and from an
  executor realm voted through a real proposal. A namespace whose key is lost,
  or whose `home` is compromised, has no exit.
- **The two checks fire in the order the code runs, not the order one
  expects.** After retiring, an old key that is still the stamped creator
  passes the creator check at `:827` and is stopped later by the namespace
  check at `:1137`. A fix that only generalises the stamp leaves half the
  decision where it was.
- **Policy in `r/sys/names` is paid for on every deploy.** The same override,
  built in the realm and in the keeper, on identical package bytes:

  | `addpkg` | base | override in keeper | override in `r/sys/names` |
  |---|---|---|---|
  | before `Enable` | 3,633,015 | 3,633,015 | 5,071,664 |
  | after `Enable`, address namespace | 3,648,721 | 3,648,721 | 5,087,370 |

  On both measured workloads the realm version adds 1,438,649 gas (+39.6%),
  the cost of loading a larger realm, which the keeper does on every deploy.
  Two workloads do not give an exact figure for every deploy, but the cost is
  structural: it grows with the size of the realm every deploy loads.

### Mainnet facts this design relies on

- `gnoland-1` runs the inert submission policy: a post-genesis deploy is
  parked and goes live when an approver (the gpao oracle) sends
  `MsgEnablePackage`.
- Upgrades are **halt-height binary swaps** voted by GovDAO
  (`misc/deployments/mainnet.gno.land/UPGRADES.md`). `node:p:halt_min_version`
  is a startup gate, not a guarantee about which binary executes old blocks:
  a node may sync from genesis with a later binary (verified with `v1.5.0`
  alone, genesis to tip). New rules must therefore activate on a fact in
  chain state, never merely on the binary version (section 8).
- `WillSetParam` panics on any unknown `vm:p:` key (`params.go:607`). A
  proposal that sets this ADR's parameters before the upgrade therefore fails
  when its executor writes the value. That proves the keys are absent from
  committed state, but not that no historical block *tried* to write them: a
  failed execution is still in a block. The new binary must keep failing those
  writes when it re-executes old blocks (section 8).
- Parameter proposals built by `r/sys/params` have no voter filter
  (`newPropRequest`, `params.gno:126`), and only `r/sys/params` can write
  parameters (`assertSysParamsRealm`, `gnovm/stdlibs/sys/params/params.go:101`,
  with the path a literal in the binary).

## Decision

### 1. One resolver, in the keeper

Every ownership-gated decision goes through one function:

```go
// effectiveAuthority decides whether addr may perform op at pkgPath.
// It replaces checkNamespacePermission and checkRedeployPermission on every
// path that runs them, and the owner branch of RejectPackage.
func (vm *VMKeeper) effectiveAuthority(ctx sdk.Context, st gno.TransactionStore,
    params Params, pkgPath string, op AuthorityOp, addr crypto.Address,
    live, parked *PackageFacts) error

type AuthorityOp int // OpDeploy, OpReplace, OpEnable, OpReject

type PackageFacts struct {
    Private bool
    Creator string // the raw [addpkg] creator string from the stored gnomod.toml
}
```

`Creator` stays the raw stamp. The off and legacy rules compare it exactly
as #6171 does, so a malformed or non-canonical stamp still fails closed
there. Under the pinned and delegated rules, a missing or malformed stamp
matches nobody: the creator branch is false, and only an admin or a pinned
address can replace the package.

- **Where it runs.**
  - `AddPackage`: in place of the creator check at `:827`, before the policy
    branch, so it covers the normal and the park path at once.
  - `EnablePackage`: for the parked creator, in place of `:258` and `:296`.
  - `RejectPackage`: only for a sender who is neither the parked submitter nor
    an approver. Those two keep their unconditional right to reject.
- **When it runs.** Before the live source is deleted, before compiling, and
  before storage-deposit accounting. A refused submitter gets no free compile
  (#6171).
- **What stays outside it.** Public packages are still never replaced, and a
  private package still cannot become public. The CLA check, the chain-domain
  check, the submission policy and the park guard against another submitter's
  parked bytes are all unchanged.
- **Read-only.** Any realm call it makes runs on a fork that sees the
  transaction's earlier writes and is discarded afterwards:
  - both the base and the IAVL store are cache-wrapped, because
    `BeginTransaction` alone reuses the backing stores and `SetObject`
    writes through immediately;
  - the fork gets fresh object and realm caches, and neither its stores nor
    its caches are ever published;
  - parameter and bank writes are refused, and events dropped.

  `callRealmBool` today hands the callee writable adapters and the live store
  (its comment at `keeper.go:408` warns about this).

### 2. The rules

Before activation (section 8), today's code runs unchanged. After it, a
namespace is in one of four states, checked in this order.

**Pinned.** The namespace has an entry in the governance override (section 3).
The entry is **terminal**, and it is checked first, so repointing or
disabling the names realm cannot get around it:
- `OpDeploy`, `OpReplace`, `OpEnable` and owner-standing `OpReject` are
  allowed exactly for the addresses listed in the entry;
- the registered holder, the team and the stamped creator count for nothing;
- an entry with no addresses freezes the namespace: nothing is deployed,
  replaced or enabled there.

The keeper decides this in Go, then applies pause (section 6). A compromised
`home` cannot add anyone back, because nothing it writes is read while the
pin stands.

**Off.** Not pinned, and `SysNamesPkgPath` points at an undeployed path, or
the names realm reports `IsEnabled() == false`. Today's rules apply byte for
byte:
- anyone may deploy a new path;
- only the stamped creator may replace a live private package;
- the stamped creator may reject parked bytes.

**Delegated.** The names realm declares `AuthorityFor` (section 4). Writing
`role(a)` for `AuthorityFor(a, ns)`:

| Operation | Allowed when |
|---|---|
| `OpDeploy` (nothing live) | `IsAuthorizedAddressForNamespace(addr, ns)`, as today |
| `OpReplace` (live private) | `role(addr) == "admin"`, or `addr == live.Creator && role(addr) != "none"` |
| `OpEnable` over a live private package | the same rule, applied to the parked creator at activation time |
| `OpEnable` with nothing live | the `OpDeploy` rule, applied to the parked creator |
| `OpReject` by owner standing | `role(addr) == "admin"`, or `addr == live.Creator && role(addr) != "none"` |

- **`r/<ns>/home` is admin-only.** Replacing or activating over `home`
  requires `role(addr) == "admin"`, with no creator branch. `home` writes the
  team, so whoever can replace it can promote themselves. A member who
  happens to be `home`'s stamped creator must not keep that power.
- A team member can replace other realms they deployed themselves, and only
  an admin can replace someone else's. Members cannot take over each other's
  realms.
- **Pause** is checked on every row except rejection: replacement and
  activation call `IsPaused()` when they do not go through
  `IsAuthorizedAddressForNamespace`.
- A retired holder that is not a team member answers `none`, so retirement
  removes its deploy, redeploy, activation and rejection rights together. A
  retired holder that is also a team member keeps its member rights. A
  migration must therefore remove the old key from the team, not only retire
  it, and `AuthorityOf` shows it until it does.

**Legacy.** The names realm is enabled but does not declare `AuthorityFor`.
This is `r/sys/names` as deployed today. Today's rules apply byte for byte.

Detecting `AuthorityFor`:
- the keeper reads the configured package's production files from the
  transaction's store (`GetMemPackage`) and looks for a top-level
  `func AuthorityFor(addr address, namespace string) string`;
- it never calls the function speculatively, and never treats a panic as a
  missing function;
- a declaration with any other signature, a panic, or a result other than
  `none`, `member` or `admin` refuses the operation; it never falls back to
  the legacy rule;
- parsing is metered.

The extra realm call happens only on `OpReplace`, `OpEnable` over a live
package, and owner-standing `OpReject`. An ordinary deploy makes no new realm
call; it adds only the bounded parameter read in section 3.

### 3. The governance override

A new parameter, `vm:p:namespace_authority` (`[]string`). Each entry is

```
<namespace>=<address>[,<address>...]
```

for example `samcrew=g1new...` or `samcrew=` (frozen).

- **Validation.** One strict parser, shared by `WillSetParam` and the
  dedicated validator used on genesis import (`validateAuthorityParams`;
  `Params.Validate` cannot see fields outside `Params`), rejects:
  - a malformed entry or an invalid address;
  - two entries for the same namespace;
  - more than 256 entries, more than 16 addresses in one entry, or a
    namespace longer than 64 bytes (the namespace pattern itself has no
    length bound);
  - a total encoded length over 16 KiB.
- **Storage and genesis.** Both new parameters live outside the `Params`
  struct, so `SetStruct` never writes them and the original genesis replays
  without new keys. An absent key means "empty" and "inactive".
  - They get their own genesis fields, validated by the same parser on
    import and written by `ExportGenesis`.
  - The resolver reads them only on ownership-gated paths, parses only the
    entry that matches the namespace, and every transaction's ante read is
    left alone.
- **Cost.** An ordinary deploy under an unpinned namespace adds one bounded
  parameter read and a scan for its namespace. That fast path, and the
  maximum-size list, are measured before any gas claim is made.
- **Name-shaped namespaces only.** An address namespace (`r/g1.../...`)
  cannot be pinned, so governance cannot take over a keyholder's own address
  namespace. Whether it should be allowed is an open question.
- **Written by an ordinary parameter proposal**, using the existing
  `NewSysParamStringsPropRequest*` factories.
  - `UpdateStrings` removes exact duplicates before validation, so adding the
    same entry twice is a no-op.
  - Adding a *different* entry for a namespace that is already listed fails
    validation. Two competing pins for one namespace therefore cannot both
    pass.
  - **Changing** a pin's addresses uses the full-set factory, so the
    namespace is never briefly unpinned between a remove and an add.
  - The generic factories let a proposal set its title but not its
    description. Ship a helper script under `misc/govdao-scripts/` whose
    title states the change (for example "pin samcrew to g1…"). Voters read
    the full current value with `gnokey query params/vm:p:namespace_authority`.
- **The whole GovDAO votes on it**, not T1 only, because the generic factory
  has no voter filter. A T1-only route would need a new setter in
  `r/sys/params`, a public realm that cannot be redeployed, and so a replay
  upgrade. `AuthorityOf` states the power openly (section 5).
- **No recovery realm.** A recovery realm could itself be badly deployed,
  compromised, or pinned to a version a retired namespace cannot escape. The
  parameter path is in the binary.
- **Precedence is scoped.** The override outranks the names realm **only**
  for namespaces it lists. Everywhere else the names realm decides, so
  repointing `sysnames_pkgpath` stays available to governance as a same-day
  fix for every namespace nobody has pinned.
- **Unpinning** returns the namespace to whatever the other rules say at
  that moment.
  - `WillSetParam` refuses a write that drops a namespace's entry unless,
    when it executes, the configured names realm is enabled and declares
    `AuthorityFor`. An unpin can therefore only fall back to the delegated
    rules, never to legacy or off. That check reads the realm through the
    same isolated fork as the resolver (section 1).
  - Before voting, the pinned addresses repair `home` and its team, which
    they can do while pinned when `home` is private. `AuthorityOf` shows
    the principals the namespace would fall back to.
  - Team state can still change between the vote and its execution. That is
    the same `home` the pinned addresses just repaired, and `AuthorityOf`
    after execution shows the result.
  - A namespace whose `home` is public cannot repair its team by replacing
    `home`, so its pin may have to be permanent.

### 4. What the names realm must provide

Phase 1 (#6341) ships the names realm and `r/sys/teams`. Both are public,
and `r/sys/teams` holds state that cannot be migrated later, so what phase 2
needs must be in them **before they are deployed**:

```go
// In the names realm. Pure reads.
func AuthorityFor(addr address, namespace string) string // "admin", "member" or "none"
func IsPaused() bool
func IsEnabled() bool
```

`r/sys/teams` needs:
- a role per member, `admin` or `member`, given explicitly on every write.
  `SeedMembersOnce` writes the seeded addresses as admins;
- `Retire` refused unless the namespace has at least one admin, and removing
  or demoting the last admin of a retired namespace refused;
- paged getters: `Members(ns, offset, limit)` returning addresses with their
  roles, and `IsRetired(ns)`.

The names realm also exports `DescribeAuthority(namespace string) string`, a
bounded JSON summary (holder, retirement, member count and the first page of
members) that `vm/qauthority` includes. Larger teams are paged through the
`r/sys/teams` getters.

`AuthorityFor` answers, for a namespace `ns`:
- **An address namespace** (`ns` is an address): that address is `admin`.
  Team rows for it count as for any namespace. Users records play no part.
- **A current name of a live record:**
  - the holder is `admin` while `ns` has not retired it;
  - team admins are `admin`, and other team members are `member`.
- **An alias** (a name the record has renamed away from): the same, using the
  alias's **own** team rows and retirement, never the renamed-to name's.
  Retiring a name under one spelling must not be undone by looking it up
  under another. `IsAuthorizedAddressForNamespace` must also refuse new paths
  under an alias for team members, not only for the holder.
- **A deleted record** (`users.IsNameTaken(ns)` is true but
  `ResolveName(ns)` is nil): `none` for everyone.
  `IsAuthorizedAddressForNamespace` must refuse it too, including for team
  members, so first deploys and first activations are frozen as well as
  replacements. Only a pin reaches it.
- **An unregistered name**: team rows only. With no record, nobody can have
  deployed `home` while names was enabled.
- Pause and `Enable` do not change the answer; the keeper reads them
  separately.

### 5. `AuthorityOf`, a query

A new keeper query, `vm/qauthority`, returns for a namespace:

- whether it is pinned, and the pinned addresses;
- the configured names realm, whether it is enabled and paused, and which
  rule applies (off, pinned, delegated, legacy);
- the identity holder (`ResolveName`) and whether the namespace has retired
  it;
- the team members and their roles, read through the names realm;
- each live package under the namespace with its visibility and stamped
  creator, and each parked submission with its creator and height (paginated);
- the number of live realms, all of which may hold an in-realm admin the chain
  cannot read;
- the standing powers outside the namespace's control:
  - GovDAO, through this parameter and through `sysnames_pkgpath`;
  - the `r/sys/users` controllers (`users.Controllers()`), which can rename
    or delete the record;
  - the stamped-creator fallback while names is off.

The guarantees it lets a namespace state:
- **Delegated:** after the old key is retired and removed from the team, the
  principals that can deploy, redeploy or activate anything under the
  namespace are its team, within each member's role, plus GovDAO through a
  pin or a names-realm change.
- **Pinned:** exactly the pinned addresses, or nobody for an empty pin. The
  team counts for nothing until unpinned.
- **Legacy or off:** today's stamped-creator rules, which governance can
  restore by repointing the names realm. `AuthorityOf` reports which state
  applies.

Per operation, `AuthorityOf` also reports the blockers that apply: pause,
freeze, deletion, an alias refusing new paths.

It does not cover in-realm admins. It reports how many realms might hold one
rather than claiming none do. Standardising a way for realms to declare their
admins is phase 3 work.

### 6. Pause

Pause today lives inside `IsAuthorizedAddressForNamespace`.

- **Delegated and legacy namespaces.** Pause behaves as today: the names realm
  refuses deploys while paused.
- **Delegated replacement and activation.** These do not go through
  `IsAuthorizedAddressForNamespace`, so the keeper calls `IsPaused()` itself.
- **Pinned namespaces.** The keeper calls `IsPaused()` when the names realm is
  deployed and enabled, and refuses `OpDeploy`, `OpReplace` and `OpEnable`
  while it returns true. A missing `IsPaused` refuses. When names is off,
  there is no pause, as today. That realm load is paid only by operations
  under a pinned namespace.
- **Rejection** is never paused.

### 7. Approvals bind the creator

Content hashes exclude `[addpkg]` (`pkghash.go:31`). Under this ADR, two
different authorized addresses can park the same bytes, and `init()` runs as
whichever one parked last. Approvals therefore need a way to pin the creator.

- Add `PkgCreator crypto.Address` to `MsgEnablePackage`:
  - after `PkgHeight`, tagged
    `json:"pkg_creator,omitempty" yaml:"pkg_creator,omitempty"`, so messages
    that leave it out keep their sign bytes;
  - with matching fields in `vm.proto` and `pb3_gen.go`, keeping fields 1–4;
  - a non-zero value that does not match the parked creator is refused before
    type-checking.
- After activation (section 8), `PkgCreator` is **required** on every
  `MsgEnablePackage` outside genesis replay. A narrower rule would leave a
  gap: on a first deploy, an admin can reject A's reviewed submission and
  park the same bytes as B within the same block, with the same hash and
  height.
- **Every approval producer** must supply it before activation:
  - `gnokey maketx enablepkg` gains a `--pkgcreator` flag;
  - gpao reads the creator from `MsgAddPackage.Creator` and signs it with the
    reviewed bytes and height, in its normal, recovery and re-offer paths
    alike;
  - gpao keys its deduplication on creator and height as well as content,
    and keeps its content-verification cache as it is;
  - approvals queued before activation without `PkgCreator` are re-signed.

  The activation proposal confirms that the configured `pkg_approvers` run
  updated software.
- **`QueryPackageMeta`** reports the parked submission's creator and height
  separately from the live package's, and reports the resolver's blockers for
  first deploys and replacements alike.

The same creator re-parking can still change its own deposit ceiling. That
only affects what that creator pays.

### 8. Rollout

1. **Phase 1 lands first**, with the interfaces in section 4 already in
   place. The current keeper never calls `AuthorityFor`.
2. **Audit before activation.** List:
   - every live private package whose stamped creator differs from its
     namespace's admins under the new rules;
   - every existing `home`, and whether it is private;
   - every unregistered namespace holding packages.

   The activation proposal states each authority change the new rules cause,
   so voters approve them explicitly rather than inheriting them.
3. **One halt-height upgrade**, with `halt_min_version` set to the release.
   The new binary runs today's rules until activation.
4. **The new keys are unknown below the upgrade height.** The release
   records the upgrade's halt height per chain id, in `gno.land/pkg/upgrades`
   next to the existing upgrade rules. At or below that height, `WillSetParam`
   panics on both new keys exactly as the old binary did (it receives the SDK
   context, so `ctx.BlockHeight()` is available). A historical block that
   tried to write them therefore fails the same way when re-executed by the
   new binary. A chain with no recorded height, such as a new testnet,
   recognizes them from genesis.
5. **Activation is a fact in chain state.** A second new parameter,
   `vm:p:authority_activation_height` (`int64`, absent = inactive), is set
   by a GovDAO proposal after the upgrade. The new rules, and the
   `PkgCreator` requirement, apply only at heights at or above it.
   - Below it, including when a node syncs old blocks with the new binary,
     today's rules run, gas included.
   - `WillSetParam` refuses a height at or below the current block, and
     refuses changing the value once set.
   - `WillSetParam` refuses `vm:p:namespace_authority` until activation has
     been set.
6. **Genesis transactions.** `IsGenesisReplay` is true for every
   transaction delivered in InitChain: both a hardfork's replayed history and
   a fresh chain's bootstrap transactions.
   - The ownership checks this ADR adds are waived there, as #6171's check
     is. Replay re-delivers only transactions the source chain accepted, and
     bootstrap transactions are reviewed with the genesis.
   - The `PkgCreator` requirement is waived there too.
   - Parameter validation is **not** waived: a genesis that sets either key
     is validated by the same parser.

### 9. Worked example: `aib` and `samcrew` on mainnet

Queried on `gnoland-1`, 2026-10-15:
- `aib` resolves to multisig `g1gkqe9…sjsl` and has nothing deployed.
- `samcrew` resolves to multisig `g136j0…5cpf`. Four private realms are
  stamped to it: `r/samcrew/home`, `barricade`, `block_party`,
  `space_invaders`. About 34 public packages are stamped to it as well.

**`aib`: phase 1 only.**
1. The old multisig deploys `r/aib/home`, whose `init()` seeds the new
   multisig as admin.
2. Through `home`, the old multisig retires itself.

**`samcrew`: phase 1, then phase 2.**
1. The old multisig redeploys its private `r/samcrew/home` with the teams
   code. Today's rules allow this, because it is both the stamped creator and
   the holder. It seeds the new multisig as admin.
2. It does **not** retire yet. Retiring now would freeze all four private
   realms, because nobody else passes the creator check under the current
   binary.
3. After the phase 2 upgrade and activation, the new multisig, as admin,
   replaces the four private realms. Each `init()` runs again as the new
   multisig.
   - Kept: each realm's address, balance, object clock and storage
     accounting (`pr6164`).
   - Not kept: package globals are rebuilt by the new `init()`, and the
     previous deployment's objects become unreachable.
   - Each realm therefore needs either a migration in its new code or an
     accepted reset of its application state, decided realm by realm before
     replacing it.
4. The old multisig retires through `home` and is removed from the team.
   `AuthorityOf("samcrew")` then lists the new multisig and GovDAO as the
   only principals.
5. The public packages cannot be replaced under any design. Any in-realm
   admin they hold must be handed over by each realm's own function while the
   old multisig can still sign.

If an old multisig can no longer sign, one parameter proposal pins the
namespace to the new multisig. The new multisig then replaces the private
realms, including `home`, and can repair the team before an unpin vote.

## Authorization matrix

`live.C` is the live package's stamped creator. `role` is the answer of
`AuthorityFor`.

States are checked in the order pinned, off, delegated, legacy. Before
activation, every column behaves as "legacy" (or "off") does.

| | pinned | off | delegated | legacy |
|---|---|---|---|---|
| New deploy | listed addresses | anyone | names check | names check |
| Replace live private | listed addresses | `live.C` | admin, or `live.C` while not `none` | `live.C` and names check |
| Replace live private `home` | listed addresses | `live.C` | admin | `live.C` and names check |
| Enable, nothing live | parked creator listed | parked creator passes "New deploy" | same | same |
| Enable over live private | parked creator listed | parked creator == `live.C` | parked creator passes "Replace" | parked creator == `live.C` and names check |
| Reject parked | submitter, approver, listed addresses | submitter, approver, `live.C` | submitter, approver, admin, `live.C` while not `none` | submitter, approver, `live.C` |
| Paused | refused except reject | no effect | refused except reject | refused except reject |

## Alternatives considered

- **The override in `r/sys/names`.** It added 39.6% gas on both measured
  deploy workloads, a cost every deploy pays through the larger realm it
  loads, and it still cannot let anyone else replace a private realm.
  Rejected.
- **A stable facade realm that homes and the keeper both call.** It would load
  even more realm code on every deploy. Rejected for the same reason.
- **A recovery realm.** It can be badly deployed or compromised, and a retired
  namespace could be stuck with whichever version is live. Rejected in favour
  of a parameter.
- **A T1-only recovery vote.** It needs a new setter in the public
  `r/sys/params`, and so a replay upgrade. Rejected. The whole-DAO power is
  declared in `AuthorityOf`.
- **An override that is just one more path, OR'd with the others.** Built
  first that way, a compromised `home` re-added the retired key the next
  block. Rejected; a pin is terminal.
- **Any team member may replace any private package.** That is simpler, but a
  member could replace another member's realm and re-run its `init()` as
  themselves. Rejected.
- **Only admins may replace private packages.** Even a package's own creator
  would need admin rights to redeploy it. Rejected as unneeded friction.
- **A retirement flag on the `r/sys/users` record now**, so that `ResolveName`
  readers see it. It changes the public `r/sys/users`, which turns phase 2
  into a replay upgrade. Deferred to phase 3. `AuthorityOf` reports the
  retirement until then.
- **Governance deletes the old record**, so `ResolveName` returns nil for it
  at once. That is realm-only, but today a deleted name can never point
  anywhere again. Not chosen as the default. Governance may still do it.
- **Rewriting the stamped creator when authority changes.** The stamp is also
  the `init()` identity and the deposit payer. Rejected; a replacement stamps
  its own submitter.

## Consequences

- **This deliberately reverses part of #6171.** Its invariants that still
  hold:
  - the namespace and ownership facts are both evaluated, now in one place;
  - an error refuses and never falls back;
  - checks run before deletion and compilation;
  - the genesis-replay waiver stays;
  - the stamp remains the `init()` caller and deposit payer;
  - an in-realm `ownable` owner never grants redeploy rights.

  What changes: a namespace admin, or a pinned address, can now replace a
  private package someone else deployed. Under the off and legacy rules,
  #6171 holds unchanged.
- **Replacing a package re-runs `init()` as the replacer**
  (`pr6164_redeploy_continues_realm.md`). The realm keeps its address,
  balance, object clock and storage accounting. Its globals are rebuilt, and
  the previous objects become unreachable. An `ownable` owner set in `init()`
  follows the replacer, who can then spend the realm's balance through the
  realm's own code. That is the intended effect for recovery, and it is why
  ordinary members cannot replace each other's realms.
- **Deposits.** The replacer pays for new storage. On activation, the parked
  creator pays under their own ceiling. Existing deposits do not move, and
  objects orphaned by a redeploy stay charged. Refunds go to the transaction
  caller, or to `StorageFeeCollector` while ugnot is restricted.
- **GovDAO gains a direct power over every name-shaped namespace.** It
  already had the indirect one, through repointing `sysnames_pkgpath`. Both
  are listed in `AuthorityOf`.
- **Three powers stay outside the namespace until phase 3:**
  - `ResolveName` still answers the old address after retirement, so a realm
    or UI doing its own ACL on the name still obeys the old key;
  - in-realm admins of existing packages;
  - the `r/sys/users` controllers.

  `AuthorityOf` names all three, so a team cannot be read as claiming more
  than it holds.
- **The CLA still applies.** A new admin needs its own `r/sys/cla` signature
  before deploying, parking or replacing.
- **Address namespaces never change hands**, under this ADR or phase 1.

## Testing

- **Keeper:** every cell of the matrix, for both the normal and the inert
  path, plus:
  - the gate order: a retired stamped creator is refused at the single
    resolver call, before compiling;
  - activation: before `authority_activation_height`, and when syncing
    historical blocks with the new binary, results and gas match the
    previous binary;
  - pins outrank off: a frozen pin stays frozen after `sysnames_pkgpath` is
    pointed at an undeployed or disabled realm;
  - `home` replacement refused to a member who is its stamped creator but not
    an admin;
  - an address namespace's own address keeps replacement under delegation;
  - aliases use their own team and retirement, and refuse new paths for team
    members;
  - deleted names refuse first deploys and first activations to team
    members;
  - raw and non-canonical stamps: unchanged results under off and legacy, and
    no creator match under delegated and pinned;
  - the resolver's realm calls cannot persist Gno object writes, including
    through another realm;
  - detecting `AuthorityFor`: absent (legacy), correct, wrong signature, a
    panic, an unknown role, a test-only declaration ignored;
  - a pinned namespace ignores a `home` that re-adds a retired key;
  - a frozen pin refuses every operation except rejection by the submitter or
    an approver;
  - pause under each rule, including a pinned namespace and a missing
    `IsPaused`;
  - the resolver's realm calls cannot write parameters or the bank.
- **Parameters:** valid, malformed, duplicate, over each bound, an address
  namespace; an identical add is a no-op and a conflicting add fails; setting
  either parameter on the current binary fails; the activation height is
  refused at or below the current block and once set; an unpin is refused
  unless the fallback is delegated; a pin change by full set is atomic.
- **Historical execution:** a block below the upgrade height that tried to
  write either key fails identically under the old and the new binary;
  genesis with and without the keys, import validation, and export.
- **`PkgCreator`:** left out (refused after activation, accepted before),
  matching, mismatched, the same-block reject-and-repark, sign-byte equality,
  and the binary round trip.
- **gpao:** a new decision for identical bytes from a different creator,
  including within one block.
- **`QueryPackageMeta` and `vm/qauthority`:** each rule, with pagination.
- **Gas:** on identical workloads, measure an ordinary deploy before and
  after activation, the same with the override at its maximum size, and
  detecting `AuthorityFor` (`go test ./gno.land/pkg/sdk/vm/ -run Gas`).
  Before activation the numbers must match exactly.
- **Integration `.txtar`:** the `samcrew` sequence end to end, the lost-key
  pin, and unpinning after the team has been repaired.

## Deferred to phase 3

A separate ADR will move a name to a new address. The earlier version of this
ADR holds a full design, which that ADR can start from:
- id and rev on records;
- holder-led offer and accept;
- a governance transfer;
- id-bound proposals;
- a replay activation.

Points it must settle:
- **A waiting period on offers** (suggested in review): a few days between
  offer and accept, during which the current holder or governance can cancel.
  This gives the real owner time to react when a key is stolen.
- **Retirement visible to `ResolveName`**, and whether a retired record should
  answer with its team.
- **Deleted records**: whether governance can point a deleted name at a new
  address.
- **A convention for realms to declare their in-realm admins**, so
  `AuthorityOf` can report them.
- **Controller policy on `r/sys/users`**, now that controllers can affect
  namespace authority.

Mainnet upgrades by halt height, so phase 3 can weigh a one-off migration in
the binary at the halt against a full replay.

## Open questions

- Should governance be able to pin an address namespace, for a stolen key's
  own packages?
- Are 256 entries, 16 addresses per entry and 16 KiB in total the right
  bounds?
