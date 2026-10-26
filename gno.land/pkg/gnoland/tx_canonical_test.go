package gnoland

import (
	"slices"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A signed tx re-encoded non-canonically decodes to the same tx, so its
// signatures still verify, but it hashes differently; before the canonical
// check the chain included it under a hash the sender never computed. Both
// variants must now fail to decode in CheckTx and DeliverTx.
func TestTxNonCanonicalEncodingRejected(t *testing.T) {
	t.Parallel()

	const chainID = "test-chain"
	key := getDummyKey(t)
	to := crypto.AddressFromPreimage([]byte("to"))

	gs := DefaultGenState()
	gs.Balances = []Balance{{
		Address: key.PubKey().Address(),
		Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
	}}
	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	app.InitChain(abci.RequestInitChain{
		ChainID: chainID,
		ConsensusParams: &abci.ConsensusParams{
			Block:     defaultBlockParams(),
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
		},
		AppState: gs,
	})
	app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: chainID, Height: 1}})

	msg := bank.NewMsgSend(key.PubKey().Address(), to, std.NewCoins(std.NewCoin("ugnot", 1000)))
	tx := createAndSignTx(t, []std.Msg{msg}, chainID, key)
	require.Empty(t, tx.Memo)
	canonical, err := amino.Marshal(tx)
	require.NoError(t, err)
	require.Less(t, canonical[0], byte(0x80))

	mutations := map[string][]byte{
		// Memo is field 4 and empty, so the encoder omitted it; write it out.
		"explicit empty memo": slices.Concat(canonical, []byte{0x22, 0x00}),
		// The first tag as a two-byte varint.
		"non-minimal tag varint": slices.Concat([]byte{canonical[0] | 0x80, 0x00}, canonical[1:]),
	}
	for name, bz := range mutations {
		var decoded std.Tx
		require.NoError(t, amino.Unmarshal(bz, &decoded), name)
		reencoded, err := amino.Marshal(decoded)
		require.NoError(t, err)
		require.Equal(t, canonical, reencoded, "mutation must decode to the same signed tx")

		check := app.CheckTx(abci.RequestCheckTx{Tx: bz})
		require.False(t, check.IsOK(), name)
		assert.Equal(t, std.TxDecodeError{}, check.Error, name)
		assert.Contains(t, check.Log, "non-canonical encoding", name)

		deliver := app.DeliverTx(abci.RequestDeliverTx{Tx: bz})
		require.False(t, deliver.IsOK(), name)
		assert.Equal(t, std.TxDecodeError{}, deliver.Error, name)
		assert.Contains(t, deliver.Log, "non-canonical encoding", name)
	}

	// The sender's own bytes are still accepted.
	deliver := app.DeliverTx(abci.RequestDeliverTx{Tx: canonical})
	require.True(t, deliver.IsOK(), deliver.Log)
}
