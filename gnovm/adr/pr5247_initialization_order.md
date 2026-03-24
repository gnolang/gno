# PR5247: Fix Variable Initialization Order Dependency Detection

## Context

Go spec requires package-level variables to be initialized in dependency order.
GnoVM's preprocessor computes this order by recording `ATTR_DECL_DEPS` on each
`*ValueDecl` and `*FuncDecl`, then running `findUnresolvedDeps` to sort declarations.

The old mechanism recorded deps inline in `preprocess1` via a `globalUse` closure
that called `addDependencyToTopDecl(ns, name)`. This worked for most cases, but
failed for variables initialized through immediately-invoked function expressions
(IIFEs) containing inner variable declarations:

```go
var A = func() int {
    var local = B  // <-- was wrongly attributed to ValueDecl{local}
    return local
}()
var B = 42
```

The root cause: `predefineRecursively` calls `Preprocess(ValueDecl)` with an
isolated, empty `ns = []`. When the inner `ValueDecl{local = B}` is processed,
`addDependencyToTopDecl([], "B")` finds no `*ValueDecl`/`*FuncDecl` ancestor in
the empty stack and records nothing, so A never learns it depends on B.

A secondary, pre-existing issue was that `preprocess1`'s SelectorExpr handling
called `evalStaticTypeOf(store, last, n.X)` to identify same-package method deps.
For realm packages, the store may contain imported packages as `RefValue`s rather
than `*PackageValue`s; evaluating the static type of cross-package selector
expressions triggered a panic in those cases.

## Decision

Replace the inline dep-recording in `preprocess1` with a dedicated post-preprocess
coda pass: `codaInitOrderDeps`.

### `codaInitOrderDeps`

Runs after `preprocess1` (all `NameExpr` paths are filled) and before
`codaPackageSelectors` (which replaces `NameExpr`s with `SelectorExpr`s for
cross-package references).

Uses `TranscribeB` to traverse the AST with a full ancestor node stack (`ns`)
always available.  For each node at `TRANS_LEAVE`:

**`*NameExpr`**: if the path type is `VPBlock`, the name is not blank, it is not
a package-name reference, it is not the LHS of a `var`/`const` declaration, and
the block that owns the name is the current package — add the name as a dep via
`addDependencyToTopDecl(ns, name)`.  Type-declaration names are excluded
(`NSTypeDecl`) because they carry no runtime initialization order.

**`*SelectorExpr`** with method path (`VPValMethod`, `VPPtrMethod`,
`VPDerefValMethod`, `VPDerefPtrMethod`): read the cached `ATTR_TYPEOF_VALUE` from
the receiver expression (unwrapping auto-generated `RefExpr` wrappers when
necessary), dereference pointer types, and if the result is a `*DeclaredType`
from the current package, add `"TypeName.MethodName"` as a dep.  This allows
`findUnresolvedDeps` to transitively discover variables referenced inside method
bodies.

`ATTR_TYPEOF_VALUE` is always present after `preprocess1` because
`evalStaticTypeOfRaw` caches the type on the node when it evaluates it.  Reading
the cached attribute is safe even for realm stores that contain `RefValue` package
references — no machine evaluation is triggered.

`FuncLitExpr` nodes marked `ATTR_PREPROCESS_SKIPPED` (IIFE bodies that contain
undefined names and are deferred to phase-2 preprocessing) are skipped entirely,
matching the guard used in the existing coda passes.

### Removed inline dep recording from `preprocess1`

The `globalUse` closure and all calls to `addDependencyToTopDecl` inside
`preprocess1` were removed.  `codaInitOrderDeps` supersedes them.

## Alternatives Considered

**Patch `predefineRecursively` to pass the outer `ns`**: would require threading
`ns` through many call sites and would not fix the analogous problem for other
isolated `Preprocess` calls.

**Record deps during `findUndefinedV`**: `findUndefinedV` already has access to
the full context but is a different traversal (it walks un-preprocessed ASTs);
mixing dep-recording there would conflate two concerns.

**Keep inline recording but fix the SelectorExpr panic separately**: would leave
the IIFE root cause in place.

## Consequences

- Variable initialization order is now correct for IIFEs with inner `var`
  declarations that reference outer package-level variables.
- The SelectorExpr panic for realm packages with `RefValue` imports is eliminated.
- The dep-recording logic is isolated in one clearly named pass with a comment
  explaining why it cannot be inline in `preprocess1`.
- `ATTR_TYPEOF_VALUE` must continue to be set on receiver expressions during
  `preprocess1` (currently guaranteed by the `evalStaticTypeOf` call at the
  SelectorExpr TRANS_LEAVE).

## Key Files

- `gnovm/pkg/gnolang/preprocess.go`: `codaInitOrderDeps`, `addDependencyToTopDecl`,
  `findUnresolvedDeps`, `resolveDeclDep`
- `gnovm/tests/files/var_initorder{9,10,11,14,17,19}.gno`: test cases
