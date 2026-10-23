package std_test

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAminoBaseAccount(t *testing.T) {
	b := []byte(`{
    "address": "g1x90eh5ejc22548hjqznm2egyvn8ny36lqu460f",
    "coins": "4200000ugnot",
    "public_key": {
      "@type": "/tm.PubKeySecp256k1",
      "value": "AwMzujfppqEi8lozMVD8ORENUR8SIE06VLNP8FGL0aQ2"
    },
    "account_number": "159",
    "sequence": "33"
}`)
	acc := std.BaseAccount{}

	err := amino.UnmarshalJSON(b, &acc)
	require.NoError(t, err)
}

func TestAminoGasPrice(t *testing.T) {
	gp := std.GasPrice{
		Gas: 100,
		Price: std.Coin{
			Denom:  "token",
			Amount: 10,
		},
	}
	// Binary
	bz, err := amino.Marshal(gp)
	require.NoError(t, err)
	err = amino.Unmarshal(bz, &gp)
	require.NoError(t, err)

	// JSON
	bz, err = amino.MarshalJSON(gp)
	require.NoError(t, err)

	err = amino.UnmarshalJSON(bz, &gp)
	require.NoError(t, err)

	bz = []byte(`{
				"gas": "10",
				"price": "100token"
		}`)
	err = amino.UnmarshalJSON(bz, &gp)
	require.NoError(t, err)
}

func TestAminoCoin(t *testing.T) {
	coin := std.Coin{
		Denom:  "token",
		Amount: 10,
	}

	// Binary
	bz, err := amino.Marshal(coin)
	require.NoError(t, err)

	err = amino.Unmarshal(bz, &coin)
	require.NoError(t, err)

	// JSON
	bz, err = amino.MarshalJSON(coin)
	require.NoError(t, err)
	err = amino.UnmarshalJSON(bz, &coin)
	require.NoError(t, err)

	bz = []byte(`"10token"`)
	err = amino.UnmarshalJSON(bz, &coin)
	require.NoError(t, err)
}

// TestAminoSignatureSessionAddrRoundTrip locks in the ante-handler
// invariant that a Signature with zero SessionAddr round-trips as a
// zero SessionAddr (i.e. stays classified as a master signature).
//
// The ante handler treats a signature as session-signed iff
// `!sig.SessionAddr.IsZero()`. This test will break loudly if amino
// ever decodes an absent/zero session_addr field as a non-zero address.
func TestAminoSignatureSessionAddrRoundTrip(t *testing.T) {
	t.Parallel()

	session := crypto.AddressFromPreimage([]byte("some-session-key"))

	cases := []struct {
		name          string
		sessionAddr   crypto.Address
		wantMasterSig bool // should the round-tripped sig look like a master sig?
	}{
		{"zero SessionAddr", crypto.Address{}, true},
		{"populated SessionAddr", session, false},
	}

	for _, tc := range cases {
		t.Run(tc.name+"/binary", func(t *testing.T) {
			t.Parallel()
			orig := std.Signature{Signature: []byte("sig"), SessionAddr: tc.sessionAddr}

			bz, err := amino.Marshal(orig)
			require.NoError(t, err)

			var got std.Signature
			require.NoError(t, amino.Unmarshal(bz, &got))

			assert.Equal(t, tc.wantMasterSig, got.SessionAddr.IsZero(),
				"round-tripped Signature SessionAddr classification mismatch: got=%v", got.SessionAddr)
			if !tc.wantMasterSig {
				assert.Equal(t, tc.sessionAddr, got.SessionAddr)
			}
		})

		t.Run(tc.name+"/json", func(t *testing.T) {
			t.Parallel()
			orig := std.Signature{Signature: []byte("sig"), SessionAddr: tc.sessionAddr}

			bz, err := amino.MarshalJSON(orig)
			require.NoError(t, err)

			var got std.Signature
			require.NoError(t, amino.UnmarshalJSON(bz, &got))

			assert.Equal(t, tc.wantMasterSig, got.SessionAddr.IsZero(),
				"round-tripped Signature SessionAddr classification mismatch: got=%v", got.SessionAddr)
			if !tc.wantMasterSig {
				assert.Equal(t, tc.sessionAddr, got.SessionAddr)
			}
		})
	}
}

// TestAminoSignatureZeroSessionAddrWire pins what a master-key signature
// writes for its zero SessionAddr: nothing, in amino binary and JSON alike,
// through amino:"skip_zero". Without the tag amino binary wrote field 3 as
// the zero address's bech32 string, judging emptiness by MarshalAmino's
// repr; signatures encoded that way must still decode, through both codecs,
// to a zero SessionAddr, so they stay master-key signatures. If any of this
// changes, update SessionAddr's comment in doc.go with it.
func TestAminoSignatureZeroSessionAddrWire(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(std.Package)
	cdc.Seal()

	sig := std.Signature{Signature: []byte("sig")}
	bz, err := cdc.MarshalReflect(sig)
	require.NoError(t, err)
	// Field 2 (0x12) with the 3 signature bytes, and no field 3.
	assert.Equal(t, []byte("\x12\x03sig"), bz)

	js, err := cdc.JSONMarshal(sig)
	require.NoError(t, err)
	assert.NotContains(t, string(js), "session_addr")

	// The encoding before the tag: field 3, length-delimited (0x1a), holding
	// the zero address's bech32 string, 42 bytes in all.
	zero := crypto.Address{}.String()
	legacy := append([]byte("\x12\x03sig\x1a"), byte(len(zero)))
	legacy = append(legacy, zero...)
	require.Len(t, legacy, len(bz)+42)
	var viaReflect, viaBinary2 std.Signature
	require.NoError(t, cdc.UnmarshalReflect(legacy, &viaReflect))
	require.NoError(t, viaBinary2.UnmarshalBinary2(cdc, legacy, 0))
	assert.Equal(t, sig, viaReflect)
	assert.Equal(t, sig, viaBinary2)

	// And its JSON, which spelled the zero address out where a client
	// wrote it so.
	var viaJSON std.Signature
	require.NoError(t, cdc.JSONUnmarshal([]byte(`{"signature":"c2ln","session_addr":"`+zero+`"}`), &viaJSON))
	assert.Equal(t, sig, viaJSON)

	// An all-zero signature, which gnoclient puts in a Tx for each signer
	// before signing, is now an empty list element (field 3, length 0); the
	// "Tx/placeholder-signatures" parity case checks that decoding keeps both.
	bz, err = cdc.MarshalReflect(std.Tx{Signatures: []std.Signature{{}, {}}})
	require.NoError(t, err)
	assert.Equal(t, []byte{0x1a, 0x00, 0x1a, 0x00}, bz)
}
