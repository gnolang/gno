# ADR: Package names must match the last element of the package path

PR: https://github.com/gnolang/gno/pull/5048
Issue: https://github.com/gnolang/gno/issues/1571

## Context

Until this change, a package could be deployed at any path regardless of its
declared package name: `package foo` could live at `gno.land/r/demo/bar`.
This diverges from the Go convention that the import path's last element and
the package name coincide, and it breaks tooling and readers that guess the
bound identifier of an import from its path (gnoweb, gnofmt, the doc server,
and the transpiler all did this guess with `path.Base`, incorrectly for
mismatched or versioned packages).

## Decision

- `gnolang.LastPathElement(path)` returns the last *meaningful* path element:
  the last element, or the one before it when the last element is a version
  suffix (`v0`, `v1`, `v2`, ..., matched by `^v(0|[1-9][0-9]*)$`).
- `gnolang.ValidatePkgNameMatchesPath(name, path)` errors when the declared
  package name differs from that element, and rejects paths ending in
  consecutive version suffixes (e.g. `gno.land/r/demo/foo/v2/v3`), for which
  no sensible expected name exists.
- The check is enforced in `ValidateMemPackageAny()` only for `MPUserAll`
  packages. This is exactly the shape used by on-chain deployment (VM keeper
  `AddPackage`, genesis loading) and by local tooling reading user packages.
  Other types are deliberately exempt:
  - `MPUserProd` as used by `MsgRun` (path `.../e/<addr>/run`, `package
    main`) would always mismatch.
  - `MP*Test`/`MPFiletests` packages have their own naming rules
    (`*_test`, arbitrary filetest names).
- `gno lint` reports mismatches with a dedicated `gnoPackageNameMismatchError`
  code, after the `ignore = true` module skip (ignored modules are opted out
  of lint processing entirely; deployment remains protected by
  `ValidateMemPackageAny`).
- The filetest runner derives the package name from the file body
  (`PackageNameFromFileBody`) instead of from the path, and validates it
  against the `PKGPATH` directive.

## Alternatives considered

- **Forbidding `/v1` like Go does** (raised by @jefft0: the Go module spec
  disallows `/v1` in import paths). Rejected: `v1` (and `v0`) suffixes are
  already widely used on-chain and in `examples/` (e.g.
  `gno.land/r/gnoland/boards2/v1`, `gno.land/p/nt/avl/v0`), and the common
  Gno pattern is a proxy realm at `pkg/` with actual versions at
  `pkg/{v1,v2,v3}` (@thehowl). So all of `v0`, `v1`, `v2`, ... are skipped.
- **Enforcing only in the VM keeper and linter** (the PR's first version).
  Moved into `ValidateMemPackageAny` conditioned on `MemPackageType`, per
  review discussion, so every consumer of user packages gets the same rule.
- **Expecting the package name of versioned paths to be the version suffix**
  (i.e. `package v2` at `.../foo/v2`). Rejected: matches neither Go
  convention nor existing Gno usage.

## Consequences

- **BREAKING**: new deployments whose package name does not match the last
  path element are rejected. Existing on-chain packages are unaffected: the
  check runs on the store's write path (`AddMemPackage`), never on reads, so
  grandfathered packages keep loading. Replaying historical transactions
  containing mismatched deployments against a new chain will fail, however.
- Path elements that cannot be valid package names become undeployable as
  the *last* element: names must match `[a-z][a-z0-9_]+`, so hyphenated
  (e.g. `gno.land/r/demo/my-realm`) or single-character last elements have
  no possible matching package. Hyphens remain fine in intermediate
  elements (namespaces). Documented in `docs/resources/gno-packages.md`.
- Paths ending in consecutive version suffixes are rejected outright.
- Several `examples/` packages were renamed to comply (`gnoblog` → `blog`,
  `gnopages` → `pages`, `foo20` → `grc20factory`, `eval` → `math_eval`,
  `todolistrealm` → `todolist`, `emitevents` → `events`, `image_embed` →
  `img_embed`, `tests` → `vm`, `mapdelete` → `map_delete`, ...). New
  networks therefore serve different package names at these paths than
  networks that deployed the old code; importers relying on the default
  bound identifier must use the new names (or an explicit alias).
- Some negative tests became unrepresentable and were converted or removed:
  filetests derive the package name from the body, so the old "expected
  package name [X] but got [Y]" error can no longer be triggered from a
  filetest (`zrealm_tests0/1.gno` became happy-path tests), and
  `addpkg_identifier_mismatch.txtar` was replaced by
  `addpkg_path_name_mismatch.txtar` since deploying a mismatched package is
  now rejected before an importer could reference it.
- gnoweb displays versioned packages as `name/vN` (`displayPackageName`);
  for grandfathered on-chain packages that violate the convention the
  displayed name is path-derived and may differ from the real package name.
