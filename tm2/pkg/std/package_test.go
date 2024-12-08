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
