# PR5247: Go-Compliant Variable Initialization Order

## Context

The Go specification mandates that package-level variables are initialized
stepwise, with each step selecting the variable earliest in declaration order
whose dependencies are all satisfied. GnoVM's original implementation used a
recursive depth-first `runDeclarationFor` function that computed dependencies
via `findDependentNames` (an AST walker) at initialization time. This had
several problems:

1. **Incorrect ordering.** The depth-first recursive approach did not implement
   the Go spec's "earliest in declaration order" rule. It processed variables
   in file-iteration order and recursively resolved deps depth-first, which
   could produce a different order than Go.

2. **Non-determinism.** Dependency sets were stored in `map[Name]struct{}`
   and iterated with `for dep := range deps`, making initialization order
   depend on Go's non-deterministic map iteration.

3. **Missed method dependencies.** `findDependentNames` relied on the
   `Externs` mechanism (`GetExternNames`) to discover names referenced from
   inside function bodies. However, method names were not tracked as externs,
   so their transitive dependencies were invisible. For example:

   ```go
   type T struct{}
   func (T) GetB() int { return B }
   var A = T{}.GetB() // dependency on B was not discovered
   var B = 42
   ```

4. **Shallow `Externs` tracking.** More broadly, `findDependentNames` depended
   on the `Externs` implementation on `StaticBlock`, which did not descend
   into function bodies. It only tracked names that crossed block boundaries
   during `GetPathForName`, not all names referenced within a function. This
   meant transitive dependencies through function calls could be missed.

## Decision

The fix has two parts: (A) how dependencies are syntactically recorded, and
(B) how the initialization order is computed from those dependencies.

### Part A: Syntactic Dependency Recording via `codaInitOrderDeps`

Replace the post-hoc `findDependentNames` AST walker with a single dedicated
coda pass: `codaInitOrderDeps`.

The pass runs after `preprocess1` (all `NameExpr` paths are filled) and before
`codaPackageSelectors` (which replaces `NameExpr`s with `SelectorExpr`s). It
uses `TranscribeB` to traverse the full AST of each file, with the ancestor
node stack always available.

For each `*NameExpr` at `TRANS_LEAVE`: if the path type is `VPBlock`, the name
is not blank, not a package reference, not the LHS of a declaration, and is
defined at package level (not a type declaration), the name is recorded via
`addDependencyToTopDecl(ns, name)` as an entry in `ATTR_DECL_DEPS` on the
nearest enclosing `*ValueDecl` or `*FuncDecl`.

For each `*SelectorExpr` with a method path (`VPValMethod`, `VPPtrMethod`,
`VPDerefValMethod`, `VPDerefPtrMethod`): the cached `ATTR_TYPEOF_VALUE` is
read from the receiver expression (unwrapping auto-generated `RefExpr`
wrappers), and if the result is a `*DeclaredType` from the current package,
`"TypeName.MethodName"` is recorded as a dep. This allows the resolution phase
to transitively discover variables referenced inside method bodies.

### Part B: Initialization Order via Memoized DFS + Kahn's Algorithm

**`resolveEffectiveDeps`** (memoized DFS, O(V+E)): For every declaration
reachable from the pending list, computes the set of `*ValueDecl` dependencies
by collapsing `FuncDecl` edges. FuncDecls are transparent pass-throughs: their
effective `*ValueDecl` deps are inherited by callers. Each `Decl` is visited at
most once thanks to a shared `cache` map, so total work is O(V+E) regardless
of the number of declarations. Circular variable dependencies are detected
during this DFS (via an `onStack` set) and produce a panic with the full
dependency chain.

**[Kahn's topological sort][kahn]** (in `runFileDecls`): Builds a
reverse-dependency index and unsatisfied-count array from the effective deps.
A min-heap keyed on declaration index ensures the Go spec's "earliest in
declaration order" tiebreaking. Each declaration enters and leaves the heap at
most once, giving O(V + E + V log V) total.

[kahn]: https://en.wikipedia.org/wiki/Topological_sorting#Kahn's_algorithm

### Removed Code

- `findDependentNames`: recursive AST walker that needed a case for every node
  type, replaced by `codaInitOrderDeps` (which uses the existing `TranscribeB`
  infrastructure to walk the full AST generically).
- `GetExternNames` / `addExternName` / `isFile`: the `Externs` tracking on
  `StaticBlock` was only used by `findDependentNames` via `FuncLitExpr` and
  `FuncDecl`. The `Externs` field is retained for amino serialization
  backward-compatibility but is no longer populated.
- `runDeclarationFor` / `loopfindr`: the recursive initialization loop in
  `runFileDecls`, replaced by Kahn's algorithm.

## Alternatives Considered

### Iteration 1: Rewrite `findDependentNames` + stepwise loop

The first approach kept the `findDependentNames` AST walker but rewrote
`runFileDecls` to use a stepwise "find earliest ready" loop instead of
recursive DFS. This fixed the ordering but kept the incomplete walker and did
not address the method dependency problem or the non-determinism from map
iteration.

### Iteration 2: Full rewrite with inline dep recording

Rewrote `findDependentNames` from scratch to work on the preprocessed AST
(using filled `NameExpr` paths instead of raw names) and moved dependency
recording inline into `preprocess1`. This fixed many issues but failed to
cover dependencies through methods: the `SelectorExpr` handling required
knowing the receiver's declared type to look up the method's `FuncDecl`, but
the inline approach did not reliably have this information for all receiver
patterns (value vs pointer, auto-addressed, etc.).

### Iteration 3: `codaInitOrderDeps` + per-decl `findUnresolvedDeps`

Moved dep recording to a dedicated post-preprocess coda pass
(`codaInitOrderDeps`) using `TranscribeB`, fixing the method dependency issue
by reading `ATTR_TYPEOF_VALUE` from the already-preprocessed receiver
expression. Used `findUnresolvedDeps` (a per-declaration DFS) to resolve
transitive deps and a stepwise scanning loop to find the earliest ready
variable. This was correct but O(n²): `findUnresolvedDeps` was called
independently for each declaration with no shared state, re-traversing the
same `FuncDecl` bodies repeatedly, and the ready-variable loop scanned all
pending entries each iteration.

### Other alternatives not pursued

**Patch `predefineRecursively` to pass the outer `ns`**: would require
threading `ns` through many call sites and would not fix the analogous problem
for other isolated `Preprocess` calls.

**Record deps during `findUndefinedV`**: already has access to the full
context but walks un-preprocessed ASTs; mixing dep-recording there would
conflate two concerns.

## Consequences

- Variable initialization order is now correct and deterministic, matching
  the Go specification's stepwise algorithm.
- O(V+E) dependency resolution (memoized DFS) and O(V + E + V log V)
  initialization (Kahn's with min-heap) replace O(n²) algorithms. Packages
  with thousands of top-level declarations are no longer a performance concern.
- `ATTR_TYPEOF_VALUE` must continue to be set on receiver expressions during
  `preprocess1` (currently guaranteed by the `evalStaticTypeOf` call at the
  SelectorExpr `TRANS_LEAVE`).
- Interface dispatch (`Getter(T{}).GetB()`) does not trace into the concrete
  method body, since the static receiver type is `*InterfaceType`. This is
  spec-compliant but diverges from gc's behavior in simple cases.

## Key Files

- `gnovm/pkg/gnolang/preprocess.go`: `codaInitOrderDeps`,
  `addDependencyToTopDecl`, `resolveEffectiveDeps`, `resolveDeclDep`
- `gnovm/pkg/gnolang/machine.go`: `initHeap`, Kahn's loop in `runFileDecls`
- `gnovm/tests/files/var_initorder*.gno`: 19 filetests
- `gnovm/pkg/gnolang/preprocess_test.go`: `TestInitOrderDeterminism`,
  `TestCircDepDeterminism`
