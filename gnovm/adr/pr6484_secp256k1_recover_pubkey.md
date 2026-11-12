# crypto/secp256k1: public-key recovery

## Status

Proposed in #6484, stacked on #5688, which adds `crypto/secp256k1` with
`Verify`. This adds `RecoverPubkey` to the same package.

## Context

Ethereum-style signers are known by an address, not a public key. A contract
checks their signature with `ecrecover`: it recovers the public key from a
32-byte digest and a signature (r, s, v), derives the address from the key and
compares it with the expected signer. Protocols built on Ethereum keys work
the same way. Wormhole is one: its guardians sign the double Keccak-256 hash of
a message, and a verifier recovers each guardian's address from the signature.

`Verify` cannot serve these callers. It needs the signer's compressed public
key, which they do not have, and it hashes the message with SHA-256 itself,
while they sign a Keccak-256 hash. Recovery written in Gno would be big-number
curve arithmetic in the interpreter, orders of magnitude over what a native
costs.

## Decision

Add one function:

```go
func RecoverPubkey(digest []byte, signature []byte) (pubkey [65]byte, ok bool)
```

- **Inputs.** The digest is exactly 32 bytes and is not hashed again. The
  signature is exactly 65 bytes: r and s as 32-byte big-endian integers, then a
  recovery ID of 0 or 1. Ethereum's v of 27/28 is not accepted.
- **Output.** The uncompressed key, `0x04 || X || Y`, and true. On any refusal
  (wrong length, r or s outside [1, n-1], a recovery ID other than 0 or 1, an r
  that is no curve point, a signature that recovers the point at infinity) a
  zero key and false. It never panics on its input.
- **High-S signatures are accepted**, as `ecrecover` accepts them.
- **The Gno wrapper checks the lengths and copies the bytes into fresh
  `[32]byte` and `[65]byte` arrays**, then calls the private native
  `recoverPubkey(digest [32]byte, signature [65]byte) ([65]byte, bool)`. The
  generated binding converts arguments before the native body runs, so a
  native taking slices would copy a caller-sized slice for a flat price; fixed
  arrays make the conversion the same for every call. The copies use `copy`,
  not byte loops.
- **The native calls decred's `ecdsa.RecoverCompact`** (secp256k1/v4, already a
  dependency; tm2's default secp256k1 code uses it through btcec). It is
  called directly rather than through tm2, so every build runs the same code
  whatever its build tags. The native refuses recovery IDs above 1 before the
  call: `RecoverCompact` would also take 2 and 3 (an x-coordinate of r + n) and
  the compressed-key headers, which would make several IDs recover the same
  key.
- **Gas: one flat row**, charged before the native runs. Every 32/65-byte input
  pays it, including the ones the native refuses, early or late; wrong lengths
  stop in the wrapper and pay ordinary VM gas only.

### Price

The row is preliminary, like `Verify`'s row in #5688, and it is derived from
that row rather than fitted. `gen_native_table.py` has a spec for `Verify` but
none for `recoverPubkey`, so a fitter run drops this row instead of producing a
wrong one. To re-derive it: re-fit `Verify` on the reference machine, rerun
`BenchmarkNative_Secp256k1_RecoverPubkey_PerVerify` there, and take the slowest
input's ratio × the `Verify` row at 64 bytes × 1.2.

`_PerVerify` runs a batch of verifications of a 64-byte message and a batch of
recoveries back to back, for each of 67 valid inputs, 200 rounds. The ratio of
two batches run back to back holds on a loaded machine, where single timings
swing several-fold; it was measured on an Apple M1 Pro under load. The inputs
are 64 signatures from deterministic keys, signer 0 of the corpus, its high-S
twin, and a dense-s input: the same digest and r, with s chosen so that
u2 = s/r has a dense signed-digit form, which makes the variable-base
multiplication do close to its most point additions. Three runs put a typical
recovery at 1.088–1.090 times a verification and the dense-s input, the
slowest, at 1.275–1.279. The row is 1.279 × 227,309 (the `Verify` row at 64
bytes) × 1.2 = 348,874, rounded up to 350,000. Without the dense-s input, the
slowest of the others measured 1.140–1.172: sampled signatures miss the slow
case.

Single timings (fastest of 10 rounds) show what refusals cost next to a
recovery of about 166 µs: 0.6 µs for a bad recovery ID or scalar, 15 µs for an
r with no curve point (a square root), and 155 µs for a signature that
recovers the point at infinity, which shows only after both multiplications.
One flat charge covers all of them.

`secp256k1_recover_bench_m1pro.txt` in `gnovm/cmd/calibrate` holds the runs.

## Alternatives considered

- **Return the 20-byte Ethereum address** (an `Ecrecover`): saves EVM callers
  one Keccak-256, but ties the primitive to Ethereum's address format. Callers
  derive it with `crypto/keccak256` in one line.
- **Accept v = 27/28 or EIP-155 values**: several encodings of one signature.
  Callers subtract 27; the function takes one encoding.
- **Hash inside, as `Verify` does**: every protocol picks its own hash (Ethereum
  Keccak-256, Wormhole double Keccak-256), so the caller passes the digest.
- **Refuse high-S signatures**: `ecrecover` accepts them, so this would refuse
  valid Ethereum signatures. The documentation says a signature is malleable
  and must not be used as a replay key instead.
- **Delegate to tm2, as `Verify` does**: tm2 has no recovery in its default
  build; its libsecp256k1 binding is internal and behind a build tag.
- **A per-byte slope**: the native's inputs have a fixed size, so there is
  nothing to scale with.

## Consequences

- Callers must compare the recovered key or address with the one they expect:
  a wrong digest or recovery ID usually still recovers a key. The doc comment
  says so.
- Callers must hash the message they authorize themselves and never take the
  digest from untrusted input. For a digest of their choosing, anyone can build
  a signature that recovers any public key, without its private key: pick u
  and w, let R = uG + wQ, r = R.x, s = r/w and digest = ru/w (digest 0 when
  u = 0). A test pins one such case. `ecrecover` has the same property; `Verify`
  avoids it by hashing the message itself. The doc comment says so too.
- The stdlib's source is part of genesis state, so the pinned hash in
  `gno.land/pkg/sdk/vm/apphash_crossrealm38_test.go` moves again.
- Activation is the same as for the rest of #5688. A new stdlib reaches an
  existing chain only through a release and a genesis replay: `loadStdlibs`
  runs only in `InitChainer`, and no migration framework exists (master's
  `gno.land/cmd/gnoland/UPGRADES.md` says so). On this base, a binary carrying a
  new stdlib package also panics at boot on existing state:
  `VMKeeper.Initialize` type-checks every `stdlibs.InitOrder()` entry and the
  missing package comes back nil. Checked with real binaries: a dev chain
  started by `gnoland` at #5688's merge base (2ed70a202e) and restarted on
  #5688's head binary panics in `TypeCheckMemPackage(nil)` under
  `VMKeeper.Initialize`, and identically with this change. #5688's package
  causes it; this change neither causes nor fixes it.

## What we checked

- The corpus in `gnovm/stdlibs/crypto/secp256k1/testdata` was generated from
  public test keys with go-ethereum v1.10.21: its pure-Go and libsecp256k1
  recovery and its Berlin `ecrecover` precompile. The native gives the outcome
  all three agree on for the 19 fixed-size vectors (full keys compared), and
  matches the precompile on every recovery byte 0–255 for two signatures, one
  with r = 2.
- 256 signatures from random keys recover their key, and so do their high-S
  twins; 64 more recovered keys are accepted by `Verify`, which goes through
  libsecp256k1 under `-tags libsecp256k1`.
- Inputs the corpus lacks are pinned: the point at infinity (refused; the
  other recovery ID recovers another key), a chosen digest that recovers
  signer 0's key, and the dense-s input. `FuzzRecoverPubkey`, seeded with the
  corpus, checks that no input panics, that refusals return a zero key and that
  recovered keys are valid points.
- In a Machine, through the generated binding and the gas table: all 28
  vectors through the wrapper; the native runs only for 32/65-byte input and
  always receives fixed-size arrays, even behind a 1 MiB capacity; a real gas
  meter one unit short of the native charge stops before the native runs; every
  fixed-size input uses the same gas, including a recovery ID of 2 (refused
  first) and the point at infinity (refused last).
- `gno.land/pkg/integration/testdata/stdlib_secp256k1_recover.txtar`: a realm
  imports the package through the keeper and writes state and emits an event
  before each recovery. A refused signature, a call short of gas, and a
  two-message transaction whose second message is refused all fail, keep none
  of those writes and return no events, while the failed transaction still
  pays its fee and uses its sequence number. The same call uses the same gas
  before and after a restart.
- `misc/multiarch-determinism/corpus.gno` now includes recovery and its
  refusals, so CI compares their output on amd64 and arm64.
- Locally only: the native against tm2's libsecp256k1 binding (`-tags
  libsecp256k1`), with the same recovery-ID boundary. The 531 corpus inputs,
  20,000 signatures from random keys with their high-S twins, and 20,000
  random digests and scalars (about half recover a key) give the same result
  from both. That test has to live in tm2's directory, the only one allowed
  to import the internal binding, so it is not part of this change.
