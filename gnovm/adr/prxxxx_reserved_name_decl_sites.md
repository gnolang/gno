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
above. Its review proposed, and has not yet settled, reserving `cur` by
*name*, like the builtins: a `cur` in a realm reads as the frame's identity to
every reader, and an int named `cur`, or a shadow in an inner block, defeats
that reading even where it is safe. That is the same question #6181 asks about
builtins, so this PR puts both on one check and offers it for that decision.

## Decision

**One name-based check, at the one funnel every binding passes through.**
`initStaticBlocks2` already enumerates every source binding — `:=`, `var`,
`const`, `type`, `func`, import, receiver, params, results, range key/value,
type-switch var — and reserves each through `StaticBlock.Reserve`. `Reserve`
now calls `checkDeclName(name)`, which refuses a builtin name and, by name,
`cur`. A first parameter named `cur` is carved out by position
(`NSFuncParam`, index 0); `checkCurParamType`, at the FuncDecl and FuncLitExpr
handlers where the function type is resolved, then requires that parameter to
be realm-typed. One extra line at the `SwitchStmt` covers a clause-less type
switch, whose variable no clause ever reserves.

A `:=` in a block where the name is already reserved (the crossing parameter's
own block) is not a declaration and is not re-reserved; the write rules refuse
it as a rebind, with their message. Function *types* and interface methods
are not bindings and are not checked; the result-name check #6196 placed in
the `FuncTypeExpr` handler is gone for that reason. The two older builtin-only
checks in `predefineRecursively2` and `fillNameExprPath` stay as they were;
`Reserve` fires before both.

**The write rules are unchanged.** They still key on name plus resolved realm
type; the invariant they rely on ("a realm-typed `cur` is the parameter") is
now implied by the stronger "a `cur` is the parameter".

## Alternatives considered

- **Keep `cur` type-based** and fix only the builtin sites. Not taken: it
  leaves `cur` locals and params of other types in realm code, which is the
  confusion the reservation exists to prevent, and keeps two checks.
- **Call the check at each declaration site in `preprocess1`.** The first cut
  did this at some fifteen sites and still missed func, type and import names.
  `Reserve` is the mechanism every binding already goes through.
- **Check in `Define`/`Define2`.** Too deep: uverse defines the builtins
  through `Define2`, faux blocks re-define already-checked names, and heap
  captures define `~name`.
- **Reserve `cur` for struct fields, methods, labels.** Not done: none is a
  block binding, all are qualified or in another namespace.
- **Fix the error span** (#6181 part B). Out of scope.

## Consequences

- Compatibility: any binding of a builtin name or of `cur` outside a crossing
  first parameter is now a preprocess error. In-tree renames: `math/modf.gno`
  (results `int` → `ip`), the `getRealm` native declarations in
  `chain/runtime/unsafe` (result `address` → `addr`; Go bindings are
  positional), `p/onbloc/diff` (`new` → `newStr`), `p/nt/bylaws` and
  `r/nt/commondao` (`cur` locals → `current`), `p/nt/seqid` test (`cur` →
  `id`), four quarantined files, and three VM fixtures.
- App hash: stdlib sources live in state, so renaming `math/modf.gno` and
  `unsafe.gno` moves the multistore root; `expectedCrossrealm38Hash` is
  re-pinned, as are the byte-size goldens in `restart_gas.txtar` and
  `storage_deposit_price_change.txtar`. Verified by reverting only the check:
  every number is the same new value, so the check charges no gas.
- On-chain: a deployed production package binding one of these names fails
  at the next node restart. `PreprocessAllFilesAndSaveBlockNodes` has no
  per-package recover, so the node does not start. The deployed-code scan
  #6193 and #6196 called for is therefore a precondition to releasing this,
  not a follow-up; the names to scan for are the uverse block names plus `cur`.
- Fixtures: one `shadow_builtin_*.gno` and one `zrealm_cur_decl_*.gno` per
  binding site, `shadow_builtin_functype_legal.gno` (types are not bindings);
  `zrealm_cur_alias*.gno` pin the resolved-type rule through a non-`cur` first
  parameter and a bare function type, since a second `cur` is refused by name.
- Docs: `go-gno-compatibility.md` and `gno-interrealm.md` state the rule.
