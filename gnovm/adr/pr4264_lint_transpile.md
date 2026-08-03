The steps of Gno 0.0 --> Gno 0.9 transpiling.
  1. `GetMemPackage()`, `ReadMemPackage()`, ....
  2. `ParseGnoMod()`: parse gno.mod (if any) and compare versions.
  3. `GoParse*()`: parse Gno to Go AST with go/parser.
  4. `Prepare*()`: minimal Go AST mutations for Gno VM compat.
  5. `gno.MustParseFile()`: re-parse prepared AST
  6. `m.PreprocessFiles()`: normal Gno AST preprocessing.
  7. `FindXforms*()`: Gno AST static analaysis to produce line-based xforms.
  8. `Transpile*()` Part 1: re-key xforms from line-based to Go node-based.
  9. `Transpile*()` Part 2: main Go AST mutations for Gno upgrade.
  10. `mpkg.WriteTo()`: write mem package to disk.

In `cmd/gno/tool_fix.go` each step is grouped into three stages for all dirs:
  * Stage 1: (for all dirs)
    1. `gno.ReadMemPackage()`
    2. `gno.TypeCheckMemPackage()` > `ParseGnoMod()
    3. `gno.TypeCheckMemPackage()`  > `GoParseMemPackage()
       `gno.TypeCheckMemPackage()`  > `g.cfg.Check()
    4. `PrepareGno0p9()`
    5. `sourceAndTestFileset()` > `gno.MustParseFile()`
    6. `tm.PreprocessFiles()`
    7. `gno.FindXformsGno0p9()`
  * Stage 2:
    8. `gno.TranspileGno0p9()` Part 1
    9. `gno.TranspileGno0p9()` Part 2
  * Stage 3:
    10. `mpkg.WriteTo()`

In `cmd/gno/tool_lint.go` each step is grouped into two stages for all dirs,
and some steps are omited as compared to `tool_fix.go`:
  * Stage 1: (for all dirs)
    1. `gno.ReadMemPackage()`
    2. `gno.TypeCheckMemPackage()` > `ParseGnoMod()
    3. `gno.TypeCheckMemPackage()`  > `GoParseMemPackage()
       `gno.TypeCheckMemPackage()`  > `g.cfg.Check()
    4. `sourceAndTestFileset()` > `gno.MustParseFile()`
    5. `tm.PreprocessFiles()`
  * Stage 2:
    6. `mpkg.WriteTo()`

In `pkg/gnolang/gotypecheck.go`, `TypeCheck*()` diverges at step 4 and terminates:
  1. `mpkg` provided as argument
  2. `ParseGnoMod()
  3. `GoParseMemPackage()
  4. `gimp.cfg.Check(): Go type-checker

In `pkg/test/imports.go`, `_processMemPackage()` after loading when .PreprocessOnly:
  3. `GoParseMemPackage()`
  4. `PrepareGno0p9()`

## Open question: fatal vs. normal type-check errors

`TypeCheckMemPackage()` returns two kinds of errors that today are
indistinguishable to callers:

- **normal diagnostics** — ordinary `go/types` type errors ("constant overflows
  uint16"). The Gno preprocessor re-checks the same code, so the filetest
  harness deliberately runs both and pins both (`// TypeCheckError:` for
  `go/types`, `// Error:` for preprocess) to cross-check them.
- **fatal rejections** — the package uses an unsupported construct or trips a
  DoS guard (generics/type-sets via `checkNoGenerics`, type-expansion fan-out
  via `checkTypeExpansionBound`). Proceeding is meaningless: preprocess then
  emits an unrelated secondary error (e.g. `name P not defined` for a type
  parameter), so such filetests must pin two directives for no real benefit.

The deploy path already stops on *any* type-check error (`AddPackage` →
`ErrTypeCheck`), so this only affects the filetest harness. A future change could
tag fatal rejections as a distinct error kind and have `runFiletest` stop before
preprocess for them — but only as part of a deliberate definition of the
"unsupported Gno subset" (which errors are fatal: these guards? `go1.18`
version errors? all unsupported-feature rejections?), not a bolt-on. ~500
existing filetests pin both directives, so the split must be introduced
carefully.

## Decision: stdlib types are a bounded leaf in the type-expansion guard

`checkTypeExpansionBound` follows value-containment across imports to catch a
fan-out split over several packages. For imported **stdlib** types it stops:
`expansionPkgResolver` returns nil, so a stdlib `pkg.T` is counted as a leaf (1)
rather than resolved.

Why this is safe (no bounded-factor argument needed at the call site):

- The exponential DoS vector is value-containment **fan-out**, and fan-out lives
  in **user** types, which the guard counts in full. A user doubling chain over a
  stdlib type still explodes the user-side count and trips the budget.
- A stdlib type **cannot import user packages**, so its own expansion is fixed
  and small (measured max ~29 across all stdlibs) and cannot grow with input.
- Net: the leaf under-counts only by a bounded per-reference constant
  (`real <= K_max * counted`, `K_max` ~29), so a package that passes the budget
  has bounded real validType cost — it can never hide a fan-out.

Why not fetch/count stdlib source: `go/types` answers stdlib imports from its
result cache (permCache) **without a store read**, so fetching stdlib source in
the guard would add store gas the deploy otherwise never pays (the same class of
regression as double-fetching a dependency).

Exact stdlib counting is possible without that gas: precompute a
`stdlibPkgPath -> max expansion` table during `LoadStdlib` (deterministic, at
init, so no cold/warm gas skew) and look it up per reference. It was considered
and deferred: it is cross-module (gnovm type-check API + gno.land keeper) and
buys exactness for a leaf that is already bounded-safe, with no package flipping
accept/reject at the current counts. Revisit if stdlib expansion ever grows.

## Decision: dot imports are rejected before go/types, not counted

`checkTypeExpansionBound` resolves an unqualified type name in the **declaring
package** only (`namedCost`), and scores anything it cannot resolve — builtins,
type parameters, unresolvable qualifiers — as a leaf worth 1 node. A dot-imported
type is referenced by a bare identifier, so it lands in that leaf case while
`validType` expands it in full across the import boundary without memoizing. The
result is the cross-package hole again, reached by an unqualified name: a chain
written as `pkg.T` is rejected in microseconds, while the identical containment
shape written as `T` under `import . "pkg"` passes the budget and leaves go/types
churning inside `validType` for tens of seconds.

Gno never accepted dot imports — the preprocessor panics on them — but on the
deploy path preprocess runs *after* the type checker, so that rejection lands
only once the unmetered CPU has already been spent. `checkNoDotImports` therefore
rejects them in the same pre-type-check step as `checkNoGenerics`, for the same
reason and with the same placement argument.

Why reject rather than teach the cost model to count them:

- **Dot-import visibility is per file; the memo is per package.** `namedCost` is
  keyed `(package, name)`, which is sound exactly because an unqualified type
  name resolves in package scope. Two files of one package may dot-import
  different packages that each export `T`, so counting would need either a
  per-file memo key (weakening the memoization that makes this guard linear) or a
  package-wide max over every dot-imported package (an over-count, and a new
  false-rejection path on honest code).
- **The leaf cases stay provably safe.** "An unqualified type name resolves in
  the declaring package or is a bounded leaf" is the invariant that justifies
  every `return 1` in `cost()`/`namedCost`. Dot imports are the sole construct
  that breaks it; removing them keeps the invariant statable in one line instead
  of leaving a resolution fallback whose completeness must be re-derived on every
  future edit.
- **No blast radius.** No `.gno` file in the stdlibs, examples, or testdata uses a
  dot import (the sole occurrence, `gnovm/tests/backup/import2.gno`, is not
  referenced by any test), so nothing false-rejects.

Consequence: the guard's soundness now rests on two syntactic preconditions —
no generics/type-sets, no dot imports — both recorded in the `MAINTENANCE:` note
on `typeExpansionBudget`. If either rejection is ever relaxed, the corresponding
edge must be counted in `cost()` first.
