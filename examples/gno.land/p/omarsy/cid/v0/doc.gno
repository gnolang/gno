// v0 - Unaudited: This is an initial version that has not yet been formally audited.
// A fully audited version will be published as a subsequent release.
// Use in production at your own risk.
//
// Package cid parses, formats and verifies IPFS content identifiers (CIDs).
//
// A CID names content by its hash. A realm can keep a CID in its state
// while the content stays on IPFS, and check that bytes someone submits are
// the content the CID names:
//
//	c, err := cid.Parse("bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44yegnrjhc4yeq")
//	if err != nil {
//		panic(err)
//	}
//	err = c.Verify([]byte("hello")) // nil: "hello" is the content c names
//
// Parsing is strict, so a CID has exactly one accepted binary form:
// non-minimal varints, wrong digest lengths, trailing bytes and
// non-canonical base encodings are rejected. Supported encodings are CIDv0
// (bare base58btc) and multibase base32 (b, B), base58btc (z), base36 (k)
// and base16 (f, F). Verify supports sha2-256, keccak-256 and identity
// multihashes.
//
// CID and Multihash values are immutable and comparable with ==. They are
// safe to store in realm state, return to callers, and use as keys
// (KeyString). A key identifies the CID, not the content: CIDs with
// different codecs, a CIDv0 and its V1, and the same bytes hashed with
// different functions all name the same content. To key by content, accept
// one hash function (Multihash().Code() == SHA256) and key by
// string(c.Multihash().Bytes()). Verify compares digests only and ignores
// the codec. An identity CID contains its content, so Verify proves nothing
// about storage for it.
// DecodeFirst reads a CID stored next to other data, as in CAR files and
// dag-pb links.
package cid
