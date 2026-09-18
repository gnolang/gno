# ADR: reserved names are refused at every declaration site

## Status

Implemented on `fix-reserved-name-decl-sites`, stacked on #6196. PR number
pending. Closes #6181 part A.

## Context

Gno refuses to shadow a builtin (`len := 3` fails with "builtin identifiers
cannot be shadowed"), but the check ran at only two places: package-level
declarations (`predefineRecursively2`) and `var`/`:=` name expressions
(`fillNameExprPath`). Parameters, named results, receivers, type-switch
variables, range DEFINEs and for-init DEFINEs were never checked, so
`func f(len int)` and `for len := range xs` compiled (#6181).

#6196 made a realm-typed `cur` declarable only as a crossing function's first
parameter, with a type-based `checkRealmCurDecl` wired at exactly the sites
above. Its review proposed (not yet settled) reserving `cur` by *name*, like the builtins:
a `cur` in a realm reads as the frame's identity to every reader, and an int
named `cur`, or a shadow in an inner block, defeats that reading even where it
is safe. That is the same question #6181 asks about builtins, so the two are
solved with one check.

## Decision

**One name-based declaration check.** `checkDeclName(name)` refuses a builtin
name and, by name, `cur`. It runs at every binding site:

- receiver, params and results of a `FuncDecl` / `FuncLitExpr`, through
  `checkFuncDeclNames`, which skips the one `cur` a crossing function may
  bind: its first parameter, realm-typed;
- `var` and fresh `:=` (`defineOrDecl`), range DEFINEs (key and value, every
  element kind), for-init DEFINEs (under their `<name>.loopvar` spelling, which
  the check strips), and type-switch clause variables (all three `Define`
  sites).

Function *types* and interface methods are not bindings and are not checked;
the result-name check #6196 placed in the `FuncTypeExpr` handler moved to the
declaration sites for that reason.

**The write rules are unchanged.** They still key on name plus resolved realm
type; the invariant they rely on ("a realm-typed `cur` is the parameter") is
now implied by the stronger "a `cur` is the parameter".

## Alternatives considered

- **Keep `cur` type-based** and fix only the builtin sites. Not taken here,
  pending the #6196 discussion: it leaves `cur` locals and params of other types in realm
  code, which is the confusion the reservation exists to prevent, and it
  keeps two checks where one suffices.
- **Extend `fillNameExprPath` alone.** Rejected: params, results, receivers
  and type-switch variables never pass through it.
- **Reserve `cur` for struct fields too.** Not done: fields are not bindings,
  are always qualified (`s.cur`), and the builtin rule leaves them alone.
- **Fix the error span** (#6181 part B). Out of scope; the panic is still
  stamped with the enclosing node.

## Consequences

- Compatibility: `func f(len int)`, a result or receiver named after a
  builtin, `for len := range`, `switch len := x.(type)`, and any `cur` outside
  a crossing first parameter are now preprocess errors. In-tree renames:
  `math/modf.gno` (results `int` → `ip`), the `getRealm` native declarations
  in `chain/runtime/unsafe` (result `address` → `addr`; Go bindings are
  positional), `p/onbloc/diff` (`new` → `newStr`), `p/nt/bylaws` and
  `r/nt/commondao` (`cur` locals → `current`), `p/nt/seqid` test (`cur` →
  `id`), four quarantined files (`new`, `address`, `cur`), and three VM
  fixtures (`typeswitch1.gno` param `len`, `std_unsafe0.gno` and the shared
  `extern/.../tree/utils.gno` local `cur`).
- App hash: stdlib sources live in state, so renaming `math/modf.gno` and
  `unsafe.gno` moves the multistore root; `expectedCrossrealm38Hash` is
  re-pinned, as are the byte-size-dependent goldens in `restart_gas.txtar`
  (9 gas less per AddPkg) and `storage_deposit_price_change.txtar` (genesis
  cost). Verified by reverting only `preprocess.go`: every number is the same
  new value, so the check itself changes no state and charges no gas.
- On-chain: a deployed production package binding one of these names would
  fail at the next node restart, which re-preprocesses stored packages. The
  same deployed-code scan #6193 and #6196 called for applies before release.
- Fixtures: `shadow_builtin_{param,result,recv,range,forinit,typeswitch,closure}.gno`
  (one per site), `shadow_builtin_functype_legal.gno` (types are not
  bindings), `zrealm_cur_decl_int_{local,param}.gno`; the seven
  `zrealm_cur_decl_*` fixtures carry the new message; the non-realm `cur`
  positive controls in `zrealm_cur_legal.gno` and `zrealm_cur_other_legal.gno`
  are removed.
