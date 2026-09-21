# Remove the legacy `cross1` migration sentinel

## Context

`cross1` was a uverse-defined sentinel introduced during the
`runtime.{Current,Previous}Realm` → `cur` migration so that
bare-`cross` call sites could be bulk-renamed (`cross` → `cross1`)
before each site was semantically threaded to the canonical
`cross(rlm)` form. It lowered to the same WithCross / Args[0]=nil AST
shape as the compiler-synthesized `.origin`, taking the
`callingCurOrOrigin` runtime path.

Policy (`interrealm_v2.md`): production code should not use `cross1`;
it was always slated for removal before testnets, betanet, and
mainnet. The migration is complete — nothing in `examples/`, `docs/`,
genesis fixtures, or integration testdata references it.

## Decision

Remove `cross1` entirely:

- uverse def (`uverse.go`).
- gno0p9 typechecker shim line (`gotypecheck.go`).
- first-arg recognizer (`nodes.go` `isLikeWithCross`).
- preprocess cases (name resolution and call-arg lowering).
- transpiler ident rewrite (`transpiler.go`).
- Filetest `zrealm_cross1_legacy.gno` → `zrealm_cross1_removed.gno`,
  now an error test asserting the name no longer resolves
  (typecheck: `undefined: cross1`; preprocess:
  `name cross1 not declared`).
- ADR docs updated: `migration_guide.md` §16 rewritten as the direct
  bare-`cross` → `cross(rlm)` recipe; `interrealm_v2.md` migration
  section marked as historical.

## Alternatives considered

Keep `cross1` as a long-tail compatibility shim (the position an
earlier revision of `migration_guide.md` §16 documented). Rejected:
nobody should be using it, and keeping a second user-reachable
spelling of the origin-crossing path enlarges the language surface
that upcoming realm-identity work must reason about.

## Consequences

- Restrictive gno0p9 type-check change: a package containing `cross1`
  no longer type-checks. Stored packages are re-preprocessed at node
  boot and historical txs re-type-check on replay, so any network
  meant to retain state or history should be grepped for stragglers
  before rolling out a binary with this change (the in-repo tree is
  verified clean).
- The dynamic-origin path (`callingCurOrOrigin` → `buildOriginRealm`)
  is now reachable only via compiler-synthesized `.origin`; the open
  question in `migration_guide.md`'s appendix about a user-facing
  primitive for the `SetRealm(NewUserRealm)` test pattern now has no
  interim workaround.
