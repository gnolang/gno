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
