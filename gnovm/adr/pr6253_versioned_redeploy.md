# PR6253: Versioned redeploy carries the realm's globals

Status: proof of concept, draft PR. Background: RFC #694, #2191, #4682.

## Context

A gno.land path is permanent. The one exception, a `private = true`
realm, may be redeployed by its creator, but the redeploy reruns `init`
over fresh globals and abandons the old object graph (#4949). So every
upgrade pattern in use (proxy + registry as in `r/gov/dao`, lazy copies,
clockworkgr/gno-upgradeable) moves state by copying it into a realm at a
new path, and pays for it twice: storage for both copies, and a new type
identity for every declared type, since `DeclaredTypeID` is pkgpath +
name.

The VM already reserves the machinery. `RunMemPackageOverRealm` hands a
redeploy the prior realm's ObjectID counter and deposit, and every
package-level variable is wrapped in a heap item so that, per
`docs/resources/gno-memory-model.md`, mutable realms could later
"swizzle" them.

## Decision

A `version` field in `gnomod.toml`, and an `[upgrade] authority` that
makes a public realm redeployable by one address. A redeploy that steps
the version by exactly one keeps the realm's state:

- `runMemPackage` reads the version itself: over a prior realm it plans
  the carry from the package still live in the store (every global by
  name, with its heap item's ObjectID and static TypeID; every declared
  type's underlying TypeID; the prior package block), then replaces the
  cached package value instead of refusing a cached one. Both keeper
  paths, AddPackage and EnablePackage, run it unchanged.
- The plan is applied after preprocessing and before any declaration
  runs. Each carried slot of the new package block is pointed at the
  prior heap item, whose declaration is then skipped. The prior block
  keeps its other slots but drops the carried ones, so the cleanup
  #4949 asks for cannot later count the live items as its own.
- Declared types are re-persisted with `Store.ReplaceType` on every
  redeploy, versioned or not, so objects loaded afterwards resolve to
  the new definition, methods included. `SetType` keeps whatever is
  cached, and the live package's types are cached by the time a
  redeploy saves, so the old redeploy served stale methods to any
  object it created.
- `migrate()` runs once in place of `init()`. `IsPkgInitFunc` gives it
  the same unreferenceable `migrate.N` suffix as `init`.
- Public realms: the authority alone may redeploy; absent, the path is
  permanent and cannot become upgradeable later; dropping it freezes
  the realm. Importers compile package selectors to block slots and
  method selectors to method indices, so a public redeploy may only
  append declarations and methods, checked against the prior layout: a
  stand-in for re-preprocessing importers, which a later layout-epoch
  design can lift. `init`, `migrate` and blank funcs are reserved after
  every other name, so adding or dropping one moves no slot. An
  immutable realm may not import an upgradeable one (type checker),
  per CONSTITUTION, Realm Upgrading.
- Node restart preprocesses a package after its imports, not in
  package-index order, which a redeploy of an imported realm breaks.
- Refused by the shared gnomod rules, on all three deploy paths, with
  the realm left as it was: a version that does not step by one,
  without `private` or an authority, or an authority without a version;
  an upgradeable realm turning private. Refused by the plan: a global
  retyped or turned into a non-variable; a declared type removed or
  given a different underlying type; for a public realm, anything
  moved or removed. A private realm may drop a global; its heap item
  is released.

Version unset on both sides is the old redeploy, unchanged.

## Alternatives

- New path per version (`/v2`), state copied: today's practice; cannot
  keep type identity or importers.
- `pkg@hash` or semver in the path (#694): the path changes, so the same
  copying follows. Auto-converting "convertible" type changes (thehowl):
  harder to make deterministic than refusing and a two-step migrate.

## Consequences

- Axioms 1 and 2 of the design note are demonstrated end to end
  (`redeploy_versioned_state.txtar`): same path, state carried, typed
  migration, new methods on old objects, survives restart.
- Not enforced: the Constitution's second rule, that an immutable realm
  may not persist values of an upgradeable realm's types (reachable
  through `any`). Enforcing it at finalize needs a mutability flag on
  the persisted package value, a proto change; reading gnomod.toml
  there costs a full package decode per foreign realm type and put an
  existing govdao txtar out of gas.
- Follow-ups: a realm-path authority (governance), a typed `path@N`
  import for reading removed or retyped globals from `migrate`, appended
  struct fields, gnoweb showing the upgradeable flag, and refusing a
  redeploy over persisted func-lit closures whose source moved.
- Consensus-breaking: reserving hidden funcs last changes package
  block slot order, so ObjectIDs, hashes and storage bytes shift (realm
  goldens, an apphash pin, a genesis balance line re-derived). Ships
  only with a chain upgrade.
- The prior blocks are not deleted; that leak predates this change
  (#4949). The carried objects are re-adopted, not re-charged.
- The inert path takes the same run and rules, untested by the txtars.
- The parser and go/types treat `migrate` as `init`; the transpiler and
  `gno fix` still special-case `init` by name.
