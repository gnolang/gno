# Type declarations in nested blocks; decl-position-aware const checks

Decision 2 reached master first, in #6504 (see
`pr6504_preprocess_closure_scan_gas.md`), which moved the same three call
sites to `GetIsConstAt`. This PR now changes only `declareWith` (decision 1)
and adds regression tests for both.

## Context

Two preprocessor bugs made Go-legal block-level declarations panic or
miscompile:

1. **A `type` declaration inside any nested block panicked.** `declareWith`,
   which constructs the `*DeclaredType` for a non-alias type declaration,
   computed the type's `ParentLoc` (the disambiguator baked into local types'
   TypeIDs) from the declaring block node — and accepted only `*PackageNode`/
   `*FileNode` (blank location) and `*FuncDecl`/`*FuncLitExpr` (the function's
   location). Any other block panicked `expected type expr but got
   *gnolang.BlockStmt` (or `*ForStmt`, `*IfCaseStmt`, `*SwitchClauseStmt`, …):

   ```go
   func main() {
       {
           type t struct{ n int } // panic: expected type expr but got *gnolang.BlockStmt
           println(t{4}.n)
       }
   }
   ```

   Shadowing was not required; a type declaration anywhere but package level
   or directly in a function body was enough.

2. **A use or assignment before a shadowing `const` in the same block was
   const-checked against the shadow.** `initStaticBlocks` reserves the
   shadow's slot — and records its name in the block's `Consts` — before the
   body is preprocessed. The const checks were name-keyed
   (`GetIsConst(store, name)`), position-blind, and hit that entry even for a
   use that `fillNameExprPath` had already (correctly) resolved to the *outer*
   binding:

   ```go
   v := 1
   {
       println(v)       // printed 0 (const-folded against a value-less slot); Go prints 1
       const v = "c"
       println(v)       // "c", correct
   }
   ```

   The assignment flavor — `v = 2` before the `const v` — was falsely rejected
   with `cannot assign to const v`. `type` shadows were affected the same way
   (type names are const-marked), though they died earlier via bug 1.

## Decision

1. **`declareWith` accepts any block parent.** Package and file nodes keep a
   blank `ParentLoc` (canonical `pkgPath.name` TypeIDs); every other block —
   function bodies as before, and now `BlockStmt`, `ForStmt`, `IfCaseStmt`,
   `SwitchClauseStmt`, etc. — contributes its own location. Using the
   *declaring block's* location (rather than walking up to the enclosing
   function) keeps same-named types in sibling blocks of one function
   distinct:

   ```go
   { type t int; a = t(1) }
   { type t int; b = t(1) }
   println(a == b) // false, as in Go: distinct declared types
   ```

   For function-body declarations the parent is still the `*FuncDecl`/
   `*FuncLitExpr` itself, so existing TypeIDs are unchanged; nested-block
   declarations previously panicked, so no state persisted before this
   change contains the newly minted TypeIDs.

2. **The const checks become path-based** (landed in #6504). The three
   position-sensitive call sites — const-folding of a name use, const-folding
   of array/slice composite-lit keys (both in `preprocess.go`), and
   `assertValidAssignLhs` (`type_check.go`, shared by assignments and
   inc/dec) — call `GetIsConstAt(store, path)` on the NameExpr's
   already-resolved path instead of `GetIsConst(store, name)`.
   `fillNameExprPath` already resolves a use before the shadowing declaration
   to the outer binding (the shadow's slot is reserved but untyped until
   preprocessing reaches it), so the path names the right declaration; the
   name-keyed walk ignored it.

   `GetIsConst` itself is kept for the one caller with no path in hand (the
   cross-package selector check in `type_check.go`), where package-level decls
   are order-independent and position cannot matter.

   `IsAssignable`, which refused assignments to package-level funcs, was a
   second name-keyed check. Once the const check was path-based, it let
   `f = func() {...}` before a block-level `const f` or `type f` overwrite
   the package-level `f`, which a plain `f := 1` shadow already did. #6089
   replaced it with a check on the `NameSources` entry the path resolves to.
   `tests/files/assign46.gno` covers the `type f` spelling: only with
   decision 1 does that program get past the type declaration, so only
   #6089's check stops the write.

## Alternatives considered

**Walk up to the enclosing function for `ParentLoc`.** Matches the old intent
("local types belong to a function") but collides the TypeIDs of same-named
types declared in sibling blocks of one function, making values of two
distinct types compare equal through interfaces and confusing the type store.

**Make `Consts` position-aware (record the declaring statement's span).**
Fixes the const checks without touching call sites, but duplicates what the
resolved path already knows, and leaves every other name-keyed check with the
same trap; `IsAssignable` was one, fixed by #6089.

**Strip the shadow's `Consts` entry until its declaration is reached.**
Order-dependent mutation of block state during preprocessing; breaks
re-preprocessing idempotence for stdlib nodes.

## Consequences

- Type declarations now work in every block a statement can appear in, with
  Go-identical scoping and identity semantics (`tests/files/type43.gno`,
  `type44.gno`).
- Uses, assignments, and inc/dec before a shadowing `const` in the same block
  resolve and check against the outer name, as they have since #6504
  (`tests/files/const64.gno`, `const65.gno`, `const66.gno`).
- Local types declared in nested blocks get TypeIDs of the form
  `pkgPath[file:span].name` keyed to the declaring block. These are
  deterministic given the source, like the function-keyed ones before them.
- A value of a nested-block type can be stored in realm state but not read
  back: the storing transaction commits, and the next one touching the
  variable, to read or to overwrite it, aborts with `unexpected type with id
  <pkgPath>[<file>:<span>].<name>`. Only package-block types are `SetType`'d
  (#6063). Function-body types already fail the same way on master; this
  change makes the gap reachable from block-level declarations, which did not
  compile before. #6084 or #5935 closes it for both unchanged, since each
  keys on `ParentLoc != zero`, which every non-package parent now sets.
- This composes with the switch/if case-body shadowing fix (#6058): there,
  `GetIsConstAt` routes paths into the faux-copied region to the parent
  statement's block, so a use of the outer name before a case-body `const`
  shadow checks the parent's (non-const) declaration. The case-body variants
  of the tests here only become expressible once that PR lands, since case
  bodies reject shadowing declarations without it.

## Tests

- `tests/files/type43.gno` — type declarations in a plain nested block,
  `for`/`if`/`case` bodies, a block-level alias, and sibling-block identity
  (distinct across blocks, identical within one).
- `tests/files/type44.gno` — a block-level type shadowing an outer
  variable (with a use of the outer name before the declaration) and an outer
  type.
- `tests/files/type45.gno` — Go's `test/bigmap.go`: map key and value types
  declared in sibling blocks (#5662).
- `tests/files/const64.gno` — use, `++`, `+=` before a shadowing `const`.
- `tests/files/const65.gno` — a slice and an array literal key before a
  shadowing `:=`.
- `tests/files/const66.gno` — assigning to an outer `const` before a
  shadowing `:=` is still rejected.
- `tests/files/assign46.gno` — assigning to a package-level func before a
  block-level `type` shadow is rejected.

Each was run under real `go run` first; the `// Output:` blocks are Go's
actual output, and the `// TypeCheckError:` blocks match Go's error.
`const64`–`const66` fail against name-keyed `GetIsConst` call sites, and
`assign46` fails without #6089's func check.
