# Package-local `.txtar` integration tests

## Status

Proposed.

## Context

Every integration test script in the repo lives in one flat directory,
`gno.land/pkg/integration/testdata/` — 186 `.txtar` files at the time of
writing. They fall into two very different groups:

1. Scripts about the platform: the VM (`interrealm_*`, `alloc_*`, `restart*`),
   the node and keeper (`storage_deposit*`, `simulate_gas`), `gnokey`
   (`gnokey_*`), or a numbered regression (`issue_*`). These define their realms
   inline in the archive and are meaningless outside the Go test tree.
2. Scripts about one specific package in `examples/`: the six `commondao_*`
   files, `atomicswap`, `wugnot`, `ghverify`, `valopers`, `moul_authz`. These
   contain no inline Gno at all (or only a thin caller realm) — they `loadpkg`
   an existing example and drive it through `gnokey`.

For group 2 the flat layout has real costs. Someone changing
`examples/gno.land/r/gnoland/wugnot/wugnot.gno` has no way to know from the
directory that a `wugnot.txtar` two subtrees away pins its exact balances and
storage fees. Conversely, reviewing a realm PR means grepping the Go tree to find
out whether integration coverage exists. Unit tests (`_test.gno`) and filetests
(`filetests/`) already live in the package directory — the integration script was
the only piece of a package's test suite that did not.

`testscript.Params` supports exactly this: when `Dir` is empty, `Files` is used
as an explicit list of scripts, from anywhere on disk.

## Decision

A `.txtar` that is about one package lives in that package's directory. Anything
under `examples/` is discovered and run automatically, by the same test that runs
`testdata/`.

Concretely:

- **`gnovm/pkg/integration.FindTestScripts(roots ...string)`** walks each root
  and returns every `*.txtar` below it, skipping hidden directories. It returns
  an *error* on duplicate base names — across all roots, since they share one
  subtest namespace — because `testscript` names each subtest after the base name
  alone and would otherwise silently emit `render` and `render#1`. Scripts are
  therefore prefixed with the package name (`commondao_council.txtar`).
- **`NewTestingParamsFromRoots`** wraps it, so the `Dir`-must-be-empty-for-`Files`
  quirk of `testscript.Params` stays inside the package that owns the params. It
  requires **every** root to contribute at least one script. An aggregate
  "at least one script total" check is not enough once the roots are spread
  across the repository: `<GNOROOT>/examples` comes from an environment
  variable, and a `GNOROOT` pointing at an unrelated or stale checkout would
  contribute nothing while `testdata/`'s 175 scripts satisfied the total on their
  own — the suite would report `ok` having silently dropped every package-local
  script. Under the flat layout a wrong `GNOROOT` failed loudly instead, and that
  property is restored here rather than lost.
- **`TestTestdata`** now passes two roots: `testdata` and `<GNOROOT>/examples`.
  One test function, so `-run TestTestdata/wugnot` selects a script wherever it
  lives, `-run TestTestdata` still means "all integration tests", and the base
  name namespace is enforced once, globally.
- Eleven scripts move in this change:

  | script | new home |
  |---|---|
  | `commondao_{ancestor,council,dao_member,election,governance,treasury}.txtar` | `examples/gno.land/r/nt/commondao/v0/` |
  | `atomicswap.txtar` | `examples/gno.land/r/demo/defi/atomicswap/` |
  | `ghverify.txtar` | `examples/gno.land/r/gnoland/ghverify/` |
  | `wugnot.txtar` | `examples/gno.land/r/gnoland/wugnot/` |
  | `valopers.txtar` | `examples/gno.land/r/gnops/valopers/` |
  | `moul_authz.txtar` | `examples/gno.land/p/moul/authz/` |

- `.txtar` stays out of `goodFileXtns` in `gnovm/pkg/gnolang/mempackage.go`, so a
  script sitting next to a realm is *not* part of the mem-package and is never
  uploaded on-chain. That was already true (the extension was listed as a
  commented-out `XXX: to be considered`); this change makes it a load-bearing
  invariant, documents why, and pins both legs — `ReadMemPackage` dropping the
  file and `ValidateMemPackage` rejecting one that got in another way — with
  `TestMemPackage_TxtarIsNotPackageSource`.

Four consumers outside the Go test had to follow:

- `update_gas_wanted.sh` globbed `testdata/*.txtar`; it now snapshots and
  rewrites a `find`-generated list over both roots. It maps captured gas back to
  a file by base name, which is exactly the invariant `FindTestScripts`
  enforces — noted in the script so the coupling is visible. The `find` prunes
  dot-directories, matching the walker: otherwise a script under one would get
  its gas numbers rewritten by this tool and never be run to check the result.
- `gno.land/Makefile`'s failing-test summary grepped for `FAIL: testdata/`; the
  moved scripts print an absolute path, so the pattern is now `FAIL: .*\.txtar:`.
- **`gno fix`** took a `.txtar` only as an explicit file target. `make -C
  examples fix` and `gno fix ./...` pass directories, so a migration sweep would
  have rewritten a package's `.gno` files and walked past the script now beside
  them — leaving the `.gno` *inside* the archive on the old language version. A
  directory target now also covers the `.txtar` files in that directory, at the
  same (non-recursive) scope as its `.gno` files. Pinned by
  `testdata/fix/fix_dir_txtar.txtar`.
- **CI path filters.** A script under `examples/` matches the filters of
  `ci-dir-gnovm` and `ci-dir-examples` as well as `ci-dir-gnoland`, so a
  one-line gas-number refresh would have started paying for two suites that
  never read the file. Both now carry `'!examples/**/*.txtar'`. `ci-dir-gnoland`,
  which owns `TestTestdata`, still matches.

## Alternatives considered

**Leave everything in `testdata/`.** Zero risk, but keeps the discoverability
problem that motivated the change, and keeps growing: `params_valset_*` alone is
18 files that are really about `r/sys/validators/v3`.

**A separate `TestTestdataExamples` test function for the new root.** The first
shape of this change. Rejected on review: it splits the base-name namespace that
`update_gas_wanted.sh` keys on, so a collision between the two directories would
silently cross-rewrite two scripts' gas values; it makes `-run 'TestTestdata$'`
quietly cover only half the suite; and it duplicates per-call setup. Merging the
roots into one `testscript.Run` removes all three.

**A `tests/` or `testdata/` subdirectory inside each package.** Safer against
tooling that reads a package directory, but `ReadMemPackage` already ignores
unknown extensions, so the extra nesting buys nothing and reads worse than
`wugnot.gno` / `wugnot.txtar` side by side.

**Make `.txtar` a good file extension so scripts travel with the package.**
Rejected: it would upload test tooling on-chain and change the storage footprint
and deposit of every existing realm that gained a script.

**Allow duplicate base names and let `testscript` disambiguate with `#1`.**
Rejected: `-run TestTestdata/render` would then select an arbitrary one of them,
and which one depends on walk order.

**Also match `*.txt`, which `testscript`'s directory mode accepts.** Rejected. A
bare `.txt` next to real code is far more likely to be a fixture than a script,
and that argument is about `examples/` rather than `testdata/` — where the old
directory mode did glob `.txt`. The narrower rule was still kept, for both roots:
matching by root would mean the same file name runs or does not run depending on
where it sits, which is a worse trap than one uniform rule. No `.txt` script
exists in either root, so nothing is dropped today; `FindTestScripts` has a case
pinning that `.txt` is ignored, so the choice is visible rather than incidental,
and `docs/resources/gno-testing.md` states the extension.

## Consequences

- A package's full test suite — `_test.gno`, `filetests/`, `.txtar` — is in one
  place. Realm PRs surface their own integration coverage in the diff.
- Base names are globally unique, enforced by a hard failure rather than
  convention. The cost is a redundant-looking prefix for single-script packages.
- `gno.land/pkg/integration/testdata/` keeps a clearer charter: platform and
  regression scripts only.
- No behavioural change to the scripts themselves. They still run against
  in-memory nodes, in parallel, with `loadpkg` resolving through `GNOROOT`
  independently of where the script sits. `UPDATE_SCRIPTS` rewrites the file in
  place at its new location.
- Discovery now costs one walk of `examples/` (~10 ms for ~1900 files) per test
  process, against a multi-minute suite.
- Where scripts live is expressed in two places — `FindTestScripts`' roots in
  `testdata_test.go` and the `find` roots in `update_gas_wanted.sh`. Adding a
  third root means editing both; the script comment says so.
- Follow-up (not in this change): the remaining realm-owned families, notably
  `params_valset_*` → `r/sys/validators/v3` and `govdao_*` → `r/gov/dao/v3/impl`.
  Left out here to keep the diff reviewable; those also touch the node's valset
  bridge, so their ownership deserves its own discussion.
