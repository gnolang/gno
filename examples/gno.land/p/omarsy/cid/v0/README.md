> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `cid` - IPFS content identifiers

Parse, format and verify [CIDs](https://github.com/multiformats/cid). A realm
can keep a CID in its state while the content stays on IPFS, and check that
bytes someone submits are the content the CID names.

```go
import "gno.land/p/omarsy/cid/v0"

c, err := cid.Parse("bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44yegnrjhc4yeq")
if err != nil {
	panic(err)
}
err = c.Verify([]byte("hello")) // nil: "hello" is the content c names
```

- `Parse` and `Decode` accept CIDv0 (`Qm...`) and multibase CIDv1 (`b`, `B`,
  `z`, `k`, `f`, `F`). Parsing is strict: each CID has one accepted form
  per multibase, so compare parsed CIDs, never input strings.
- `String` gives the canonical form (base58btc for CIDv0, base32 for CIDv1),
  `Encode` the other multibase forms. `V1` converts a CIDv0 so the two
  versions can be compared.
- `Sum(codec, data)` is the CIDv1 of `data`, hashed with sha2-256.
  `Sum(cid.Raw, file)` is what `ipfs add --cid-version=1` prints for a file
  of up to 256 KiB.
- `Verify(block)` supports sha2-256, keccak-256 and identity multihashes. It
  checks one block: only a Raw CID of a single-block file verifies against
  the file's bytes. An identity CID contains its content, so verifying it
  proves nothing about storage; reject `Identity` where that matters.
- CID values are immutable and comparable with `==`, so they are safe to
  store, return and use as keys (`KeyString`). A key identifies the CID, not
  the content: CIDs with different codecs, and a CIDv0 and its V1, can name
  the same bytes. Key by `string(c.Multihash().Bytes())` to identify content.
- `DecodeFirst` reads a CID at the start of a buffer, as stored in CAR files
  and dag-pb links.

Gas, measured with `gno test`: parsing a CIDv1 costs about 0.75M in base32
and 1.3M in base58btc, parsing a CIDv0 about 1.2M, and verifying a 256 KiB
block about 2.4M.
