# ADR: Graceful Package Preprocessing Recovery on Node Restart

## Context

When a gno.land node restarts, `VMKeeper.Initialize()` calls
`PreprocessAllFilesAndSaveBlockNodes()`, which iterates **every** persisted
`MemPackage` and re-preprocesses it. If any single package's preprocessing
panics (due to a VM update, a GnoVM bug, or incompatible code), the panic
propagates up and **crashes the entire node**.

This means one broken user-deployed package can prevent the node from starting
at all — even though all other packages (including stdlibs and critical system
realms) are perfectly fine.

## Decision

Add **per-package panic recovery** during restart preprocessing. If a package
fails, log the error, skip it, and continue with the remaining packages. The
node starts successfully with all healthy packages available.

### GnoVM: `PreprocessAllFilesAndSaveBlockNodes`

The method is refactored to process each package in a separate function with
`defer recover()`. On panic, the error is captured and the package path is
added to a returned `[]string` of failed paths. The loop continues with the
next package.

Return type changes from `void` to `[]string` (failed package paths).

### Keeper: `VMKeeper.Initialize`

Consumes the returned failed list and logs a warning for each failed package.
The preprocessing summary log now includes a `failures` count. Stdlib type
checking remains unchanged — stdlib failures are still fatal, since they
indicate a fundamental incompatibility.

### Gnoweb: broken package detection

When a realm query fails with a non-specific error (not "package not found" or
"no render declared"), gnoweb checks whether source files exist for the path
via `ListFiles()`. If source exists, the package is broken rather than missing,
and a dedicated "Package Unavailable" page is shown with a link to view source.

This approach requires no new error types, ABCI endpoints, or store interface
changes. It relies on the existing `vm/qfile` query which reads directly from
the persistent `MemPackage` store — independent of preprocessing.

### Broken package behavior at query time

A broken package may have partially-preprocessed block nodes in the in-memory
cache. Any attempt to evaluate code on it (e.g. `Render()`) will likely panic,
which is caught by the existing `doRecoverQuery` mechanism and returned as a
VM error to the client. Source viewing (`vm/qfile`) continues to work normally
since it reads from the persistent store.

## Alternatives Considered

- **Tracking broken packages in a store field (`map[string]error`):** Rejected
  for simplicity. The error is already logged, and downstream code handles
  failures naturally via existing `doRecoverQuery` panic recovery.
- **Storing nil in `cacheObjects` to make `GetPackage` return nil:** Rejected
  because transaction stores created via `BeginTransaction` have their own
  `cacheObjects` and would bypass the nil entry, loading the package from
  `baseStore` instead.
- **Adding a distinct `PackageBrokenError` at the keeper level:** Rejected
  as unnecessary complexity. The gnoweb source-file check achieves the same
  user-facing distinction without new error types.
- **Recovering at genesis/first-boot too:** Rejected. Genesis failures
  indicate real deployment errors that should be surfaced immediately.

## Consequences

- **Positive:** A single broken package no longer prevents the node from
  starting. The node remains operational with all healthy packages available.
- **Positive:** Operators get clear log warnings identifying which packages
  failed and why.
- **Positive:** Users see a clear "Package Unavailable" page on gnoweb with
  a link to view source code, rather than a generic error or a crash.
- **Positive:** Source viewing continues to work for broken packages, allowing
  developers to inspect the code.
- **Trade-off:** Partially-preprocessed block nodes may remain in the
  in-memory cache for broken packages. These are harmless — evaluation
  attempts will panic and be caught by `doRecoverQuery` — but they consume
  some memory. This is acceptable since the data is transient and rebuilt on
  each restart.
- **Trade-off:** The gnoweb broken-package check makes an additional
  `ListFiles` RPC call when a realm query fails. This only happens on error
  paths, so the performance impact is negligible.
