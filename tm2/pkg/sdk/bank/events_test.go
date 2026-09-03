package bank

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestTransferEventAminoRoundTrip(t *testing.T) {
	t.Parallel()

	from := crypto.AddressFromPreimage([]byte("sender"))
	to := crypto.AddressFromPreimage([]byte("recipient"))
	coins := std.NewCoins(std.NewCoin("ugnot", 5))
	tests := []struct {
		name string
		want abci.Event
		json string
	}{
		{
			name: "transfer",
			want: TransferEvent{From: "g1sender", To: "g1recipient", Coins: coins},
			json: `{"@type":"/bank.TransferEvent","from":"g1sender","to":"g1recipient","coins":"5ugnot"}`,
		},
		{
			name: "multisend",
			want: MultiTransferEvent{
				Inputs:  []Input{{Address: from, Coins: coins}},
				Outputs: []Output{{Address: to, Coins: coins}},
			},
			json: fmt.Sprintf(`{"@type":"/bank.MultiTransferEvent","inputs":[{"address":"%s","coins":"5ugnot"}],"outputs":[{"address":"%s","coins":"5ugnot"}]}`,
				from, to),
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
