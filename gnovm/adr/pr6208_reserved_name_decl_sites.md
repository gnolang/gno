# ADR: reserved names are refused at every declaration site

## Status

Proposed in #6208 as an alternative to #6196, whose fixes it carries.
Closes #6181 part A.

## Context

#6193 made the crossing `cur` parameter a fixed binding (three preprocess
write rules plus a runtime identity check). Review found: a same-scope
`cur, x :=` rebound it; the write rules refused a package-level `var cur
realm` and a named result `(cur realm)`, and `func F(_ realm) (cur realm)` let
a result pose as the binding; `t.Run` after `cross()` panicked in the identity
check; the panic named nothing for a func literal. #6196 fixed these and made
a realm-typed `cur` declarable only as a crossing first parameter, with a
type-based `checkRealmCurDecl` wired per site.

Separately, Gno refuses to shadow a builtin (`len := 3` fails), but only at
package-level declarations and `var`/`:=`. Parameters, results, receivers,
type-switch variables, range and for-init DEFINEs were never checked (#6181).
The rule exists for the reader, not the VM: in contract code a reviewer reads
`panic`, `cross`, `revive`, `attach`, `len` as VM semantics, and Go lets any
of them be rebound (`var panic = func(string) {}`; `panic("unauthorized")`
then returns). Every unchecked site was a way to spoof a builtin in a body.
#6196's review asked whether `cur` should be reserved by *name* for the same
reason from the other side: a reader must be able to trust that `cur` is the
frame's identity, and an int named `cur` defeats that even where it is safe.
Both are one question, and #6196's per-site check kept missing sites (the
type-switch variable; then func, type and import names). #5981 meanwhile
landed an `iota`-only check in `Reserve`, at the same funnel; the general
check here subsumes it (`iota` is a uverse name) and its fixtures stay.

## Decision

**One name-based check at the deploy gate.** `TypeCheckMemPackage` is the
only place the chain admits new source: `AddPackage` and `Run` call it, and
boot calls it for the binary's own stdlibs. `go/types` hands back `Defs`, its
record of every defining identifier, plus `Implicits` for the type-switch
guard — a complete enumeration that no future binding form can slip past, the
same property `Reserve` was chosen for. `checkDeclNames` walks it and refuses
a uverse name or a contextual keyword (`contextualKeywords` in `misc.go`:
reserved as a binding name only, usable elsewhere, a third tier beside Go
keywords and uverse names; today only `cur`).

Fields, methods and labels are skipped: each has its own namespace, so
`struct{ bool bool }` and `func (T) len() int` shadow nothing and Go accepts
both. A `cur` is allowed only as a function type's first parameter whose type
is `realm`, read from `go/types` so an alias of `realm` is held to the same
rule; position alone would admit `func(x int, cur realm)`, type alone
`func(cur int)`. The refusal is reported in source order, because `Defs` is a
map.

**The rule may not live in preprocess.** `PreprocessAllFilesAndSaveBlockNodes`
re-preprocesses every stored package at every node restart, so a rejection
there refuses source the chain already accepted — a validator fails to start
on a package it has served for months. The older builtin-only checks in
`Reserve`, `predefineRecursively2` and `fillNameExprPath` stay as they are:
they are already enforced on chain, so no stored package carries what they
refuse.

**The write rules keep #6196's resolved-type test** (assignment,
address-of, range target), since stored code may bind `cur` to anything; the
same-scope `cur, x :=` form is refused by running the check after the
`DEFINE`/`ASSIGN` branch, which is #6196's bug fixed. The cur-call provenance
check is belt-and-braces.

**Carried from #6196 unchanged.** The identity check exempts calls dispatched
by package `testing` (`harnessSeedsCur`): the harness seeds a crossing sub-test
with the top-level test's `cur` on purpose, and the package is unreachable from
chain code. `funcDisplayName` falls back to the source location, then `<func
literal>`. The `curUsesPreprocessOrigin` header states the pointer-identity
semantics. The unmetered frame walk in `installInheritedCur` is left as is: a
gas change is consensus-visible and rides with the schedule work
(gnolang/gno-fixes#115, `NOTE` at the call site).

## Alternatives considered

- **#6196's type-based check, per site.** Enough for the write rules, but it
  runs where each type resolves, so it is wired by hand and missed sites
  twice; and `cur := 41` or `var cur any = rlm` stay legal, so a reader must
  resolve a type to know what `cur` is.
- **Check in `Define`/`Define2`.** Too deep: uverse, faux-block copies and
  heap captures (`~name`) all define names there.
- **Leave function types and interface methods alone.** Rejected: "no
  exception" is easier to learn, and the names still print in types and docs.
- **Struct fields, methods, labels.** Not bindings; left alone.
- **Keep the `DEFINE` carve-out in the assignment rule.** "The name is new" is
  false for a same-scope `:=`.
- **Fix the error span** (#6181 part B). Out of scope.

## Consequences

- Compatibility: any binding of a builtin name, or of `cur` outside a crossing
  first parameter, is a preprocess error. In-tree: `math/modf.gno` results
  `int, frac` → `integer, fractional`, upstream Go's current names;
  `chain/runtime/unsafe` `getRealm` result `address` → `addr`
  (Go bindings are positional); `p/onbloc/diff` `new` → `newStr`; `cur` locals
  in `p/nt/bylaws`, `r/nt/commondao`, `p/nt/seqid` test, four quarantined
  files, three VM fixtures; the `var cur realm` nil placeholders in 28 examples
  test files renamed to `rlm` or deleted.
- App hash: stdlib sources live in state, so the renames move the multistore
  root; `expectedCrossrealm38Hash` and the byte-size goldens in
  `restart_gas.txtar` and `storage_deposit_price_change.txtar` are re-pinned.
  Reverting only the check leaves every number at the new value: the check
  charges no gas. No gas schedule change.
- On-chain: no fork, and it took two things rather than one. The gate never
  re-reads stored USER source, so a package already deployed with a
  builtin-named parameter keeps loading and keeps running; only its next
  deploy is refused. Stdlib source is not like that: it is written to state
  at genesis, and node start re-type-checks the STORED copy
  (`VMKeeper.Initialize` reads `vm.gnoStore.GetMemPackage(stdlib)`), never
  the binary's. The rule alone therefore stops a restarting node on the
  genesis-era `math/modf.gno` — measured by omarsy on a dev chain, and the
  reason the first cut of this ADR was wrong to claim otherwise.
  `TypeCheckOptions.AlreadyOnChain`, set at that one call site, skips
  admission rules for source the chain has already admitted; what the
  type-checker does to decide what a package MEANS still runs. So the node
  starts, and the renames below only have to hold for chains made with this
  binary. The genesis-replay hardfork an in-preprocess rule would have
  required of every live network, `gnoland-1` included, is not needed.
  `TestCheckDeclNamesSkipsSourceAlreadyOnChain`,
  `TestPreprocessAcceptsStoredReservedNames` and
  `TestBootPreprocessesStoredReservedNames` pin those paths.
- Fixtures: one `shadow_builtin_*.gno` and one `zrealm_cur_decl_*.gno` per
  binding site, plus `shadow_builtin_{functype,iface_method}.gno` and
  `zrealm_cur_decl_{functype,functype_second,functype_result,iface_method}.gno`
  for types, with the crossing shapes `func(cur realm)` and `M(cur realm)`
  kept legal in `zrealm_cur_legal.gno`; `zrealm_cur_alias*.gno`
  pin the resolved-type rule through a non-`cur` first parameter and a bare
  function type; the four `zrealm_cur_shadow*.gno` are deleted (the shadow is
  now a declaration error); `zrealm_cur_reassign_define.gno`,
  `zrealm_cur_closure.gno`, `cur_subtest_test.gno`, `op_call_test.go`.
- Docs: `go-gno-compatibility.md` and `gno-interrealm.md` state the rule.
- The runtime identity check has no known reachable source route and is kept.
