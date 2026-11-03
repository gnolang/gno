# ADR: Build the linux/amd64 release binaries in an Ubuntu 22.04 container

## Context

`release / chain-tag` (added in #5420) builds the linux/amd64 binaries natively on `ubuntu-latest`, so cgo is on. A binary built with cgo needs at most the glibc of the system that built it, and may need exactly that. `ubuntu-latest` has been Ubuntu 24.04 (glibc 2.39) since before #5420, and on it `gnoland`, `gnokey` and `gnodev` need glibc 2.38: their C code (the `lmdb` and `mdbx` database backends, `zstd`, and `zondax/hid` for Ledger) is compiled against 24.04's headers. The goreleaser builds before #5420 used `CGO_ENABLED=0`. v1.7.0's `gnoland_linux_amd64` needs `GLIBC_2.38` and does not start on Debian 12 (glibc 2.36) or Ubuntu 22.04 (2.35), and v1.5.1's has the same problem. Yet the validator guide recommends that native binary.

The linux/arm64 binaries are cross-compiled from the same runner. Go turns cgo off by default when cross-compiling, so they are static. The release Docker images build with `CGO_ENABLED=0`, so they are static too.

## Decision

1. The linux/amd64 row of `release / chain-tag` runs in an `ubuntu:22.04` container, pinned by digest, on `ubuntu-latest`. A step installs the compiler and sets `CGO_ENABLED=1`, because Go turns cgo off by itself when it finds no C compiler, which would drop `lmdb`, `mdbx` and Ledger support without failing anything. Built in amd64 Ubuntu 22.04, all five binaries need `GLIBC_2.34` at most, keep `lmdb`, `mdbx` and `zondax/hid`, and start on Debian 12, Ubuntu 22.04 and AlmaLinux 9.
2. A step refuses any Linux binary that needs a glibc newer than 2.35, before the upload. A change of build system that raises the floor, such as a newer image, then fails the release run instead of shipping. It reads each binary with `readelf`, which reads any ELF, so the static arm64 binaries are checked too. It fails as well on anything it cannot read, and on an empty `dist/`.

## Alternatives considered

- **The `ubuntu-22.04` runner.** Same glibc, and simpler: the row only names another runner. But GitHub began deprecating it on 2026-09-17, and retires it on 2027-04-17 (actions/runner-images#14254). It browns out before that, failing any job that starts then, and once retired the linux/amd64 leg would fail before any step. The default `fail-fast` then cancels the other legs, so no tag would publish anything until the workflow changed. A container does not depend on GitHub's runner lifecycle.
- **Build the Linux binaries without cgo,** as the arm64 binaries and the Docker images already are. They would be static and start on any Linux. But on linux/amd64, `gnoland` and `gnodev` would lose the `lmdbdb` and `mdbxdb` backends, and `gnokey` and `gnodev` would lose Ledger support, since `zondax/hid` finds no device without cgo. That is a regression for anyone using those today.
- **Without cgo for `gnoland`, `gno` and `gnoweb` only,** keeping cgo in the container for `gnokey` and `gnodev`. That means two build paths in one workflow, and `gnoland` still loses `lmdbdb` and `mdbxdb`.
- **Document the 2.38 floor in the validator guide.** The binary would still not start on two common server systems.

## Consequences

- From the first release built after this, the linux/amd64 binaries start on glibc 2.35 or newer: Ubuntu 22.04 and Debian 12 or later. They still do not start on older systems, or on musl-based ones such as Alpine, which use the Docker images instead.
- Each release run installs the compiler from Ubuntu's archive, so it needs that archive to answer. A failure there fails the run, and re-running it recovers.
- Moving to a newer image would raise the floor. The glibc step would fail that release run, so the move has to go with a deliberate new ceiling.
- Releases already published keep their binaries: v1.7.0's still needs 2.38.
- A tag push runs the tagged commit's workflows, so this applies to mainnet releases once it is on `chain/mainnet`.
