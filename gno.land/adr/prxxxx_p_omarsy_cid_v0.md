# ADR: `p/omarsy/cid/v0`, IPFS content identifiers in Gno

## Context

Realms increasingly point at content on IPFS (NFT metadata, avatars, files),
but nothing in the examples could parse or check a CID: `ipfs://` strings were
stored and rendered as opaque text. Storing the content itself on chain is not
an option beyond small sizes (the storage deposit is 100 ugnot per byte, and a
transaction is at most 1 MB).

A CID is a hash, so a realm can hold the CID while the content stays on IPFS,
and check that bytes someone submits are the content the CID names. Gno already
has what that needs: native `crypto/sha256` and `crypto/keccak256`. What was
missing is the multiformats layer (unsigned varints, multihash, multibase,
base58btc, the CID binary format).

The package lives under `p/omarsy` rather than `p/nt` because `p/nt/*` paths
cannot be added to mainnet after genesis
(`misc/deployments/mainnet.gno.land/gen-genesis.sh`), while a personal
namespace can be deployed today.

## Decision

- `CID` and `Multihash` hold their binary form in a string: immutable,
  comparable with `==`, safe to store in realm state, return to other realms,
  and use as keys. There are no exported mutators and no shared byte slices.
- `Parse` and `Decode` are strict, so each CID has one accepted form:
  - unsigned varints must be minimal and at most 9 bytes;
  - sha2-256 and keccak-256 digests must be 32 bytes, and any digest at most
    128 bytes;
  - no trailing bytes; version 1 only in the CIDv1 form, and a CIDv0 only as
    the bare 46-character base58btc string;
  - base encodings must be canonical: no padding, whitespace, case mixing, or
    non-zero leftover bits.
- `DecodeFirst` reads a CID at the start of a buffer. CAR sections and dag-pb
  links store CIDs next to other data, and a deployed package cannot gain
  functions later, so the future range verifier needs it now.
- Supported multibases are the ones IPFS tools emit: base32 (`b`, `B`),
  base58btc (`z`), base36 (`k`) and base16 (`f`, `F`).
- `Verify` supports sha2-256, keccak-256 and identity, compares full digests,
  and returns `ErrUnsupportedHash` for anything else rather than skipping it.
- Gas shapes the codecs:
  - digits are mapped with arithmetic, not by searching the alphabet;
  - base58btc and base36 convert with multi-digit limbs;
  - an input longer than any CID in its base is rejected before decoding.

  An uncapped 512-character base58btc string cost 561M gas to reject with the
  first, digit-by-digit decoder and about 37M with limbs; with the cap, the
  worst rejected input costs about 9M.

Gas, measured with `gno test` as the difference from an empty test:

| Operation | Gas |
|---|---|
| `Parse`, base32 CIDv1 | 0.74M |
| `Parse`, base58btc CIDv1 | 1.3M |
| `Parse`, CIDv0 | 1.2M |
| `String`, CIDv1 / CIDv0 | 0.28M / 0.73M |
| `Parse`, longest rejected input (base58btc, 204 characters) | 9M |
| `Verify`, 256 KiB block (sha2-256) | 2.4M |

Tests use the multibase spec's vectors, CIDv0/v1 pairs from the IPFS docs, and
two root blocks fetched from IPFS. Every expected value was cross-checked with
an independent implementation. The limb codecs are checked against the plain
digit-by-digit conversion at every length up to the longest CID. Each
strictness rule, and each length cap, has a test that fails when the rule is
removed (checked by mutating the code).

## Alternatives considered

- **`encoding/binary`, `encoding/base32` and `encoding/hex` from the standard
  library.** They are not strict. `Uvarint` accepts non-minimal and 10-byte
  forms, and base32 strips `\r\n`. Wrapping them with extra checks or a
  re-encode-and-compare would work, but loading such a package costs a
  transaction up to about 0.9M gas the first time it is used (measured for
  `encoding/binary` and `encoding/base32`), more than the code it replaces.
- **Put it in `p/nt`.** Not deployable on mainnet without a hardfork.
- **Separate multibase, multihash or varint packages.** Nothing outside CIDs
  uses them yet. Each extra package is another path frozen at deploy, and
  another dependency to deploy in order.
- **Ship the byte-range verifier (CAR and UnixFS) now.** Deferred to the
  pinning-bounty design that will use it. Its soundness depends on that threat
  model: taking a leaf's type from the parent link's codec, checking blocksizes
  against real sizes, rejecting identity roots.

## Consequences

- Realms can parse user-supplied CIDs, store them compactly, and verify
  single-block content (up to 256 KiB with default chunking) on chain.
- `Verify` checks one block: only a Raw CID of a single-block file verifies
  against the file's bytes; other files are named by a DAG root node. The docs
  say so.
- A CID identifies content only up to its codec and version: CIDs with
  different codecs, and a CIDv0 and its V1, name the same bytes. Realms that
  key by content should key by the multihash, and realms that pay for storage
  should refuse identity CIDs, which contain their content. The docs say so.
- The API freezes at the first mainnet deploy: a deployed path cannot be
  changed, so later additions go to a new version path.
- Deploying to mainnet needs the `omarsy` namespace registered (or an address
  namespace).
