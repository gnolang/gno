# ADR: Embed the stdlibs in `gnoland` and stop deriving `GNOROOT` from the build

## Status

Proposed. Implements #6240. No code accompanies this document yet.

## Context

### What `gnoland` reads from `GNOROOT`

The node locates a checkout of this repository through
`gnoenv.RootDir()` (`gnovm/pkg/gnoenv/gnoroot.go`) and reads three things
from it:

| What                                    | When                                                   | Where                                                     |
|-----------------------------------------|--------------------------------------------------------|-----------------------------------------------------------|
| `gnovm/stdlibs/**`                      | `InitChainer`, once, on a node whose data dir is empty | `gno.land/pkg/gnoland/app.go:60,326` → `vm/keeper.go:336` |
| `examples/**`                           | `gnoland start -lazy` when `genesis.json` is missing   | `gno.land/cmd/gnoland/start.go:451`                       |
| `gno.land/genesis/genesis_balances.txt` | default of `-genesis-balances-file`, same path         | `start.go:87`                                             |

The stdlibs are read with `gno.ReadMemPackage(dir, ..., MPStdlibAll)`,
test files included, and executed into the store during `InitChain`. They
are therefore part of the genesis state every validator must agree on.
Nothing reads them from disk afterwards: `PopulateStdlibCache` and every
later access go through the store, and a node restarting on existing
state opens no file under `GNOROOT` at all.

The examples are turned into `MsgAddPackage` transactions by
`gnoland.LoadPackagesFromDir`, the same function
`gnogenesis txs add packages` uses, and only when `-lazy` has to create a
`genesis.json`. Without `-lazy`, a missing genesis is an error
(`missing genesis.json`) and `examples/` is never opened. With an existing
genesis, `-lazy` only fills in missing config and secrets. The node code
does not require any example realm either: the three it knows about,
`r/sys/names`, `r/sys/cla` and `r/sys/validators/v0`, are each skipped
when not deployed (`keeper.go:504`, `keeper.go:545`, `app.go:1041`).

`RootDir()` resolves, in order: the `GNOROOT` environment variable, a
`_GNOROOT` value injected with `-ldflags -X`, the output of a
`go list -m -mod=mod github.com/gnolang/gno` subprocess, and finally the
absolute path of `gnoroot.go` recorded at compile time. That last step is
what a downloaded release binary ends up with. It only works when the
binary was built without `-trimpath` and runs on the machine that built
it; anywhere else it returns a directory that does not exist, without an
error. When every step fails, `RootDir()` panics. `start.go:86` calls it
while registering flags, which `tm2/pkg/commands` does at command
construction, so a `gnoland` that cannot find a root panics on every
subcommand, `version` included.

### What the mainnet validators actually needed

For mainnet the examples never went through the node.
`misc/deployments/mainnet.gno.land/gen-genesis.sh` filtered `examples/`
to a curated list (`gen-genesis.sh:108-123`), copied it to a staging
directory and handed it to `gnogenesis`; the resulting `genesis.json` was
published on the `chain/mainnet` release; validators downloaded it and ran
`gnoland start` without `-lazy`. On a validator, `GNOROOT` did exactly one
thing: provide `gnovm/stdlibs` to `InitChain` on the first start.

### How the two are shipped together today

- **Container images.** The root `Dockerfile` sets `ENV GNOROOT=/gnoroot`
  and, for five images (`gnoland`, `gnodev`, `gpao`, `gno` and the
  all-in-one), copies `examples/`, `gnovm/stdlibs/` and
  `gnovm/tests/stdlibs/` into it; `gnoland`, `gnodev` and the all-in-one
  also get the two genesis files (`Dockerfile:127-218`).
  `.github/goreleaser.yaml` lists the same trees as `extra_files`.
- **Release binaries.** `release-chain-tag.yml` builds the 20 binaries
  *without* `-trimpath`, on purpose, so that the compile-time path
  fallback survives (`release-chain-tag.yml:70-74`). That path is the CI
  runner's, so on the downloader's machine the node fails at `InitChain`
  with `failed loading stdlib "...": does not exist`.
  `misc/deployments/mainnet.gno.land/VALIDATOR.md:19-28` (and onyx's copy)
  therefore tells operators to `git clone --branch v1.5.0 --depth 1` next
  to the binary and `export GNOROOT=~/gno`.
- **Local builds.** `gno.land/Makefile:27` builds with `-trimpath` and no
  `_GNOROOT`, so an installed `gnoland` works only through the environment
  variable or the `go list` subprocess. `gnovm/Makefile:30` and
  `contribs/gnodev/Makefile:2` bake the absolute path of the checkout into
  `gno` and `gnodev` instead; the `Dockerfile` does the same with the
  constant `/gnoroot` for `gnodev`, `gnobro` and `gpao`, and goreleaser for
  `gnodev`. `gnoland` has never been built with `_GNOROOT`.

### The inconveniences this causes

1. **The operator path has a silent failure mode.** The checkout must be
   at the same tag as the binary, and nothing checks that it is. The
   native bindings are compiled into the binary while the stdlib sources
   come from the checkout, so a mismatch is either a panic at `InitChain`
   (#3361: `function GetChainDomain does not have a body but is not
   natively defined`, a `gnoland` built at one commit reading the stdlibs
   of another) or, worse, a genesis state that differs from the network's
   and surfaces as a consensus failure. The clone also pulls the whole
   repository to obtain about 2 MB of `.gno` files.
2. **Release builds are not reproducible.** #6177 tried to add `-trimpath`
   to the release workflow and had to revert it because a trimmed
   `gnoland` panics without `GNOROOT`. The checksums recorded in
   `misc/deployments/mainnet.gno.land/upgrades.json` attest a binary that
   contains the runner's file system paths.
3. **Docker got promoted over native binaries for the wrong reason.**
   `VALIDATOR.md` lists the container image's caveats (base image trust,
   registry as a single point of failure, a root daemon, a restart policy
   that double-signs). In the #6228 review the image was nonetheless the
   convenient path because it is "the same binary with `GNOROOT` preset";
   the reviewer's own conclusion was that embedding is "the real fix for
   why Docker got promoted".
4. **Locally built binaries depend on `go` at run time.** The `go list`
   step shells out, needs the toolchain on `PATH`, and `-mod=mod` may reach
   the network to resolve modules.
5. **Every tool built on the node re-derives `GNOROOT`.** The integration
   testscripts (`gno.land/pkg/integration/testscript_gnoland.go:119`),
   gnodev, gnoe2e and gpao each resolve it, export it to subprocesses, and
   document hygiene rules such as "`GNOROOT` must not be set in the
   environment" (`misc/gnoe2e/AGENTS.md:29`).
6. **`gnoland version` needs a resolvable root.** A `make install` build
   is trimmed, so run outside the checkout without `GNOROOT` it panics
   before printing anything.

Size is not on this list on purpose: the images copy the whole
`gnovm/stdlibs` directory today, and embedding moves the `.gno` part of it
into the binary rather than removing it (see Consequences).

### Why this was never embedded

The idea is not new. The history, in order:

- **#585 (2023-04, repository reorganisation)** introduced the layout in
  which `gnoland start` walks `examples/` to build its genesis and reads
  the stdlibs from the same tree. The node was a development tool that
  lived in the monorepo.
- **#1014 (2023-08)** added the `GNOROOT` environment variable and the
  `go list` guess, so that the node could run from outside the checkout
  without a `-root-dir` flag on every call. This is a heuristic layer on
  top of the file dependency, not a removal of it.
- **#1217 (2023-10, `gno env`)** proposed a production convention
  (`GNOROOT=/usr/local/gno`) and stated that `gnovm` and `tm2` "should not
  attempt to guess paths". Only `gno env` landed.
- **#1236 (2023-10)** reported that `-trimpath` made every binary panic
  inside amino, which derived a package directory from the caller's
  absolute source path and asserted it was absolute. Until #4854
  (2025-10-27) made that derivation tolerate a relative path, `-trimpath`
  was impossible regardless of `GNOROOT`; #1236 was closed in April 2026
  once it could no longer be reproduced. The `gnovm/Makefile:28` comment
  about amino dates from that period and is stale.
- **#1248 (2023-10, "`make install` should `go:embed` stdlibs, examples
  folder, and other dynamic dependencies")** is the direct ancestor of
  this ADR. The maintainer's answer agreed to embed the stdlibs ("a few
  megabytes"), added that `gnoland` would still need `GNOROOT` "for the
  examples directory", and expressed a preference for keeping the stdlibs
  "as inspectable files" in proper installations. Neither half was
  implemented. The issue went stale and was closed in November 2025.
- **#1952 (2024-04)** took the opposite route: move genesis generation out
  of `gnoland start` so that the node "should be able to run without
  needing a root directory". Part of it shipped as `gnogenesis` (#1988,
  moved to `contribs/` in #3041), but `-lazy` kept walking `examples/`,
  `start.go:52` still carries the `TODO: remove as part of #1952`, and the
  issue was closed in January 2026 with "never going to happen".
- **#3361 (2024-12)** is the mismatch failure observed in the wild, see
  inconvenience 1.
- **#6177 and #6228 (2026-09)** built the mainnet release process around
  the constraint instead of removing it: no `-trimpath`, a clone step in
  `VALIDATOR.md`, images preferred. **#6240 (2026-09-25)** asks for the
  stdlibs to be embedded so that `-trimpath` can come back.

Reading these together, the reasons it never happened are:

1. **The original design assumed the node runs inside the monorepo.**
   Embedding looked like a solution to a problem only production nodes
   have, and there was no mainnet until September 2026.
2. **The agreed fix was a different one, and it stalled.** The 2024 plan
   was to make `gnoland` not need the tree at all (#1952). When it died,
   the status quo stayed by default.
3. **The examples were bundled into the question and blocked it.** The
   2023 discussion settled on "embed stdlibs, `gnoland` keeps `GNOROOT`
   for examples". Since the node would need a checkout either way, the
   stdlib half lost its incentive. This ADR separates the two.
4. **Preference for inspectable files on disk**, and the fear of two
   sources of truth for developers editing the tree.
5. **Mechanics.** Every reader (`ReadMemPackage`, `loadStdlibPackage`,
   `ReadPkgListFromDir`) takes a directory path, not an `fs.FS`;
   `-trimpath` had the amino blocker; and the size was never measured.

## Decision

Make `gnoland` self-contained for everything a network node does, keep
`-lazy` reading `examples/` from a checkout, and remove the build-path
guess from `GNOROOT` resolution so that `-trimpath` can be used
everywhere.

1. **Embed the stdlibs, in a package only the node links.** `//go:embed`
   can only reach files below the embedding package's directory, so the
   obvious home is `gnovm/stdlibs` itself, which is already a Go package.
   It is the wrong one: every binary links `gnovm/stdlibs` for the native
   bindings, `gnokey` and `gnoweb` included, so all 20 release binaries
   would carry the sources. The embed therefore lives in a new package at
   `gnovm/` (the directory has no Go files today), and the only importer
   is `gno.land/pkg/gnoland`, which hands the FS to the keeper's loader as
   an argument. `gnokey`, `gnoweb` and `gno` do not link
   `gno.land/pkg/gnoland`, so only `gnoland` and `gnodev` grow. The
   patterns select `*.gno`, `*.md` and `gnomod.toml` at the three directory
   depths the tree uses (`//go:embed` has no `**`); the `.md` because four
   package directories hold a `README.md` that is part of the on-chain
   package today. The directory is not embedded as a whole: `gnovm/stdlibs`
   also holds 2.9 MB of Go source, the `time` zip data and the generated
   tables, that must stay out of the binary. What the FS
   yields is every file `ReadMemPackage` accepts with `MPStdlibAll`, test
   files included, so the packages written at `InitChain` are
   byte-identical to the ones read from a checkout at the same commit.
2. **`gnoland` loads the stdlibs only from that FS.** `InitChainerConfig`
   loses `StdlibDir` and gains an `fs.FS` that defaults to the embedded
   one; `InMemoryNodeConfig.StdlibDir` goes the same way. No override flag
   is added: it would reintroduce the mismatch this ADR removes, and a
   developer editing the stdlibs recompiles with `go run` anyway.
   `LoadStdlibCached`, keyed by directory today, becomes a single cached
   load.
3. **`fs.FS` readers.** `ReadMemPackage` and the keeper's stdlib loader
   get `fs.FS` variants; the path-based functions become thin wrappers
   over `os.DirFS`. `ReadPkgListFromDir` and `LoadPackagesFromDir` are
   left as they are, since `-lazy` keeps reading a directory.
4. **`-lazy` keeps relying on a checkout.** When `-lazy` is set and
   `genesis.json` is missing, the examples and the default balances file
   come from `-gnoroot-dir`, else `GNOROOT`, else the `go list` guess. When
   none resolves, the command fails with an error that names the flag and
   the variable, instead of panicking. `gnoenv.RootDir()` is no longer
   called while registering flags: the default of `-gnoroot-dir` is empty
   and resolution happens only at the moment a genesis has to be derived.
5. **`GNOROOT` resolution stops guessing from the build.** The
   `runtime.Caller` fallback in `gnoenv.guessRootDir` is deleted. It is the
   step that forbids `-trimpath`, and its failure mode is a wrong path
   returned without an error. The order becomes: `GNOROOT`, `_GNOROOT`,
   `go list`, then an error. The `_GNOROOT` ldflag is kept for now: it is
   an explicit build input, not a guess, it is compatible with `-trimpath`,
   the images set it to the constant `/gnoroot`, and `gno`, `gnodev`,
   `gnobro` and `gpao` still need a tree on disk. `gnoland` is never built
   with it, so nothing in this decision depends on it; it goes away with
   those tools when they embed their own trees (see Follow-ups).
6. **`-trimpath` everywhere.** `release-chain-tag.yml` (and its comment
   block explaining why not), `.github/goreleaser.yaml`, the `Dockerfile`
   builds and `gnovm/Makefile` (whose amino comment is stale) all build
   with `-trimpath`. `gno.land/Makefile` already does.

   This is about reproducible builds. Without `-trimpath`, Go records the
   absolute path of every source file in the binary, so the bytes depend
   on the directory the build ran in and two builds of the same tag
   differ. With it, a build is a function of the source at the tag alone:
   anyone can rebuild `v1.5.0` and compare byte for byte with
   `CHECKSUMS.txt` and the digests in `upgrades.json`, which is what makes
   those checksums a claim about the source rather than about the file.
7. **Images and docs.** The `gnoland` image stops copying
   `gnovm/stdlibs/` and `gnovm/tests/stdlibs/`; it keeps `examples/` and
   `ENV GNOROOT=/gnoroot` because `misc/e2e/docker-compose.yml` and the
   portal loop start it with `-lazy`. `VALIDATOR.md` loses the clone and
   `GNOROOT` steps; `install.md` and `misc/install.sh` follow.
8. **The release workflow checks the result.** Its existing step that runs
   the fresh binaries without `GNOROOT` is extended to a trimmed
   `gnoland start` against a provided genesis that reaches height 1.

Out of scope, deliberately: embedding `examples/` (below), the `gno` CLI,
gnodev, `gnogenesis`, and `gnovm/tests/stdlibs` (used by `gno test`, not
by the node).

## Alternatives considered

1. **Keep shipping the tree next to the binary, with better docs.** This
   is where #6177 and #6228 landed. It leaves the silent tag-mismatch
   failure, forbids `-trimpath`, and keeps the image as the path of least
   resistance for validators.
2. **Publish a `gnoroot-<tag>.tar.gz` release asset** that `gnoland`
   fetches and verifies on first start. Smaller than a clone, but a
   network dependency, a second artifact to sign, still an on-disk path,
   and still no `-trimpath`.
3. **Embed `examples/` as well** (#1248's original scope). Inside
   `gnoland` the examples are read by `-lazy` alone, to derive a genesis
   for a development chain, and everyone who runs `-lazy` has a checkout
   because they are developing against it. Embedding them would grow the
   node binaries for a feature validators never use. Deferred, not
   rejected: the `fs.FS` readers from this ADR are what it would need.
4. **Embed the examples behind a build tag**, in `make` builds only.
   `-lazy` would then behave differently between a local build and a
   release build, which is the kind of split the release process has just
   spent two PRs removing.
5. **Remove `_GNOROOT` now, together with the compile-path fallback.**
   Cleaner, and the user-visible half of "no build-time paths". But
   `make install` of `gno` and `gnodev` would then work outside the
   checkout only with `GNOROOT` set, since the `go list` guess needs a
   working directory inside a module that requires `github.com/gnolang/gno`
   and a Gno project has a `gnomod.toml`, not a `go.mod`. That is a
   regression for every `gno test` run on an external realm, for a
   decision that concerns `gnoland`. It belongs with the change that embeds
   the trees those tools need.
6. **A fixed install prefix baked with `_GNOROOT`** (#1217's model; what
   the `Dockerfile` does with `/gnoroot`). Works only for binaries an
   installer places together with the files, not for a download from a
   release page.
7. **Remove the dependency instead of embedding** (#1952): `gnoland` never
   generates a genesis and `gnogenesis` does. This does not touch the
   stdlibs, which are the operator-facing half, and drops `-lazy`, which
   every dev network relies on. Closed as "never going to happen".

## Consequences

### Size

| Tree            | Files embedded                        | Bytes that ship                                             |
|-----------------|---------------------------------------|-------------------------------------------------------------|
| `gnovm/stdlibs` | 254 `.gno`, 50 `gnomod.toml`, 4 `.md` | 2.3 MB (1.5 MB production, 0.9 MB tests), 0.6 MB compressed |

The directory itself is 5.2 MB (6.5 MB in `du` blocks) because it also
holds 175 Go files, 2.9 MB, of which `time/zzipdata.go` alone is 1.3 MB and
the generated `*.gno.gen.go` tables another 1.3 MB. None of that is
embedded; only the `.gno` sources, the `gnomod.toml` files and the four
package `README.md` are.

About 2.3 MB uncompressed on a `gnoland` that is 57 MB today. `gnodev`
grows by the same amount because it runs an in-memory `gnoland` node
(`gnoland.NewInMemoryNode`) and so links the package that imports the
embed; it gets the same guarantee in return. `gnokey`, `gnoweb` and `gno`
are unchanged. A release ships 5 tools for 4 platforms, 20 files; the two
affected tools are 8 of those files, about 18 MB per release in total.
The `gnoland` image stops copying the `gnovm/stdlibs` directory, but the
binary inside it grows by the same `.gno` bytes, so the image is not
meaningfully smaller and native downloads are bigger. This change is not
a size optimisation; its case is correctness and self-containment.

### Positive

- A downloaded `gnoland` runs against a published `genesis.json` after
  `chmod +x`. `VALIDATOR.md` drops from a clone-and-export procedure to
  download, checksum, run.
- The stdlibs a node writes at `InitChain` are the ones its native
  bindings were generated from. The #3361 panic and the silent app-hash
  divergence both become impossible for a node started from a release
  binary.
- `-trimpath` comes back on every build, so a release can be rebuilt from
  its tag and the checksums in `upgrades.json` attest something
  reproducible. This is the precondition for the signing and provenance
  work listed in the #6228 design note.
- Native binaries become the safe default and the container image loses
  its only advantage for validators.
- `gnoland version`, `config` and `secrets` no longer need a root, and a
  non-lazy `start` needs no `go` toolchain, no subprocess and no absolute
  path.
- gnodev's in-memory node and the integration testscripts get the stdlibs
  compiled into their own binary instead of reading them from `GNOROOT`.

### Negative and to watch

- **The reader refactor touches consensus-critical code.** The stdlib
  loading path feeds the genesis state, so the change follows the
  `AGENTS.md` verification rules for stdlib and execution-context changes
  (the examples suite, the gas tests, `TestTestdata`, the `Files` tests).
- **`-lazy` still needs a checkout.** That is today's behaviour; what
  changes is the failure: a clear error instead of a panic, and no more
  "works on the machine that built it". A dev image that runs `-lazy`
  keeps copying `examples/`.
- **The `go list` guess remains for development convenience**, with its
  subprocess and toolchain dependency. It is reached only by `-lazy`, and
  only when neither the flag nor the variable is set.
- **`gno` and `gnodev` stay checkout-bound.** Their Makefile and image
  builds keep `_GNOROOT`; their release binaries are built without it, so
  once the compile-path fallback is gone a downloaded `gno` fails with a
  clear error where it used to fail with a wrong path. Same outcome,
  better message, until they embed their own trees.
- **Stdlib churn changes the node's hash.** Since 2026-07-01, 10 of 165
  commits touched `gnovm/stdlibs`. That is the intended "one commit, one
  binary" property.

### Neutral

- The mainnet genesis path is unchanged: `gen-genesis.sh` keeps filtering
  the checkout into a staging directory and handing it to `gnogenesis`.
- The examples half stays where it is today, on disk, read by `-lazy`
  through `GNOROOT`, until someone picks up alternative 3.

## Testing to ship with the implementation

- A trimmed `gnoland` (`go build -trimpath`) with `GNOROOT` unset prints
  its version, starts against a provided `genesis.json` and reaches
  height 1. The release workflow's "assert the binaries carry the tag"
  step already runs the binaries without `GNOROOT` "on purpose" and is the
  natural home for this check.
- `gnoland start -lazy` with no genesis, no flag and no `GNOROOT`, run
  outside the repository: exits with the new error, does not panic.
- The `AGENTS.md` suites for stdlib and gas changes.

## Follow-ups (not decided here)

- Embed `examples/` for `-lazy` (alternative 3), if a checkout-free dev
  bootstrap turns out to be wanted.
- Embed `gnovm/stdlibs` and `gnovm/tests/stdlibs` in the `gno` CLI and
  gnodev, then delete `_GNOROOT` (alternative 5) and the `go list` guess.
- Sign the now-reproducible release binaries (#6228 design note,
  follow-up list).
- Delete the stale amino comment in `gnovm/Makefile:28` with the
  `-trimpath` change.

## AI assistance

Drafted with an AI agent from the repository, `git log`, and the linked
issues and pull requests; the size and churn numbers were measured on the
tree at the time of writing. The scoping to the stdlibs, the decision to
keep `-lazy` on a checkout and the removal of the compile-path guess are
the author's; the reading of the history is the agent's and is theirs to
confirm.
