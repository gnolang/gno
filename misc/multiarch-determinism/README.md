# multiarch-determinism

A small Go driver that **embeds gnovm directly** and runs a fixed
corpus of gno stdlib operations, used by CI to verify byte-identical
output across CPU architectures.

## Shape

- `main.go` — builds a `gno.Machine` using `gnovm/pkg/test.ProdStore`
  (which loads the full gno stdlib), parses the embedded `corpus.gno`,
  and calls `m.RunFiles` + `m.RunMain`. The machine's `Output` is
  wired to `os.Stdout`.
- `corpus.gno` — a `package main` gno program that imports
  `crypto/sha256`, `crypto/keccak256`, `crypto/ed25519`,
  `crypto/chacha20`, `crypto/bech32`, `crypto/subtle`,
  `crypto/modexp`, `crypto/merkle`, `crypto/bn254`, `hash/adler32`,
  ... and prints one canonical `<op> <args_hex...> <output>` line per
  case via gno's `println`.

The driver is a separate binary from the `gno` CLI on purpose: it
avoids the CLI surface (flag parsing, lint, fmt, project loading) and
only links what's needed to exercise gnovm + stdlibs. The resulting
binary is ~33 MB cross-compiled.

## What this tests — and why this shape matters

Because the program runs inside an embedded `gno.Machine`, every
stdlib call traverses gnovm's evaluation primitives:

- **Native bindings** (e.g. `crypto/sha256.Sum256` is `// injected`,
  `crypto/ed25519.Verify` likewise) flow through gnovm's Go2Gno /
  Gno2Go marshalling and the native dispatch table — that's where
  arch-sensitive assembly fallbacks in the underlying Go
  implementation could surface.
- **Pure-gno code** (e.g. `crypto/chacha20`, `hash/adler32`,
  `crypto/bech32`) runs through gnovm's interpreter — that's where
  int/uintptr sizing, allocation order, or interpreter heuristics
  could surface.

Both layers run on every line of the corpus. This is what user code
on-chain actually exercises — not Go's stdlib directly, and not the
`gno` CLI's higher-level layers.

## CI

`.github/workflows/ci-multiarch-determinism.yml` cross-compiles the
driver with `CGO_ENABLED=0` for `linux/amd64` and `linux/arm64` on a
single Ubuntu runner, runs each binary under
[`qemu-user`](https://www.qemu.org/docs/master/user/main.html), and
fails the job on any byte-level diff between the captured stdouts.

Running on one host (instead of a matrix of GitHub runners) removes
runner-to-runner variability and keeps the comparison cheap — seconds
of CPU per run.

## Running locally

```sh
# from the repo root
export GNOROOT=$(pwd)
go build -trimpath -o /tmp/mad-amd64 ./misc/multiarch-determinism
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath \
  -o /tmp/mad-arm64 ./misc/multiarch-determinism

/tmp/mad-amd64              > /tmp/amd64.txt
qemu-aarch64 /tmp/mad-arm64 > /tmp/arm64.txt
diff /tmp/amd64.txt /tmp/arm64.txt
```

(`apt install qemu-user` on Debian/Ubuntu.)

`GNOROOT` is required so `gnovm/pkg/test.ProdStore` can locate the
stdlib `.gno` sources. The driver falls back to
`gnoenv.GuessRootDir()` when the env var is unset.

## Extending the corpus

Add cases by editing `corpus.gno`:

1. Import the new gno stdlib package.
2. Add a `runFoo()` function that iterates a fixed list of inputs and
   calls `emit("op-name", args..., result)` for each.
3. Call it from `main()` in a stable position.

Keep inputs deterministic — no timestamps, no PRNG state, no map
iteration order. Cover boundary lengths (empty, one block, multi-block,
unaligned) and known-tricky values (all-zero, all-ones).

**Convention:** PRs that add a new gno stdlib are expected to extend
this corpus in the same change, so a single CI run attests that the
new primitive is bytewise-deterministic across all supported
architectures.
