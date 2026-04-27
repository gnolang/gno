package std_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestCodecParity_Std(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(ed25519.Package)
	cdc.RegisterPackage(secp256k1.Package)
	cdc.RegisterPackage(std.Package)
	cdc.Seal()

	addr := crypto.AddressFromPreimage([]byte("std-parity"))
	pk := ed25519.PubKeyEd25519{0x01, 0x02, 0x03}

	cases := []struct {
		name string
		v    any
	}{
		// Coin + GasPrice: plain scalars with edge values.
		{"Coin", &std.Coin{Denom: "ugnot", Amount: math.MaxInt64}},
		// Note: negative coin amounts are rejected by Coin.UnmarshalAmino,
		// so they can't round-trip. Not included.
		{"Coin/zero", &std.Coin{}},
		{"GasPrice", &std.GasPrice{Gas: 1000, Price: std.Coin{Denom: "ugnot", Amount: 1}}},
		{"Fee", &std.Fee{GasWanted: 200000, GasFee: std.Coin{Denom: "ugnot", Amount: 5000}}},

		// BaseAccount: has Address (AminoMarshaler), PubKey (interface).
		{"BaseAccount/zero-address", &std.BaseAccount{
			// Address left zero — exercises the repr-zeroness surface.
			PubKey:        pk,
			AccountNumber: 1,
			Sequence:      42,
		}},
		{"BaseAccount/full", &std.BaseAccount{
			Address: addr,
			// Coins are sorted on roundtrip — keep test input in canonical
			// (alphabetical by Denom) order to preserve strict DeepEqual.
			Coins:         std.Coins{{Denom: "ugno", Amount: 200}, {Denom: "ugnot", Amount: 100}},
			PubKey:        pk,
			AccountNumber: 7,
			Sequence:      9,
		}},

		// BaseSessionAccount: BaseAccount embedded + extra fields + SpendLimit
		// (a Coins slice under json omitempty).
		{"BaseSessionAccount", &std.BaseSessionAccount{
			BaseAccount: std.BaseAccount{
				Address:       addr,
				PubKey:        pk,
				AccountNumber: 11,
			},
			MasterAddress: crypto.AddressFromPreimage([]byte("master")),
			ExpiresAt:     1700000000,
			SpendLimit:    std.Coins{{Denom: "ugnot", Amount: 500}},
			SpendPeriod:   3600,
		}},

		// Signature: has PubKey (interface) + Signature bytes.
		{"Signature", &std.Signature{
			PubKey:    pk,
			Signature: []byte{0xde, 0xad, 0xbe, 0xef},
		}},
		{"Signature/nil-pubkey", &std.Signature{
			// PubKey left nil — interface-is-nil case.
			Signature: []byte{0x01, 0x02},
		}},

		// MemFile + MemPackage: nested pointer slice.
		{"MemFile", &std.MemFile{Name: "foo.gno", Body: "package foo\n"}},
		{"MemPackage", &std.MemPackage{
			Name: "foo",
			Path: "gno.land/r/test/foo",
			Files: []*std.MemFile{
				{Name: "a.gno", Body: "package foo\n"},
				{Name: "b.gno", Body: "package foo\nfunc Bar() {}\n"},
			},
		}},

		// One representative error type — all the std errors share a struct
		// shape (wraps abciError) so this is a class-level smoke test.
		{"OutOfGasError", &std.OutOfGasError{}},

		// Tx: the headline type. Interface slice of Msgs. No msgs here
		// because Msg concrete types live in other packages.
		{"Tx/empty-msgs", &std.Tx{
			Msgs:       nil,
			Fee:        std.Fee{GasWanted: 1000, GasFee: std.Coin{Denom: "ugnot", Amount: 10}},
			Signatures: []std.Signature{{PubKey: pk, Signature: []byte{0x01}}},
			Memo:       "hello",
		}},
		// String field with embedded NUL bytes — amino's length-prefix
		// encoding must be binary-safe.
		{"Tx/memo-with-nul", &std.Tx{
			Msgs: nil,
			Fee:  std.Fee{GasWanted: 1000, GasFee: std.Coin{Denom: "ugnot", Amount: 10}},
			Memo: "hel\x00lo\x00world",
		}},
		// Single-byte Signature of 0x00 — exercises the wire-format
		// path for a byte slice whose ENTIRE content is a single null
		// byte (distinct from the empty []byte case).
		{"Signature/single-zero-byte", &std.Signature{
			PubKey:    pk,
			Signature: []byte{0x00},
		}},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
