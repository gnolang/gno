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
// invariant that a Signature with no SessionAddr (nil pointer) round-trips
// through amino without becoming a non-nil non-zero pointer.
//
// The ante handler treats a signature as session-signed iff
// `sig.SessionAddr != nil && !sig.SessionAddr.IsZero()`. Amino may decode
// an absent *crypto.Address field as a non-nil pointer to the zero address
// — the IsZero() half of the check guards against that. This test will
// break loudly if either half starts returning a non-zero, non-nil address
// on round-trip.
func TestAminoSignatureSessionAddrRoundTrip(t *testing.T) {
	t.Parallel()

	session := crypto.AddressFromPreimage([]byte("some-session-key"))

	cases := []struct {
		name          string
		sessionAddr   *crypto.Address
		wantMasterSig bool // should the round-tripped sig look like a master sig?
	}{
		{"nil SessionAddr", nil, true},
		{"zero-value SessionAddr pointer", &crypto.Address{}, true},
		{"populated SessionAddr", &session, false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name+"/binary", func(t *testing.T) {
			t.Parallel()
			orig := std.Signature{Signature: []byte("sig"), SessionAddr: tc.sessionAddr}

			bz, err := amino.Marshal(orig)
			require.NoError(t, err)

			var got std.Signature
			require.NoError(t, amino.Unmarshal(bz, &got))

			isMasterSig := got.SessionAddr == nil || got.SessionAddr.IsZero()
			assert.Equal(t, tc.wantMasterSig, isMasterSig,
				"round-tripped Signature SessionAddr classification mismatch: got=%v", got.SessionAddr)
			if !tc.wantMasterSig {
				assert.Equal(t, *tc.sessionAddr, *got.SessionAddr)
			}
		})

		t.Run(tc.name+"/json", func(t *testing.T) {
			t.Parallel()
			orig := std.Signature{Signature: []byte("sig"), SessionAddr: tc.sessionAddr}

			bz, err := amino.MarshalJSON(orig)
			require.NoError(t, err)

			var got std.Signature
			require.NoError(t, amino.UnmarshalJSON(bz, &got))

			isMasterSig := got.SessionAddr == nil || got.SessionAddr.IsZero()
			assert.Equal(t, tc.wantMasterSig, isMasterSig,
				"round-tripped Signature SessionAddr classification mismatch: got=%v", got.SessionAddr)
			if !tc.wantMasterSig {
				assert.Equal(t, *tc.sessionAddr, *got.SessionAddr)
			}
		})
	}
}
