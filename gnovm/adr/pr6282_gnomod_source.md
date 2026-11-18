# gnomod.toml `[source]`: linking a package to its repository

## Context

A deployed package carries its code but not its history, tests outside the
package, CI, issues or license context. Nothing records which repository a
`gno.land/r/x/y` came from. Cargo (`repository`), npm (`repository`) and Go
binaries (VCS stamping) all record it.

`gnomod.toml` is re-encoded through `gnomod.File` by the keeper on every
AddPackage (`stampGnomod`), so any field added to `gnomod.File` is part of the
bytes a node stores. A node that does not know the field drops it silently.

## Decision

An optional section:

```toml
[source]
  repository = "https://github.com/someone/contracts"
  path = "r/app"
  revision = "3cc494ec4d1f"   # stamped by publish tooling, never committed
```

- **Informational.** The chain stores it as written and never fetches or
  verifies it. gnoweb shows it as an outbound link, marked as unverified.
- **Fixed ASCII grammar** (`gnomod.Source.Validate`): `https://` + lowercase
  DNS host with at least one dot + optional port + one or more
  `[A-Za-z0-9._~-]` segments, no `.`/`..`; `path` is the same segments;
  `revision` is 7-64 lowercase hex. No `net/url`: its behaviour moves with the
  Go version and GODEBUG (`urlstrictcolons`, Go 1.26), and this runs in the
  keeper, so a later toolchain bump could refuse a replayed AddPackage. The
  grammar also closes percent-encoded dot segments that a browser resolves out
  of the repository, `?`/`#`, credentials, and non-ASCII homoglyphs or bidi
  controls in the displayed label.
- **Validated where a section is written, not where it is read.**
  `File.Validate` (hence `ParseBytes`) does not check `[source]`. It is checked
  in `checkGnomodConstraints` (AddPackage normal and inert paths, and
  EnablePackage) and by `gno mod init/tidy`. Read paths (gno test, gnodev,
  gnoweb, the keeper re-reading a stored package) ignore an invalid section,
  and `Source.URL` returns nothing for one, so a bad informational field never
  makes the whole file unreadable. MsgRun does not validate it either, so a
  MsgRun that succeeds on an older node still succeeds.
- **Rules may only be loosened.** Once a package with `[source]` is stored,
  tightening the grammar would refuse its replay. Starting strict keeps the
  only safe direction open.
- **`revision` is out of the approval hash.** `PackageContentHash` clears
  `Source.Revision` before hashing: publish tooling stamps it on the uploaded
  copy only, and an approver hashing a checkout with `enablepkg -pkgdir` never
  has it. `repository` and `path` are committed with the package and stay
  bound, so a re-park cannot swap the link under an approval. The stored copy
  keeps the revision.
- **Filled opt-in** by `gno mod init -source` / `gno mod tidy -source` from the
  git `origin` remote (`gitsource.Detect`). Credentials are always dropped, ssh
  aliases and drive letters (dotless hosts) are refused, and the whole section
  is replaced. `gno mod` never writes `revision`, and neither does gnokey: it
  signs what it is given.

## Alternatives considered

- **`net/url` validation.** Simpler, but version-dependent on a consensus path
  and too permissive (percent-encoded dot segments, non-ASCII).
- **Validate in `File.Validate`.** One call site, but a bad section made the
  whole gnomod unreadable everywhere (gnodev skipped the package, gnoweb lost
  every Package Info row), and MsgRun started refusing payloads an older node
  accepts.
- **Exclude all of `[source]` from the approval hash.** Would let a submitter
  change the displayed repository after approval with the same hash.
- **Keep `revision` bound and tell approvers to use `-pkg-hash`.** Breaks the
  `-pkgdir` flow for every stamped package for no security gain: a commit id is
  not code the chain runs.
- **Auto-detect on every `init`.** Rejected: a private repository's name must
  never reach the chain because someone ran `init`.

## Consequences

- **Consensus-breaking, needs a coordinated upgrade.** Nodes with and without
  the field store different bytes for a package declaring `[source]`, and the
  new AddPackage check refuses sections an older node accepted and dropped.
  Before an upgrade height, scan AddPackage payloads in the chain's tx export
  for a `source` key (none in the mainnet and testnet exports of 2026-09-28).
  MsgRun payloads need no scan: their behaviour is unchanged.
- `addpkg_gnomod_source.txtar` fails on a node without the field even though
  addpkg prints `OK!`, which is what guards the stored bytes.
- Links from today's tooling point at `/tree/HEAD/<path>` until a publish tool
  stamps `revision`. gnoweb only shows `@rev` when the link carries it.
- Deep links exist for github.com (exactly owner/repo) and gitlab.com (nested
  groups); any other host links to the repository root.
