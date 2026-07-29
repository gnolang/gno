# ADR: commondao least-authority proposal executor

## Context

Follow-up to the treasury ADR (`pr6012_commondao_treasury_sub.md`).
That design already minted each DAO's sub-identity and sent treasury
funds through it, but the executor itself still ran with the hosting
realm's full authority: `Execute` invoked the proposal executor as
`fn(cross(rlm))`, a **crossing** call that handed the executor the
realm's live `cur`. Nothing structural stopped an executor from minting
*any* sub the realm owns and moving value from *any* DAO — value moved
from the correct address only because each fund-moving executor
disciplined itself to mint its own DAO's sub. Least authority was a
convention, not a default. This ADR makes it the default — and, for
this realm's closed, audited executor set, a property (see Scope of the
guarantee).

## Decisions

### The executor type becomes non-crossing

`ExecFunc` changes from the crossing `func(realm) error` to the
non-crossing `func(int, realm) error` (interrealm v2: a leading
non-`realm` parameter makes the function type non-crossing;
`types.go`). The `int` is unused — its only job is to strip the
crossing property. The executor therefore no longer receives the
caller's `cur` (and cannot mint anything); it receives a plain
`sub realm` **value** passed in by the host.

### The host mints exactly the operative DAO's sub

The realm `Execute` wrapper picks the **operative DAO** and mints its
sub before calling into the package:

```
op := dao                                  // host DAO by default
if f, ok := p.Definition().(commondao.Funded); ok { op = f.FundingDAO() }
sub := cur.Sub(subpathOf(op.ID()))
err := dao.Execute(proposalID, sub)        // package calls fn(0, sub)
```

`Funded` is a new optional interface (`FundingDAO() *CommonDAO`) that a
fund-moving definition implements to name the DAO whose address funds
its executor: clawback names the **target**, sub-DAO dissolution names
the **dissolved descendant**, ordinary spend and every non-fund
executor default to the **host**. A non-fund executor still receives a
sub (the host's) and simply ignores it.

### Why this reduces authority

A sub-identity minted by `cur.Sub(subpath)` is **terminal** (it cannot
`.Sub()` again) and **RealmSend-only** (its banker can perform only
`BankerTypeRealmSend`, and `SendCoins` enforces `pkgAddr == from`).
So an executor that *uses* only its `sub` moves value from **exactly
one address** — the operative DAO's — and holds no primary cur. This is
a strict reduction from the previous `fn(cross(rlm))`, where the
executor received the realm's live **primary** `cur` and could mint any
sub the realm owns: escalation was the default, and least authority
depended on each fund-mover choosing to mint only its own DAO's sub.

### Scope of the guarantee (not a sandbox)

The guarantee is a *default*, not a hard sandbox against untrusted
executor code. `cross(rlm realm) realm` requires a live, current
`realm` value, and the only `cross`-able value an executor is handed is
its DAO's `sub` — `unsafe.CurrentRealm()` returns the unrelated
`runtime.Realm` type, which is rejected at type-check time (the builtin
`realm` is a sealed interface `runtime.Realm` cannot satisfy) and so can
never be passed to `cross`, and `realm` values cannot be persisted into
a definition field. An executor could
therefore regain the realm's **primary** authority only by an
*explicit* `cross(sub)` into one of the realm's crossing functions
(crossing semantics mint a fresh primary cur for the callee's realm
regardless of the `prev`). That is a visible, greppable call. The
reference realm's executor set is closed and realm-authored
(`mustPropose` is unexported; no exported entry accepts a user-supplied
`ProposalDefinition`/`ExecFunc`), and none of its executors cross back —
so for **this** realm the blast radius is one treasury. A downstream
realm that runs untrusted or user-registered executors does **not**
inherit that property and must not treat `sub` as a capability
boundary; the `ExecFunc` doc says so.

Address derivation is unchanged (`subpathOf` → `DerivePkgSubAddr`); the
mint simply moved one frame out, from inside each executor into the
`Execute` wrapper. `z_15_a` pins that the address is byte-identical
before and after. `assertRunningPath` already fails closed unless the
running path equals `pkgPath`, so the minted sub's address agrees with
the stored `daoAddress` const or the realm refuses to run at genesis.

### Residual author-time concern (not a runtime hole)

A mis-wired `FundingDAO()` would make the wrapper mint the *wrong*
DAO's sub and the executor would draw from that DAO's treasury. This is
**not** caught by `pkgAddr == from` (both sides are the minted sub's own
address, so the check is vacuous here), and it is **not** an authority
escalation (the realm may mint any of *its own* subs regardless; the
clawback proper-ancestry check still independently authorizes the
target). It is an author-time correctness obligation, pinned by the two
host≠operative fixtures where a wrong wiring would move real coins from
the wrong account: grandparent clawback (`z_12_i`, and the wider `z_12`
clawback family, which all fail on a mis-wire) and sub-DAO dissolution
(`z_13_a`).

## Alternatives considered

- **Keep the crossing executor (`fn(cross(rlm))`)** — rejected. It gives
  every executor the realm's primary `cur` by default, so least authority
  depends entirely on each executor's restraint. The non-crossing form
  flips the default: the executor gets only its DAO's `sub`, and reaching
  primary now requires an explicit, auditable `cross(sub)`.
- **Pass `cur` and let each executor mint its own sub** — rejected. That
  hands the executor the realm's full minting authority, the exact
  capability this change removes.
- **Thread the operative DAO id and mint inside the package** — rejected.
  The package has no `cur` (only the realm does), and keeping minting in
  the realm wrapper concentrates the one privileged operation at one
  audited site.
- **Hand the executor a non-`realm` capability (e.g. a `TreasurySender`
  backed by a pre-built RealmSend banker) instead of the `sub`** —
  considered as a by-construction seal: an executor with no `cross`-able
  `realm` value could not escalate at all. Rejected here because, to be
  un-defeatable, the capability's concrete type must live in a package
  the executor cannot type-assert into — which means moving fund movement
  into the `/p/` layer, against the treasury ADR's "the package never
  moves funds" invariant, or an awkward separate realm package. For a
  closed, audited executor set the precise documentation above is the
  proportionate fix; a VM-level "no-cross" execution mode for executors
  is the only complete seal and is out of scope for this PR.

## Consequences

- Package surface (breaking; quarantined realm + package, no live
  state): `ExecFunc` is now `func(int, realm) error`; `Execute` takes
  `(proposalID uint64, sub realm)` and calls `fn(0, sub)`; new optional
  `Funded` interface.
- Realm surface: the `Execute` wrapper computes the operative DAO and
  mints its sub; the three fund-movers rebuild their banker from the
  passed `sub` (their own `cur.Sub` removed) and declare `FundingDAO()`
  (spend/dissolve → `p.dao`, clawback → `p.target`); the non-fund
  executors ignore `sub`.
- Tests: `z_commondao_execute_0` rewritten for the non-crossing
  executor — the stale `unsafe.PreviousRealm()` cross-identity
  assertion is gone (an executor is non-crossing and must not read the
  frame for realm identity), replaced with checks that it ran and
  received a usable `sub`; the re-entrancy probe (`z_commondao_execute_3`)
  and the standard `_test.gno` fakes adopt the new signature. Both gno
  suites (package + realm) and all five commondao txtars
  pass; the txtars exercise the operative-sub path on a real node
  including the host≠operative clawback and dissolution sweeps.
