package cid

import (
	"crypto/sha256"
	"errors"
)

// Codecs of the content a CID names (multicodec table). A CID may carry
// any codec; these are the common ones.
const (
	Raw       uint64 = 0x55
	DagPB     uint64 = 0x70
	DagCBOR   uint64 = 0x71
	Libp2pKey uint64 = 0x72
	DagJSON   uint64 = 0x0129
)

// maxCIDLen is the longest binary CID accepted: the version byte, two
// 9-byte varints, a 2-byte digest length and a 128-byte digest.
const maxCIDLen = 149

var errCID = errors.New("cid: invalid CID")

// CID is an IPFS content identifier. It holds the CID's binary form, so it
// is immutable and comparable with ==. The zero value is undefined (see
// Defined).
type CID struct{ b string }

// Parse parses the string form of a CID: a CIDv0 ("Qm...", 46 characters of
// base58btc) or a multibase-encoded CIDv1.
func Parse(s string) (CID, error) {
	if len(s) == 46 && s[:2] == "Qm" {
		b, err := decodeBaseN(s, 58, base58Digit)
		if err != nil || !isV0(string(b)) {
			return CID{}, errCID
		}
		return CID{string(b)}, nil
	}

	b, err := decodeMultibase(s)
	if err != nil {
		return CID{}, err
	}
	if isV0(string(b)) {
		return CID{}, errCID // a CIDv0 is only ever written in bare base58btc
	}
	return Decode(b)
}

// Decode parses the binary form of a CID, which must span all of b.
func Decode(b []byte) (CID, error) {
	c, n, err := DecodeFirst(b)
	if err != nil {
		return CID{}, err
	}
	if n != len(b) {
		return CID{}, errCID
	}
	return c, nil
}

// DecodeFirst decodes the binary CID at the start of b and returns it with
// its length, for formats that store CIDs next to other data (CAR files,
// dag-pb links).
func DecodeFirst(b []byte) (CID, int, error) {
	if len(b) > maxCIDLen {
		b = b[:maxCIDLen]
	}
	s := string(b)
	if len(s) >= 34 && isV0(s[:34]) {
		return CID{s[:34]}, 34, nil
	}

	version, n, err := uvarint(s)
	if err != nil {
		return CID{}, 0, err
	}
	if version != 1 {
		return CID{}, 0, errCID
	}
	_, m, err := uvarint(s[n:])
	if err != nil {
		return CID{}, 0, err
	}
	k, err := multihashLen(s[n+m:])
	if err != nil {
		return CID{}, 0, err
	}
	end := n + m + k
	return CID{s[:end]}, end, nil
}

// NewV1 returns the CIDv1 of content with the given codec and multihash.
func NewV1(codec uint64, mh Multihash) (CID, error) {
	if codec > maxVarint || mh.b == "" {
		return CID{}, errCID
	}
	return CID{string(appendUvarint([]byte{1}, codec)) + mh.b}, nil
}

// Sum returns the CIDv1 of data with the given codec, hashed with sha2-256.
// For a file that `ipfs add --cid-version=1` stores as a single block (up to
// 256 KiB with the default chunker), Sum(Raw, file) is the CID it prints.
// It panics if codec does not fit in a 9-byte varint.
func Sum(codec uint64, data []byte) CID {
	h := sha256.Sum256(data)
	c, err := NewV1(codec, Multihash{"\x12\x20" + string(h[:])})
	if err != nil {
		panic(err)
	}
	return c
}

// isV0 reports whether b is a CIDv0: a bare 34-byte sha2-256 multihash.
func isV0(b string) bool {
	return len(b) == 34 && b[0] == 0x12 && b[1] == 0x20
}

// Defined reports whether c is a CID rather than the zero value.
func (c CID) Defined() bool {
	return c.b != ""
}

// Version returns 1 for a CIDv1, whose binary form starts with the version
// byte 1, and 0 for a CIDv0.
func (c CID) Version() uint64 {
	if c.b != "" && c.b[0] == 1 {
		return 1
	}
	return 0
}

// split returns the codec and the multihash of c.
func (c CID) split() (uint64, Multihash) {
	switch {
	case c.b == "":
		return 0, Multihash{}
	case c.Version() == 0:
		return DagPB, Multihash{c.b}
	}
	codec, n, _ := uvarint(c.b[1:])
	return codec, Multihash{c.b[1+n:]}
}

// Codec returns the codec of the content c names. A CIDv0 always names
// dag-pb content.
func (c CID) Codec() uint64 {
	codec, _ := c.split()
	return codec
}

// Multihash returns the multihash c carries.
func (c CID) Multihash() Multihash {
	_, mh := c.split()
	return mh
}

// V1 returns c as a CIDv1: a CIDv0 becomes the CIDv1 of the same dag-pb
// content, which is how to compare the two versions.
func (c CID) V1() CID {
	if c.b == "" || c.Version() == 1 {
		return c
	}
	return CID{"\x01\x70" + c.b} // version 1, dag-pb
}

// String returns the canonical string form of c: base58btc for a CIDv0 and
// base32 for a CIDv1. It returns "" for the zero value.
func (c CID) String() string {
	switch {
	case c.b == "":
		return ""
	case c.Version() == 0:
		return encodeBaseN([]byte(c.b), base58Alphabet)
	}
	return "b" + encodeBase32([]byte(c.b), base32Lower)
}

// Encode returns a CIDv1 in the multibase whose prefix is base: 'b' or 'B'
// (base32), 'z' (base58btc), 'k' (base36), 'f' or 'F' (base16). A CIDv0 has
// no multibase form; encode c.V1() instead.
func (c CID) Encode(base byte) (string, error) {
	if c.Version() != 1 {
		return "", errCID
	}
	return encodeMultibase(base, []byte(c.b))
}

// Bytes returns a copy of the binary form of c.
func (c CID) Bytes() []byte {
	return []byte(c.b)
}

// KeyString returns the binary form of c as a string, a compact and exact
// key for maps and avl trees. It identifies the CID, not the content: CIDs
// with different codecs, a CIDv0 and its V1, and the same bytes hashed with
// different functions all name the same content. To key by content, accept
// one hash function (Multihash().Code() == SHA256) and key by
// string(c.Multihash().Bytes()).
func (c CID) KeyString() string {
	return c.b
}

// Verify reports whether block is the content c names: it hashes block with
// the CID's hash function and compares the digests. It returns
// ErrDigestMismatch if they differ and ErrUnsupportedHash for a hash
// function other than sha2-256, keccak-256 or identity.
//
// Verify compares digests only and ignores the codec: any CID whose
// multihash matches block verifies. Of the CIDs `ipfs add` prints, only the
// Raw CID of a single-block file has the file's bytes as its block; other
// files are named by a DAG root node. Check Codec() == Raw where that
// matters. An identity CID contains its content, so verifying it proves
// nothing about storage or possession; reject Multihash().Code() == Identity
// where that matters.
func (c CID) Verify(block []byte) error {
	if c.b == "" {
		return errCID
	}
	return c.Multihash().verify(block)
}
