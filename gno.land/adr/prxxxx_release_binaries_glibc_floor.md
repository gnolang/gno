# ADR: Build the linux/amd64 release binaries on Ubuntu 22.04

## Context

`release / chain-tag` builds the linux/amd64 binaries natively on `ubuntu-latest`, so cgo is on, and a binary built with cgo needs at least the glibc of the runner that built it. Once `ubuntu-latest` became Ubuntu 24.04, `gnoland`, `gnokey` and `gnodev` started needing glibc 2.38: their C code (the `lmdb` and `mdbx` database backends, `zstd`, and `zondax/hid` for Ledger) was compiled against 24.04's headers. v1.7.0's `gnoland_linux_amd64` needs `GLIBC_2.38` and does not start on Debian 12 (glibc 2.36) or Ubuntu 22.04 (2.35), and v1.5.1's has the same problem. Yet the validator guide recommends that native binary.

The linux/arm64 binaries are cross-compiled from the same runner. Go turns cgo off by default when cross-compiling, so they are static. The release Docker images build with `CGO_ENABLED=0`, so they are static too.

## Decision

1. The linux/amd64 row of `release / chain-tag` runs on `ubuntu-22.04`, whose glibc is 2.35. A `gnoland` built with cgo on Ubuntu 22.04 (checked on an arm64 build) needed `GLIBC_2.34` at most, kept `mdbx` and `lmdb`, and started on Debian 12 and Ubuntu 22.04. The step below checks the real linux/amd64 build on every release.
2. A step refuses any Linux binary that needs a glibc newer than 2.35, before the upload, so a runner change that raises the floor again fails the release run instead of shipping. It reads each binary's version needs with `readelf`, which reads any ELF, so the static arm64 binaries are checked too.

## Alternatives considered

- **Build the Linux binaries without cgo,** as the arm64 binaries and the Docker images already are. They would be static and start on any Linux. But on linux/amd64, `gnoland` and `gnodev` would lose the `lmdbdb` and `mdbxdb` backends, and `gnokey` and `gnodev` would lose Ledger support, since `zondax/hid` finds no device without cgo. That is a regression for anyone using those today.
- **Without cgo for `gnoland`, `gno` and `gnoweb` only,** keeping cgo on `ubuntu-22.04` for `gnokey` and `gnodev`. That means two build paths in one workflow, and `gnoland` still loses `lmdbdb` and `mdbxdb`.
- **Document the 2.38 floor in the validator guide.** The binary would still not start on two common server systems.

## Consequences

- From the first release built after this, the linux/amd64 binaries start on glibc 2.35 or newer: Ubuntu 22.04 and Debian 12 or later. They still do not start on older systems, or on musl-based ones such as Alpine, which use the Docker images instead.
- GitHub will retire `ubuntu-22.04` some day. When the row moves to a newer runner, the glibc step fails the run. The fix is then a container with an older glibc, not a raised ceiling.
- Releases already published keep their binaries: v1.7.0's still needs 2.38.
- A tag push runs the tagged commit's workflows, so this applies to mainnet releases once it is on `chain/mainnet`.
