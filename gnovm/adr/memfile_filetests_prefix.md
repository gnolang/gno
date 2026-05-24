# ADR: MemFile.Name encodes the `filetests/` subdir

## Status

Proposed.

## Context

A `MemFile` is the wire-format unit of a Gno package's file contents
(`tm2/pkg/std/memfile.go`). Its `Name` field has historically been a
flat basename — no directory separators. The set of valid `Name`
strings is enforced by the `reFileName` regex, which carries a load-bearing
DO-NOT-MODIFY warning because it is wire-format-adjacent: amino-marshaled
MemPackages stored on chain were encoded against this regex, and any
relaxation accepts MemPackages that older nodes would reject.

Filetests (standalone `.gno` programs that the GnoVM runs as their own
"main" package) needed a way to live alongside their parent package
without colliding. Three iterations brought us here:

1. **Original** — filetests were `*_filetest.gno` files at the package
   root. The suffix was the sole identifier.

2. **PR #5669 (interrealm Phase 3, Jae Kwon)** — established the
   `filetests/` subdirectory convention on disk. Moved many existing
   filetest files into the new layout across `examples/`.

3. **PR #5704 (Morgan)** — wired `MemPackage.WriteTo` to route
   `*_filetest.gno` files into `filetests/` so writes followed the new
   layout. `MemFile.Name` remained a flat basename; the path was
   invented at write time from the suffix.

After (3), in-memory and on-disk representations were inconsistent: a
filetest lived at `dir/filetests/foo_filetest.gno` on disk but its
`MemFile.Name` was just `foo_filetest.gno`. Consumers that read from
disk and consumers that read from in-memory MemPackages had to disagree
about the file's identity, and the `filetests/` location was implicit
in the suffix rather than explicit in the data.

Two follow-on needs surfaced:

- Authors wanted to drop the `_filetest.gno` suffix entirely once the
  directory was the identifier (a `.gno` file in `filetests/` is
  unambiguously a filetest).
- Tools that round-trip MemPackages (e.g. `pkgdownload.Download`,
  `MemPackage.WriteTo`, golden-file sync) need a single source of
  truth for layout — duplicating the rule in every caller leads to
  bugs (and did, before this PR).

## Decision

`MemFile.Name` may carry an optional single-level `filetests/` prefix.
A `.gno` file with that prefix is a filetest. The legacy
`*_filetest.gno` suffix at the package root remains recognized for
backward compatibility with stored data.

Specifically:

- **Constants** (`tm2/pkg/std/memfile.go`):
  ```go
  const (
      FiletestsDir    = "filetests"
      FiletestsPrefix = "filetests/"
  )
  ```

- **Regex**: `reFileName` is split into `reBaseName` (root files, any
  allowed extension) and `reFiletest` (files under `filetests/`, `.gno`
  only). The full pattern is
  `^(filetests/<reFiletest>|<reBaseName>)$`.

- **Helper**: `IsFiletestName(name)` returns true iff
  - `name` starts with `filetests/` AND ends in `.gno`, or
  - `name` ends in `_filetest.gno` (legacy).

- **WriteTo** writes files at `dir/<MemFile.Name>` (not
  `dir/<basename>`). A legacy fallback prepends `filetests/` to any
  bare `*_filetest.gno` name so on-chain MemPackages still land at the
  expected on-disk location.

- **ReadMemPackage** scans both the package root and the `filetests/`
  subdir; any `.gno` file under `filetests/` is loaded with its
  `MemFile.Name` carrying the `filetests/` prefix.

- All call sites that previously branched on `_filetest.gno` suffix
  (FilterGno, ExcludeGno, GoParseMemPackage, IsTestFile, ParseMemPackageTests,
  doc/pkg, gnoweb view_source, gnodev loader, gnovm cmd/gno, etc.)
  now funnel through `std.IsFiletestName`.

## Consequences

### Wire-format compatibility

- **Old nodes receiving new MemPackages**: a node running pre-this-PR
  code that receives a MemPackage with a `filetests/foo.gno` `Name`
  will reject it as malformed during `ValidateBasic` (the old regex
  rejects `/`). MemPackages without filetests are unaffected.

- **New nodes receiving old MemPackages**: a new node receives a
  MemPackage with a bare `foo_filetest.gno` `Name` — `IsFiletestName`
  still classifies it as a filetest (legacy branch), and `WriteTo`'s
  fallback routes it to `filetests/`. Fully compatible.

- **Mixed-version network**: same-version peers must produce
  identical MemFile.Name sets for the same source. A network in mixed
  state during a rollout could see hash divergence on packages that
  include filetests. **Mitigation**: deploy in lockstep, or treat as a
  breaking upgrade (this PR is marked `!`).

### Hash stability

The `filetests/` prefix is part of `MemFile.Name`, which is included
in MemPackage hash inputs. Consequence:

- A package whose filetests were authored with the legacy bare
  suffix, then round-tripped through `WriteTo + ReadMemPackage` on a
  new node, comes back with `filetests/`-prefixed names. **The
  rewritten MemPackage has a different hash from the original.**

- Documented as a one-way migration in `WriteTo`'s doc comment.
  Callers that need hash stability (e.g. comparing on-chain hashes
  pre/post upgrade) must avoid the round-trip or accept the divergence.

### Stored MemPackage migration

Existing on-chain MemPackages keep working unchanged — the legacy
fallback in `WriteTo` and `IsFiletestName` handles them. No data
migration is required at the protocol level.

To get hash stability post-upgrade for a specific package, that package
must be redeployed with the new `MemFile.Name` layout (an explicit user
action, not an automatic rewrite).

### Sort order

`MemPackage.Sort()` orders by `Name`. With the prefix, a
`filetests/x.gno` file sorts BEFORE a `foo.gno` (because `f` < `f` then
`i` < `o`). Stored MemPackages with sorted files therefore have a
different ordering after migration. This affects:

- Stream-hashed digests that include order.
- Iteration order in consumers that assume the old shape.

### Reduced complexity at call sites

~10 functions across `mempackage.go`, `gotypecheck.go`, `cmd/gno`,
`pkgdownload`, `doc`, `gnoweb`, and `contribs/gnodev` previously
duplicated the `_filetest.gno` suffix check inline. They now funnel
through `std.IsFiletestName`. Adding new filetest naming conventions
in the future is a single-site change.

## Alternatives considered

1. **Keep the suffix as the only identifier**. Rejected because it
   couples in-memory layout to a string-suffix convention that the
   on-disk world has already diverged from (since PR #5669).

2. **Store filetests in a separate `FiletestFiles []*MemFile` field on
   `MemPackage`**. Cleaner type-level separation, but it's a much
   larger wire-format change (new field, amino schema migration, every
   consumer of `mpkg.Files` updated). Rejected as out of scope; could
   be a future cleanup once `FileRole` (below) lands.

3. **Compute `FileRole` (Prod | UnitTest | IntegrationTest | Filetest)
   once per file and have all callers branch on the enum**. This is
   the right long-term shape — it would collapse `FilterGno`,
   `ExcludeGno`, `GoParseMemPackage`, `IsTestFile`, and
   `IsFiletestName` into a single lookup table. Deferred to a separate
   ADR; this PR keeps the change narrow to the regex + Name encoding.

## Migration plan

1. **This PR (breaking, marked `!`)** — relax the regex, add the
   helper, update all call sites. Backward-compatible reads
   (legacy filetest data still loads). On-chain hashes for packages
   without filetests are unchanged.

2. **Downstream consumers** — third-party tools that parse
   `MemFile.Name` (RPC clients, indexers, gnopls) must accept the
   `filetests/` prefix. Provide release notes flagging the regex
   relaxation.

3. **Optional future cleanup** — once the legacy data is migrated
   (either via redeployments or a chain hard fork that rewrites stored
   names), the `WriteTo` legacy fallback and the `_filetest.gno`
   suffix branch in `IsFiletestName` can be removed. Not required for
   this PR.

## Testing

- Unit tests in `tm2/pkg/std/memfile_test.go` cover `IsFiletestName`,
  `ValidateBasic` accepting/rejecting the prefix, and `WriteTo`
  routing for new + legacy forms.
- Unit tests in `gnovm/pkg/gnolang/mempackage_test.go` cover
  `ValidateMemPackageAny` filetests-dir rules and the
  `ReadMemPackage → WriteTo → ReadMemPackage` round-trip.
- Tests in `gnovm/pkg/packages/filekind_test.go` and
  `gnovm/pkg/packages/pkgdownload/examplespkgfetcher/` cover the
  consumer pipelines.
- End-to-end coverage via the integration test suite
  (`gno.land/pkg/integration`) — `restart_nonval`, `addpkg_cla`,
  `realm_sync`, and `TestModApp/mod_*` exercise the full
  genesis/type-check/fetch pipeline with packages whose filetests now
  use the new layout (e.g. `examples/gno.land/p/nt/avl/v0/filetests/`).

## References

- Parent commit: `da7302ba5` (feat: allow MemFile.Name to encode
  filetests/ subdir).
- Predecessor PR #5669 (interrealm Phase 3) — first introduced the
  on-disk `filetests/` subdir.
- Predecessor PR #5704 — wired `WriteTo` to route filetests into the
  subdir while keeping `MemFile.Name` flat.
