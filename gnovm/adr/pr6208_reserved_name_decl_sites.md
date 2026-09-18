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
type-switch variable; then func, type and import names).

## Decision

**One name-based check at the one funnel every binding passes through.**
`initStaticBlocks2` enumerates every source binding (`:=`, `var`, `const`,
`type`, `func`, import, receiver, params, results, range key/value,
type-switch var) and reserves each through `StaticBlock.Reserve`. `Reserve`
now calls `checkDeclName`, which refuses a builtin name and any
`declReservedNames` entry (`misc.go`, a third tier beside Go keywords and
uverse names; today only `cur`). A
first parameter named `cur` is carved out by position; `checkCurParamType`,
where the function type is resolved, requires it to be realm-typed. One line
at the `SwitchStmt` covers a clause-less type switch. A `:=` in a block where
the name is already reserved is not re-reserved; the write rules refuse it as
a rebind. Function types and interface methods declare nothing, but the
`FuncTypeExpr` handler checks their parameter and result names the same way,
so the rule has no exception; it also hosts `checkCurParamType`, since every
function's type passes through it. The older builtin-only checks in
`predefineRecursively2` and `fillNameExprPath` stay; `Reserve` fires first.

**The write rules key on name plus resolved realm type**, after the
`DEFINE`/`ASSIGN` branch so a same-scope `cur, x :=` is refused. "A `cur` is the
parameter" implies the invariant they rely on; the cur-call provenance check is
belt-and-braces.

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
- On-chain: this is a consensus change that existing networks cannot restart
  into. Stdlib source is written to state once, at genesis; at every restart
  `PreprocessAllFilesAndSaveBlockNodes` re-preprocesses the stored packages
  with no per-package recover, so the stored `math` alone stops the node. The
  path is the genesis-replay hardfork of `pr5511`: halt, export, new genesis
  (which loads the new stdlibs), replay. Replay requires that every historical
  `AddPkg` still passes, so a scan is a precondition: preprocess every stored
  production package with this binary and list those refused, for the uverse
  block names plus `cur`; each must get a fix-up in the new genesis. Mainnet
  `gnoland-1` is such a network. Gating the rule per package (#5929) would
  avoid the fork but is not in tree.
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
