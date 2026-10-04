package cid

import (
	"encoding/hex"
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// Expected values come from an independent implementation, except the
// docs pair (docs.ipfs.tech) and the root blocks (fetched from IPFS).
const (
	helloCID = "bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44yegnrjhc4yeq" // raw, sha2-256("hello")
	docsV0   = "QmbWqxBEKC3P8tqsKc98xmWNzrzDtRLMiMPL8wBuTGsMnR"
	docsV1   = "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi"
)

func TestParseEncodings(t *testing.T) {
	want, err := Parse(helloCID)
	urequire.NoError(t, err)

	encodings := []struct {
		base byte
		s    string
	}{
		{'b', helloCID},
		{'B', "BAFKREIBM6JG3UX5QUMHCN2B3FLC3TYU6DMLB4XA7U5BF44YEGNRJHC4YEQ"},
		{'z', "zb2rhZfjRh2FHHB2RkHVEvL2vJnCTcu7kwRqgVsf9gpkLgteo"},
		{'k', "k2cwue9rqdypmt3thjky14z1tk9fi9f0o5w7b3ofitdewlcf87lismqs"},
		{'f', "f015512202cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{'F', "F015512202CF24DBA5FB0A30E26E83B2AC5B9E29E1B161E5C1FA7425E73043362938B9824"},
	}
	for _, e := range encodings {
		c, err := Parse(e.s)
		urequire.NoError(t, err, e.s)
		uassert.True(t, c == want, e.s)

		s, err := want.Encode(e.base)
		urequire.NoError(t, err)
		uassert.Equal(t, e.s, s)
	}

	uassert.Equal(t, helloCID, want.String())

	_, err = want.Encode('x')
	uassert.Error(t, err)
}

func TestSum(cur realm, t *testing.T) {
	uassert.Equal(t, helloCID, Sum(Raw, []byte("hello")).String())
	uassert.Equal(t, "bafkreihdwdcefgh4dqkjv67uzcmw7ojee6xedzdetojuzjevtenxquvyku", Sum(Raw, nil).String())
	// dag-json's code, 0x0129, takes a two-byte varint.
	uassert.Equal(t, "baguqeeraiqjw7i2vwntyuekgvulpp2det2kpwt6cd7tx5ayqybqpmhfk76fa", Sum(DagJSON, []byte("{}")).String())
	uassert.Equal(t, DagJSON, Sum(DagJSON, nil).Codec())

	uassert.PanicsWithMessage(t, cur, errCID.Error(), func() {
		Sum(1<<63, nil)
	})
}

func TestV0(t *testing.T) {
	c, err := Parse(docsV0)
	urequire.NoError(t, err)
	uassert.Equal(t, uint64(0), c.Version())
	uassert.Equal(t, DagPB, c.Codec())
	uassert.Equal(t, SHA256, c.Multihash().Code())
	uassert.Equal(t, docsV0, c.String())

	v1 := c.V1()
	uassert.Equal(t, uint64(1), v1.Version())
	uassert.Equal(t, docsV1, v1.String())
	uassert.False(t, c == v1, "a CIDv0 and its V1 are different CIDs")
	uassert.True(t, v1.V1() == v1)
	uassert.Equal(t, string(c.Multihash().Bytes()), string(v1.Multihash().Bytes()))

	_, err = c.Encode('b')
	uassert.Error(t, err, "a CIDv0 has no multibase form")
}

func TestVerify(t *testing.T) {
	hello, _ := Parse(helloCID)
	uassert.NoError(t, hello.Verify([]byte("hello")))
	uassert.ErrorIs(t, hello.Verify([]byte("hellp")), ErrDigestMismatch)
	uassert.ErrorIs(t, hello.Verify(nil), ErrDigestMismatch)

	// The identity multihash inlines the content: raw "hi".
	identity, err := Parse("bafkqaatine")
	urequire.NoError(t, err)
	uassert.Equal(t, Identity, identity.Multihash().Code())
	uassert.NoError(t, identity.Verify([]byte("hi")))
	uassert.ErrorIs(t, identity.Verify([]byte("ho")), ErrDigestMismatch)
	uassert.ErrorIs(t, identity.Verify([]byte("hii")), ErrDigestMismatch)

	// keccak-256("") is c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470.
	keccak, err := Parse("bafkrwigf2jdadbxxem6je7t5wlomoa6a4ualmu6kqittw6723acf3bneoa")
	urequire.NoError(t, err)
	uassert.Equal(t, Keccak256, keccak.Multihash().Code())
	uassert.NoError(t, keccak.Verify(nil))
	uassert.ErrorIs(t, keccak.Verify([]byte("x")), ErrDigestMismatch)

	// blake2b-256 parses, but Verify does not implement it.
	mh, err := NewMultihash(0xb220, make([]byte, 32))
	urequire.NoError(t, err)
	blake, err := NewV1(Raw, mh)
	urequire.NoError(t, err)
	uassert.ErrorIs(t, blake.Verify(nil), ErrUnsupportedHash)

	uassert.Error(t, CID{}.Verify(nil))
}

// Root blocks of two files on IPFS: a CIDv0 file with dag-pb leaves and a
// CIDv1 file with raw leaves. Each root links to the file's chunks.
var (
	coinflipRoot = "122a0a221220ebba04a2f8c0faa60166fdc3ed576944ac37196c2493a29b5ce1" +
		"aca3c01b53e41200188e8010122a0a221220eb470dbb02d858b2e6e2b125fdb5" +
		"aad3c0cb8cb176dfdf1e5a2179d39345be821200188e8010122a0a221220a1a4" +
		"c371c79d98197ab34d5f469a3c17085ff25b835ac4fce988f1464e10facc1200" +
		"188e8010122a0a221220a8c6281f7ef49c6adebb576a5d2d6363670b37e6eea1" +
		"0d10827e28091152d6bb1200188e8010122a0a2212204c00f2cb0092f20c01db" +
		"76cba70326b52b278ba5cd30f1c00310a79826c7dc4d120018b3880e0a1a0802" +
		"18a5884e2080801020808010208080102080801020a5880e"
	fomo3dRoot = "122c0a2401551220cb7ad3340ed87b19c10730690917cf9e13316926e39b71e0" +
		"123fb7ae8702dc94120018808010122c0a2401551220928346826ea253359038" +
		"6c0db16fd365b4894f92033c1929a7b7ea0476c97f79120018808010122c0a24" +
		"015512208e26a99ce34036a2b461ac5b10a81a71b007b80b69c1f23d4ddc9a4b" +
		"aa4cf36e120018808010122c0a2401551220478360217bf37512202ca81123be" +
		"cd707e5a0d214187413febeda6cb0c81b5a8120018808010122c0a2401551220" +
		"1ce779d9f7ea3bfe7ed405fd0d3688a8c29a3e80beb124b6faef31959abda9c1" +
		"1200189f8b060a1a0802189f8b4620808010208080102080801020808010209f" +
		"8b06"
)

func TestVerifyRootBlocks(t *testing.T) {
	for _, tc := range []struct{ cid, block string }{
		{"QmVZxCKz3J5iMHrACNLfFWpfpfcKT2aSmZMfvZsCQnMWbf", coinflipRoot},
		{"bafybeidayyli6bpewkhgtwqpgubmo77kmgjn4r5zq2i7usoyadcmvynhhq", fomo3dRoot},
	} {
		c, err := Parse(tc.cid)
		urequire.NoError(t, err)
		block, err := hex.DecodeString(tc.block)
		urequire.NoError(t, err)

		uassert.Equal(t, DagPB, c.Codec())
		uassert.NoError(t, c.Verify(block), tc.cid)
		block[len(block)-1] ^= 1
		uassert.ErrorIs(t, c.Verify(block), ErrDigestMismatch, tc.cid)
	}
}

func TestDecode(t *testing.T) {
	for _, s := range []string{helloCID, docsV0, docsV1, "bafkqaatine"} {
		c, err := Parse(s)
		urequire.NoError(t, err)
		d, err := Decode(c.Bytes())
		urequire.NoError(t, err)
		uassert.True(t, d == c, s)
		uassert.Equal(t, string(c.Bytes()), c.KeyString())
	}
}

func TestDecodeFirst(t *testing.T) {
	for _, str := range []string{helloCID, docsV0, "bafkqaatine"} {
		c, _ := Parse(str)
		b := append(c.Bytes(), 0x01, 0x02, 0x03)

		got, n, err := DecodeFirst(b)
		urequire.NoError(t, err, str)
		uassert.True(t, got == c, str)
		uassert.Equal(t, len(c.Bytes()), n)

		_, _, err = DecodeFirst(b[:len(c.Bytes())-1])
		uassert.Error(t, err, "truncated "+str)
		_, err = Decode(b)
		uassert.Error(t, err, "trailing bytes after "+str)
	}
}

// The longest CID the package accepts must round-trip through every
// multibase and DecodeFirst, so no length cap is too tight.
func TestMaxCID(t *testing.T) {
	digest := []byte(strings.Repeat("\xff", 128))
	mh, err := NewMultihash(maxVarint, digest)
	urequire.NoError(t, err)
	c, err := NewV1(maxVarint, mh)
	urequire.NoError(t, err)
	uassert.Equal(t, 149, len(c.Bytes())) // literal: pins maxCIDLen

	for _, base := range []byte("bBzkfF") {
		s, err := c.Encode(base)
		urequire.NoError(t, err)
		got, err := Parse(s)
		urequire.NoError(t, err, s)
		uassert.True(t, got == c, s)
	}

	got, n, err := DecodeFirst(append(c.Bytes(), 1, 2))
	urequire.NoError(t, err)
	uassert.True(t, got == c)
	uassert.Equal(t, 149, n)

	// A 128-byte identity digest is the largest inline content.
	ident, err := NewMultihash(Identity, digest)
	urequire.NoError(t, err)
	ic, err := NewV1(Raw, ident)
	urequire.NoError(t, err)
	parsed, err := Parse(ic.String())
	urequire.NoError(t, err)
	uassert.NoError(t, parsed.Verify(digest))
}

func TestParseRejects(t *testing.T) {
	hex129 := "f0155008101" + strings.Repeat("61", 129)
	for _, s := range []string{
		"",
		"b",
		"b" + strings.Repeat("a", 240),       // longer than any CID in base32
		"z" + strings.Repeat("2", 205),       // or base58btc
		"k" + strings.Repeat("2", 232),       // or base36
		"x" + helloCID[1:],                   // unknown multibase
		helloCID[:len(helloCID)-1] + "Q",     // case mixing
		helloCID[:10] + "\n" + helloCID[10:], // whitespace
		helloCID[:len(helloCID)-1] + "r",     // non-zero leftover bits
		helloCID[:len(helloCID)-1],           // impossible base32 length
		"z0OIl",                              // not in base58btc
		"f015512202Cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", // case mixing
		"f0155122", // odd hex length
		"babkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44yegnrjhc4yeq",   // version 0 in CIDv1 form
		"bajkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44yegnrjhc4yeq",   // version 2
		"bahkqaeraftze3os7wcrq4jxihmvmlopctynrmhs4d6tuexttaqzwfe4ltasa", // non-minimal codec varint
		"bafkreaa",                     // sha2-256 digest of length 0
		helloCID + "aa",                // trailing byte
		"z" + docsV0,                   // a CIDv0 in multibase form
		docsV0[:45] + "0",              // CIDv0 with a character outside base58btc
		docsV0[:45],                    // too short for a CIDv0
		"Qm" + strings.Repeat("1", 44), // base58btc, but 0x12 0x1e...: not a sha2-256 multihash
		hex129,                         // identity digest over 128 bytes
	} {
		_, err := Parse(s)
		uassert.Error(t, err, s)
	}
}

func TestNewMultihash(t *testing.T) {
	mh, err := NewMultihash(SHA256, make([]byte, 32))
	urequire.NoError(t, err)
	uassert.Equal(t, SHA256, mh.Code())
	uassert.Equal(t, string(make([]byte, 32)), string(mh.Digest()))

	for _, tc := range []struct {
		code uint64
		size int
	}{
		{SHA256, 31},
		{Keccak256, 33},
		{Identity, maxDigestLen + 1},
		{1 << 63, 0},
	} {
		_, err := NewMultihash(tc.code, make([]byte, tc.size))
		uassert.Error(t, err)
	}
}

func TestZeroValue(t *testing.T) {
	var c CID
	uassert.False(t, c.Defined())
	uassert.Equal(t, "", c.String())
	uassert.False(t, c.V1().Defined())
	_, err := c.Encode('b')
	uassert.Error(t, err)

	_, err = NewV1(Raw, Multihash{})
	uassert.Error(t, err)
}

func TestUvarint(t *testing.T) {
	for _, tc := range []struct {
		in   []byte
		want uint64
		n    int
	}{
		{[]byte{0x00}, 0, 1},
		{[]byte{0x7f, 0xff}, 127, 1},
		{[]byte{0x80, 0x01}, 128, 2},
		{[]byte{0xa9, 0x02}, 0x129, 2},
		{[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, maxVarint, 9},
	} {
		got, n, err := uvarint(string(tc.in))
		urequire.NoError(t, err)
		uassert.Equal(t, tc.want, got)
		uassert.Equal(t, tc.n, n)
		uassert.Equal(t, string(tc.in[:n]), string(appendUvarint(nil, tc.want)))
	}

	for _, in := range [][]byte{
		nil,
		{0x80},       // truncated
		{0x80, 0x00}, // non-minimal zero
		{0xff, 0x80, 0x00},
		{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, // 10 bytes
	} {
		_, _, err := uvarint(string(in))
		uassert.Error(t, err)
	}
}
