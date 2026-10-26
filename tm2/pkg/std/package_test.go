package std_test

import (
	"encoding/json"
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
			orig := std.Signature{Signature: []byte("sig"), SessionAddr: std.SessionAddress(tc.sessionAddr)}

			bz, err := amino.Marshal(orig)
			require.NoError(t, err)

			if tc.wantMasterSig {
				assert.Equal(t, []byte{0x12, 0x03, 's', 'i', 'g'}, bz, "a zero SessionAddr must have no wire presence")

				// Chains produced before this change wrote the zero address out as
				// its bech32 string; those bytes must keep decoding to a master
				// signature.
				written := append([]byte{0x12, 0x03, 's', 'i', 'g', 0x1a, 0x28}, []byte(crypto.Address{}.String())...)
				var legacy std.Signature
				require.NoError(t, amino.Unmarshal(written, &legacy))
				assert.True(t, legacy.SessionAddr.IsZero(), "the written-out zero address must still decode as a master signature")
			}

			var got std.Signature
			require.NoError(t, amino.Unmarshal(bz, &got))

			assert.Equal(t, tc.wantMasterSig, got.SessionAddr.IsZero(),
				"round-tripped Signature SessionAddr classification mismatch: got=%v", got.SessionAddr)
			if !tc.wantMasterSig {
				assert.Equal(t, tc.sessionAddr, got.SessionAddr.Address())
			}
		})

		t.Run(tc.name+"/json", func(t *testing.T) {
			t.Parallel()
			orig := std.Signature{Signature: []byte("sig"), SessionAddr: std.SessionAddress(tc.sessionAddr)}

			// encoding/json, used by tools that import std, must agree with the
			// amino form: a bech32 string, "" when zero.
			plain, err := json.Marshal(orig.SessionAddr)
			require.NoError(t, err)
			wantJSON := `"` + tc.sessionAddr.String() + `"`
			if tc.wantMasterSig {
				wantJSON = `""`
			}
			assert.Equal(t, wantJSON, string(plain))
			var back std.SessionAddress
			require.NoError(t, json.Unmarshal(plain, &back))
			assert.Equal(t, tc.sessionAddr, back.Address())

			bz, err := amino.MarshalJSON(orig)
			require.NoError(t, err)

			var got std.Signature
			require.NoError(t, amino.UnmarshalJSON(bz, &got))

			assert.Equal(t, tc.wantMasterSig, got.SessionAddr.IsZero(),
				"round-tripped Signature SessionAddr classification mismatch: got=%v", got.SessionAddr)
			if !tc.wantMasterSig {
				assert.Equal(t, tc.sessionAddr, got.SessionAddr.Address())
			}
		})
	}
}
