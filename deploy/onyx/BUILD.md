# Building the two binaries

Both are static `linux/amd64` ELF binaries, `CGO_ENABLED=0`, `-trimpath`. Both
builds are byte reproducible: building `gnoland` twice from the same tag with
the same Go toolchain produced the identical sha256 (verified on lyra,
2026-11-23).

Toolchain used for the checksums below: **go1.25.9 linux/amd64**. A different Go
minor version produces a different, equally valid binary with a different
sha256.

## Versions, and why these ones

Onyx runs mainnet's code as a release candidate. The ladder lives in
[`misc/deployments/onyx.gno.land/upgrades.json`](../../misc/deployments/onyx.gno.land/upgrades.json);
the two versions below are the only ones a node joining today needs.

| Role | Version | Commit |
|---|---|---|
| genesis up to height 1203400 | `v1.7.1-rc.1` | `83431a7f0261882541a8a8dcf1643e6f246bbbe5` |
| 1203401 up to the tip | `v1.8.0-rc.1` | `fc79041f4cd1a7b3ed4b316dc5ac14162eb78730` |
| tx-indexer | `v1.5.1` | `08522889...` (see below) |

`v1.8.0-rc.1` **cannot** replay the blocks before its own halt height: it
refuses transactions in a non canonical encoding, which earlier blocks hold.
Starting it at genesis fails within a few blocks with `wrong
Block.Header.AppHash`. Both binaries are therefore required, not just the
latest one.

`GNOROOT` must point at a checkout of the **same tag** as the running binary:
the node loads `gnovm/stdlibs` from it, and the two tags' stdlibs differ.

## gnoland

```shell
git clone https://github.com/gnolang/gno.git gno
cd gno

for TAG in v1.7.1-rc.1 v1.8.0-rc.1; do
  git worktree add --detach "../gnoroot-$TAG" "$TAG"
  ( cd "../gnoroot-$TAG/gno.land" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath \
      -ldflags "-X github.com/gnolang/gno/tm2/pkg/version.Version=$TAG" \
      -o "../../gnoland-$TAG" ./cmd/gnoland )
done
```

`VERSION` must be passed explicitly. A release candidate shares its commit with
the final release, so without it the build reports the final version and fails
the upgrade gate.

Verify:

```
$ ./gnoland-v1.7.1-rc.1 version
gnoland version: v1.7.1-rc.1
$ ./gnoland-v1.8.0-rc.1 version
gnoland version: v1.8.0-rc.1
```

sha256 as built on lyra with go1.25.9:

```
a1b5435fe0faf48eaa05ee301ae76a5b429ec34d3610138549ac15e9fdf747bf  gnoland-v1.7.1-rc.1
c1b05afe2b3bdf0fe24da33c379ce723fac56cf702fbb79215405b1085668189  gnoland-v1.8.0-rc.1
```

These are **not** the checksums on the GitHub release pages. The official
binaries are built by goreleaser with different flags, and `v1.8.0-rc.1`'s
`gnoland_linux_amd64` is dynamically linked against glibc 2.34. The builds
above are static and have no libc dependency at all. Either is a legitimate
`v1.8.0-rc.1`; what matters for consensus is the commit, which is pinned by the
tag.

## tx-indexer

```shell
git clone https://github.com/gnolang/tx-indexer.git
cd tx-indexer
git checkout v1.5.1
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o ../tx-indexer-v1.5.1 ./cmd
```

`./cmd`, not `.`: the repository root has no buildable package.

Commit: `085228830cd5feda125e924720e3c0dfaeaf1663` (tag `v1.5.1`), the tip of `master` as of 2026-11-23, titled
*"fix(deps): build against gno v1.8.0-rc.1 so MsgRun's max_send decodes"*. Its
`go.mod` pins `github.com/gnolang/gno v1.8.0-rc.1`, which is exactly the
version onyx runs. An older tx-indexer decodes onyx's recent `MsgRun`
transactions incorrectly.

sha256 as built on lyra with go1.25.9:

```
138a1b855294c93798916d4a715792840526bc635c5755600d515960acd373b1  tx-indexer-v1.5.1
```

Full commit: `085228830cd5feda125e924720e3c0dfaeaf1663`.
