The steps of Gno 0.0 --> Gno 0.9 transpiling.
  1. `GetMemPackage()`, `ReadMemPackage()`, ....
  2. `ParseGnoMod()`: parse gno.mod (if any) and compare versions.
  3. `GoParse*()`: parse Gno to Go AST with go/parser.
  4. `Prepare*()`: minimal Go AST mutations for Gno VM compat.
  5. `m.PreprocessFiles()`: normal Gno AST preprocessing.
  6. `FindXItems*()`: Gno AST static analaysis to produce xitems.
  7. `Transpile*()` Part 1: re-key xitems by Go Node before step 2 line changes.
  8. `Transpile*()` Part 2: main Go AST mutations for Gno upgrade.
  9. `mpkg.WriteTo()`: write mem package to disk.

In `cmd/gno/tool_lint.go` each step is grouped into stages for all dirs:
  * Stage 1: (for all dirs)
    1. `gno.ReadMemPackage()`
    2. `gno.TypeCheckMemPackage()` > `ParseGnoMod()
    3. `gno.TypeCheckMemPackage()`  > `GoParseMemPackage()
       `gno.TypeCheckMemPackage()`  > `g.cfg.Check()
    4. `PrepareGno0p9()`
    5. `tm.PreprocessFiles()`
    6. `gno.FindXItemsGno0p9()`
  * Stage 2:
    7. `gno.TranspileGno0p9()` Part 1
    8. `gno.TranspileGno0p9()` Part 2
  * Stage 3:
    9. `mpkg.WriteTo()`

In gotypecheck.go, TypeCheck*() diverges at step 4 and terminates:
  1. `mpkg` provided as argument
  2. `ParseGnoMod()
  3. `GoParseMemPackage()
  4. `gimp.cfg.Check(): Go type-checker
