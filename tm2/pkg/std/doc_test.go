package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignDoc_GetSignaturePayload(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name string
		doc  SignDoc
	}{
		{
			"empty sign doc",
			SignDoc{},
		},
		{
			"non-empty sign doc",
			SignDoc{
				ChainID:       "dummy",
				AccountNumber: 10,
				Sequence:      20,
				Fee: Fee{
					GasFee:    NewCoin("ugnot", 10),
					GasWanted: 10,
				},
				Msgs: []Msg{},
				Memo: "totally valid transaction",
			},
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			signPayload, err := GetSignaturePayload(testCase.doc)
			require.NoError(t, err)

			assert.NotNil(t, signPayload)
			assert.NotEmpty(t, signPayload)
		})
	}
}
