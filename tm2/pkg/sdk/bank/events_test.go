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

	want := TransferEvent{
		From:   "g1sender",
		To:     "g1recipient",
		Amount: std.NewCoins(std.NewCoin("ugnot", 5)),
	}
	bz, err := amino.Marshal(abci.ResponseDeliverTx{
		ResponseBase: abci.ResponseBase{Events: []abci.Event{want}},
	})
	require.NoError(t, err)

	var got abci.ResponseDeliverTx
	require.NoError(t, amino.Unmarshal(bz, &got))
	require.Equal(t, []abci.Event{want}, got.Events)
	require.Contains(t, string(amino.MustMarshalJSON(got)),
		`{"@type":"/bank.TransferEvent","from":"g1sender","to":"g1recipient","amount":"5ugnot"}`)
}
