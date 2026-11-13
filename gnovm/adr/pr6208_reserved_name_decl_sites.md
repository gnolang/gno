# ADR: reserved names are refused at every declaration site

## Status

Proposed in #6208 as an alternative to #6196, whose fixes it carries.
Closes #6181 part A.

## Context

#6193 made the crossing `cur` parameter a fixed binding (three preprocess
write rules plus a runtime identity check). Review found holes: a same-scope
`cur, x :=` rebound it; the write rules refused a package-level `var cur
realm` and a named result `(cur realm)`; `t.Run` after `cross()` panicked in
the identity check. #6196 fixed these and made a realm-typed `cur` declarable
only as a crossing first parameter, with a type-based check wired per site.

Separately, Gno refuses to shadow a builtin (`len := 3` fails), but only at
package-level declarations and `var`/`:=`; parameters, results, receivers,
type-switch, range and for-init variables were never checked (#6181). The
rule exists for the reader: a reviewer reads `panic`, `cross`, `len` as VM
semantics, and Go lets any of them be rebound (`var panic = func(string) {}`
makes `panic("unauthorized")` return). #6196's review asked for `cur` by
*name* for the same reason, and its per-site check kept missing sites. #5981's
`iota`-only check in `Reserve` is subsumed (`iota` is a uverse name).

## Decision

**One name-based check at the deploy gate.** `TypeCheckMemPackage` is where
the chain admits source: `AddPackage`, `Run` and `EnablePackage` call it.
`go/types` hands back `Defs`, its record of every defining identifier, plus
`Implicits` for the type-switch guard and for an import without an alias
(`import ".../cur"` binds `cur` in file scope, so `cur.Previous()` would call
whatever that package defines) — a complete enumeration no future binding
form can slip past. `checkDeclNames` walks it and refuses a uverse name or a
contextual keyword (`contextualKeywords` in `misc.go`: reserved as a binding
name only, usable elsewhere, a third tier beside Go keywords and uverse
names; today only `cur`).

Fields, methods and labels are skipped: each has its own namespace, so
`struct{ bool bool }` and `func (T) len() int` shadow nothing and Go accepts
both. A `cur` is allowed only as a function type's first parameter whose type
is `realm`, read from `go/types` so an alias of `realm` is held to the same
rule; position alone would admit `func(x int, cur realm)`, type alone
`func(cur int)`. The refusal is reported in source order, because `Defs` is a
map. `prepareGoGno0p9` strips only a literal `cur realm` from `main` and
`init` before `go/types` runs; any other parameter stays to be judged, since
stripped, `func main(panic func(string))` would bind `panic` for the VM while
the checker still saw the builtin.

**The gate sees only source being admitted.** Not an import (already on
chain), not a stdlib (`LoadStdlib` type-checks the binary's own, and node
start re-checks the stored copy; neither is an admission, so stdlibs keep Go's
names and the stored stdlib source, part of the app hash, is unchanged), and
not a genesis replay: `TypeCheckOptions.AlreadyOnChain`, set from
`auth.IsGenesisReplay`, skips admission rules for the history a hardfork
re-delivers through `InitChain`. gnoland-1's history holds 33 packages the
gate refuses, `p/onbloc/diff/v0` in genesis itself; without the exemption the
upgraded binary cannot replay the chain it upgrades. Everything the
type-checker does to decide what a package MEANS still runs.

**Nothing in preprocess may tighten.** `PreprocessAllFilesAndSaveBlockNodes`
re-runs every stored package at node restart, so a rejection there refuses
source the chain already accepted — a validator fails to start on a package
it has served for months. Two cuts of this change tried: a name test on the
cur-call provenance check refused a stored `func F(_ realm, c realm) { cur :=
c; G(cur) }` at boot (dora, `348078dbe`;
`TestBootStoredCurThreadedFromBlankParam`), and the same-scope `cur, x :=`
rebind refused in preprocess would have stopped every upgraded node
(`TestBootStoredRealmTypedCurRebind`). `checkCurRebind` covers the rebind at
the gate (`:=` with one new variable ASSIGNS, so the ident is in `Uses`);
`doOpCall`'s identity check is what blocks a rebound `cur` at call time.
`preprocess.go` now differs from master by one const string and comments,
the invariant to hold this change to. The write rules keep #6196's resolved-type test, and the
older builtin-only checks in `Reserve`, `predefineRecursively2` and
`fillNameExprPath` stay: already enforced on chain, no stored package carries
what they refuse.

**Carried from #6196 unchanged:** the `testing` harness exemption
(`harnessSeedsCur`), `funcDisplayName`, and the unmetered frame walk in
`installInheritedCur` (rides with gnolang/gno-fixes#115).

## Alternatives considered

- **#6196's type-based check, per site.** Wired by hand and missed sites
  twice; `cur := 41` stays legal, so a reader must resolve a type.
- **Check in `Define`/`Define2`.** Too deep: uverse, faux-block copies and
  heap captures all define names there.
- **An activation height for the rule.** Keeps master's accept path in every
  later binary only so one binary can replay history; `VALIDATOR.md` already
  routes a release that cannot replay through the upgrade ledger.

## Consequences

- Compatibility: production source that binds a builtin name, or `cur`
  outside a crossing first parameter, is refused at `addpkg`, `run` and
  `enable`. Test files, filetests, `vm/qeval` and `gno run` are not checked;
  a package on chain keeps loading and running. In-tree: `p/onbloc/diff`
  `new` → `newStr`, `cur` locals in `p/nt/bylaws`, `r/nt/commondao`, four
  quarantined files, three VM fixtures. Stdlibs are untouched.
- On-chain: a coordinated upgrade, not "no fork". The gate changes which
  transactions the chain accepts, so the release carrying it is the halt's
  `halt_min_version`; and since the blocks that admitted the 33 packages do
  not replay on it, a node joining from genesis runs `v1.7.0` up to that halt
  (`misc/deployments/mainnet.gno.land/UPGRADES.md`). `RELEASING.md`'s rc
  rehearsal expects a fresh node to sync from genesis and needs a carve-out
  for such releases, as its own PR. A package parked before the halt with a
  refused name cannot be enabled after it. `gnogenesis fork` replays inside
  one `InitChain`, which is what `AlreadyOnChain` covers
  (`TestChainUpgradeGenesisReplay`, `TestCheckDeclNamesSkipsSourceAlreadyOnChain`).
- App hash: stdlib source is unchanged, so `expectedCrossrealm38Hash` and
  `restart_gas.txtar` stay at master's values; only
  `storage_deposit_price_change.txtar` moves, by the `p/onbloc/diff` bytes.
  The check charges no gas.
- Fixtures: one `shadow_builtin_*.gno` and one `zrealm_cur_decl_*.gno` per
  binding site, the implicit import included; the four
  `zrealm_cur_shadow*.gno` keep master's preprocess `Error:` with the gate's
  `TypeCheckError:` added, so a no-op provenance check still fails a fixture.
  `TestPreprocessAcceptsStoredReservedNames` pins what preprocess must accept.
- Docs: `go-gno-compatibility.md` and `gno-interrealm.md` state the rule and
  its scope. The runtime identity check stays: `zrealm_cur_backstop.gno` and
  `zrealm_cur_reassign_define.gno` reach it from source.
