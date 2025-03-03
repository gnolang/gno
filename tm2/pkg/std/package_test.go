package std_test

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
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
