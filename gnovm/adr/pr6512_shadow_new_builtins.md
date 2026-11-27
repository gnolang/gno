# ADR: Let new builtins be shadowed

## Status

Proposed.

## Context

Since #1793 (issue #1779), no declaration may shadow a builtin identifier: a
package-level declaration, an import name, a local definition or a parameter
named `len`, `int` or `realm` stops preprocessing with
`builtin identifiers cannot be shadowed`. The reason was readability and
security: a reader assumes they know what a builtin name means.

The issue named the cost: every new builtin is a breaking change. Code that
already declares or imports the new name stops compiling, and on a running
chain that code is deployed and cannot change. Every node re-preprocesses every
stored package at boot, so such a package panics the node at startup.

#6387 adds `int256` and `uint256`. On `gnoland-1`,
`gno.land/p/onbloc/int256/v0` imports `gno.land/p/onbloc/uint256/v0` without an
alias, which names the import `uint256`: a binary with #6387 would refuse it at
boot. The core team agreed to lift the restriction for new builtins only.

## Decision

The builtins that exist today are listed in `protectedUverseNames`, and only
they cannot be shadowed. A builtin added later can be: a package-level
declaration, import name, local definition or parameter with its name takes
precedence in its scope, as Go's predeclared identifiers do.

- The list is frozen: nothing is ever added to it. A new builtin is
  shadowable by default, so adding one never needs to remember an opt-in.
- Names are resolved through the block chain before the uverse, so a
  shadowing declaration already wins in its scope; only the two checks that
  refuse the declaration change (`predefineRecursively2` for declarations and
  imports, and local definitions in the name-path pass).
- The uverse itself is exempt: it defines its native builtins by running
  their declarations through the same preprocessing (`DefineNative`), so
  `predefineRecursively2` skips the check for declarations into the uverse.

On master the behavior is unchanged: every builtin is in the list.

## Alternatives considered

- **Mark each new builtin as shadowable when it is defined.** Same effect
  for the builtins it marks, but every new builtin must remember the opt-in,
  and forgetting it is the breaking change this ADR removes.
- **Lift the restriction for every builtin.** Simpler, but reverses #1779's
  decision for the builtins it was made for, which the core team did not
  agree to.
- **Rename the new builtins** to names nothing uses. Every new builtin would
  need a search of every deployed package first, and the name space shrinks
  with each one.
- **Rewrite the deployed packages.** Not possible on a running chain without
  a genesis replay.

## Consequences

- A new builtin can be shadowed, so where a package shadows it, its name
  means something else in that scope. The readability concern of #1779
  applies to new builtins as it does to Go's.
- The VM is only half of it for a builtin that is also a type: the Go type
  checker sees Gno's builtins as declarations in a file injected into every
  package (`makeGnoBuiltins`), at package scope, where a file-scope import of
  the same name is a redeclaration. A new builtin type must be declared there
  so that an import can shadow it, which is #6387's to do.

## Tests

`TestShadowNewBuiltin` registers a builtin outside the protected list and
checks that an import alias, a package-level declaration, a local definition
and a parameter can use its name, that it still works where not shadowed, and
that `len` stays protected. `TestProtectedUverseNames` pins that every
protected name is a builtin. The existing `iota_identifier*` filetests pin the
protected cases.
