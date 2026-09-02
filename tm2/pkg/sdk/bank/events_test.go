package bank

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestTransferEventAminoRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want TransferEvent
		json string
	}{
		{
			name: "transfer",
			want: TransferEvent{From: "g1sender", To: "g1recipient", Amount: std.NewCoins(std.NewCoin("ugnot", 5))},
			json: `{"@type":"/bank.TransferEvent","from":"g1sender","to":"g1recipient","amount":"5ugnot"}`,
		},
		{
			name: "mint-shaped",
			want: TransferEvent{To: "g1recipient", Amount: std.NewCoins(std.NewCoin("ugnot", 5))},
			json: `{"@type":"/bank.TransferEvent","from":"","to":"g1recipient","amount":"5ugnot"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			bz, err := amino.Marshal(abci.ResponseDeliverTx{
				ResponseBase: abci.ResponseBase{Events: []abci.Event{tc.want}},
			})
			require.NoError(t, err)

			var got abci.ResponseDeliverTx
			require.NoError(t, amino.Unmarshal(bz, &got))
			require.Equal(t, []abci.Event{tc.want}, got.Events)
			require.Contains(t, string(amino.MustMarshalJSON(got)), tc.json)
		})
	}
}
