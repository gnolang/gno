# Ledger support without CGO

## Context

`gnokey` reached Ledger devices through `cosmos/ledger-cosmos-go`, which uses
`zondax/ledger-go`, which uses `zondax/hid`, a CGO binding of hidapi. With
`CGO_ENABLED=0`, the default in the Makefile and in `.github/goreleaser.yaml`,
`zondax/hid` compiles to a stub, so the released `gnokey` binaries cannot use a
Ledger at all. Users had to build `gnokey` themselves with CGO and a C
toolchain for the target ([#3051](https://github.com/gnolang/gno/issues/3051)).

[#5494](https://github.com/gnolang/gno/pull/5494) proposed a shim implementing
the `zondax/hid` API on the pure-Go `rafaelmartins/usbhid`, selected with a
`replace` directive in `go.mod`. That approach has problems beyond its open
review comment:

- `go install github.com/gnolang/gno/gno.land/cmd/gnokey@<version>` refuses a
  module whose `go.mod` has `replace` directives, and gno's has none today.
  A `replace` also applies only to the main module: every contrib that imports
  the keybase (gnokms, gnogenesis, ...) would need its own copy.
- `usbhid` does not expose interface numbers, so the shim reports interface 0
  for every interface. ledger-go's fallback matches interface 0 of a known
  Ledger product, so it can pick the FIDO interface, which some platforms
  enumerate first.
- `usbhid` refuses writes longer than one 64-byte report, while ledger-go
  passes all the reports of an APDU to a single write.
- `github.com/rafaelmartins/usbhid` was archived in September 2026; the
  library continues as `rafaelmartins.com/p/usbhid`.

[pr6332](pr6332_gnokey_ledger_gno_app.md) already moved the Gno.land app onto a
transport of its own (`hidDevice`), writing one report per call.

## Decision

Drop `ledger-cosmos-go`, `ledger-go` and `zondax/hid`. `tm2/pkg/crypto/internal/ledger`
talks to the device itself:

- **HID backend** (`hid_usbhid.go`, Linux, macOS and Windows):
  [`github.com/allinbits/usbhid`](https://github.com/allinbits/usbhid), a fork
  of `rafaelmartins.com/p/usbhid` (see below), which needs no CGO (it uses
  `purego` on macOS and system calls elsewhere). The APDU interface is chosen
  by its usage page, `0xffa0`, so interface numbers are not needed. Other
  platforms build a stub (`hid_unsupported.go`) and report that Ledger devices
  are not supported.
- **Exclusive access**: the device is opened exclusively, because every
  connection uses the same HID channel and their responses could not be told
  apart; a second gnokey, or Ledger Wallet, gets a clear "in use" error. This is
  enforced on macOS and Windows; on Linux `usbhid` takes an advisory `flock`,
  which applications using hidapi, such as Ledger Wallet, ignore.
- **Framing** (`hid.go`): Ledger's HID framing (channel `0x0101`, tag `0x05`,
  sequence numbers, length in the first report), with one report per write.
  Status words other than `0x9000` are returned as a typed `*statusError`
  carrying the response data, instead of ledger-go's error strings.
- **Cosmos app client** (`cosmos.go`): ported from `ledger-cosmos-go` v1.0.0
  under its Apache-2.0 licence (tm2's own). It covers what gnokey uses: the
  version check (2.1.0 or later), the public key and address, and signing in
  48-byte chunks with the sign mode in P2.
- **Discovery**: one connection to the first Ledger device; the Gno.land app is
  asked for its name, then the Cosmos app for its version, on that same
  connection. Another app or the dashboard gives "open the Gno.land or the
  Cosmos app"; a locked device or a transport failure is reported as such.

### HID dependency

The library is used through a fork,
[`github.com/allinbits/usbhid`](https://github.com/allinbits/usbhid), taken at
upstream commit `9e34b6c698d7` and changed only in its module path (v0.1.0).
Upstream has a single maintainer, no tagged releases, and is served from the
maintainer's own domain and git server; its GitHub repository is archived. The
fork keeps a dependency in the signing path off a personal domain. It has its
own module path rather than a `replace`, for the reasons above: a `replace`
would break `go install` and apply to the main module only.

The fork has been audited
([allinbits/usbhid#1](https://github.com/allinbits/usbhid/issues/1)): no
memory-safety defect was found on the Ledger path on any platform. The open
items are a DLL-loading fix on Windows, timeouts, error-path lifecycle on macOS,
and the `purego` version, tracked in that issue to be fixed in the fork. Changes
to the fork are reviewed as security-relevant.

## Alternatives considered

- **#5494's shim and `replace`**, with its issues fixed (per-report writes,
  unknown interface numbers, the current module path). It keeps
  `ledger-cosmos-go`, but the `replace` still breaks `go install @version` and
  must be repeated in every module that imports the keybase.
- **Keep CGO and build releases with a C cross-toolchain.** Release builds
  become heavier and platform-specific, and `go install` users still need a C
  toolchain.
- **Wait for a CGO-free `zondax/hid`.** None is planned.

## Consequences

- Every `gnokey` built for Linux, macOS or Windows, including the release
  binaries and `go install` builds, supports the Gno.land and Cosmos apps.
- The dependency graph loses `ledger-cosmos-go`, `ledger-go`, `zondax/hid` and
  `zondax/golem`, and gains `github.com/allinbits/usbhid` and `purego`. `go.mod`
  needs no `replace`.
- The HID library is a young, single-maintainer module, and its macOS backend
  calls IOKit and CoreFoundation through `purego`. Through the fork, its
  maintenance, audit and updates are on us.
- The Cosmos client is now maintained here. It is about 200 lines, and the
  Cosmos app's protocol has been stable for years.
- Tests cover the framing, the transport, both clients against fake devices
  following each app's protocol, and app selection. Both apps were also used
  through a `CGO_ENABLED=0` gnokey on a real device.
