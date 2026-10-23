package gnoland

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// legacySignature and legacyTx are std.Signature and std.Tx as amino
// encoded them before SessionAddr took amino:"skip_zero": the same fields
// and tags, so a zero SessionAddr goes on the wire as the zero address's
// bech32 string. Wallets built before the change still send this encoding.
type legacySignature struct {
	PubKey      crypto.PubKey  `json:"pub_key" yaml:"pub_key"`
	Signature   []byte         `json:"signature" yaml:"signature"`
	SessionAddr crypto.Address `json:"session_addr,omitempty" yaml:"session_addr,omitempty"`
}

type legacyTx struct {
	Msgs       []std.Msg         `json:"msg" yaml:"msg"`
	Fee        std.Fee           `json:"fee" yaml:"fee"`
	Signatures []legacySignature `json:"signatures" yaml:"signatures"`
	Memo       string            `json:"memo" yaml:"memo"`
}

func legacyTxBytes(t *testing.T, tx std.Tx) []byte {
	t.Helper()

	ltx := legacyTx{Msgs: tx.Msgs, Fee: tx.Fee, Memo: tx.Memo}
	for _, sig := range tx.Signatures {
		ltx.Signatures = append(ltx.Signatures, legacySignature(sig))
	}
	bz, err := amino.Marshal(ltx)
	require.NoError(t, err)
	return bz
}

// TestDeliverTxZeroSessionAddr: a master-key signature leaves its zero
// SessionAddr off the wire, and the chain still takes the same transaction
// encoded the old way, as wallets built before the change keep sending it.
// Delivered to two identical chains, the two encodings do the same thing
// and reach the same app hash; the old one only costs more gas, for its
// extra bytes.
func TestDeliverTxZeroSessionAddr(t *testing.T) {
	t.Parallel()

	key := getDummyKey(t)
	from := key.PubKey().Address()
	tx := createAndSignTx(t, []std.Msg{
		bank.NewMsgSend(from, crypto.AddressFromPreimage([]byte("to")), std.NewCoins(std.NewCoin("ugnot", 1))),
	}, "test-chain", key)

	newBz := amino.MustMarshal(tx)
	oldBz := legacyTxBytes(t, tx)
	// The zero address's field is 42 bytes, and it takes this secp256k1
	// signature from 126 bytes to 168, past the 127 a one-byte length
	// prefix holds.
	require.Len(t, oldBz, len(newBz)+43)

	state := DefaultGenState()
	state.Balances = []Balance{{
		Address: from,
		Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
	}}
	deliver := func(txBytes []byte) (abci.ResponseDeliverTx, []byte) {
		t.Helper()

		app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
		require.NoError(t, err)
		initRes := app.InitChain(abci.RequestInitChain{
			ChainID:         "test-chain",
			ConsensusParams: &abci.ConsensusParams{Block: defaultBlockParams()},
			AppState:        state,
		})
		require.True(t, initRes.IsOK(), "InitChain failed: %v", initRes.Error)
		app.BeginBlock(abci.RequestBeginBlock{
			Header: &bft.Header{ChainID: "test-chain", Height: 1},
		})
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		app.EndBlock(abci.RequestEndBlock{Height: 1})
		return res, app.Commit().Data
	}
	newRes, newHash := deliver(newBz)
	oldRes, oldHash := deliver(oldBz)
	require.True(t, newRes.IsOK(), "new encoding: %+v", newRes)
	require.True(t, oldRes.IsOK(), "old encoding: %+v", oldRes)
	assert.Equal(t, newHash, oldHash)

	// The signature was checked, not skipped: the old encoding with a
	// signature byte flipped is refused.
	tampered := tx
	tampered.Signatures = []std.Signature{tx.Signatures[0]}
	tampered.Signatures[0].Signature = bytes.Clone(tx.Signatures[0].Signature)
	tampered.Signatures[0].Signature[0] ^= 1
	badRes, _ := deliver(legacyTxBytes(t, tampered))
	assert.IsType(t, std.UnauthorizedError{}, badRes.Error, "tampered: %+v", badRes)

	// The ante handler charges tx_size_cost_per_byte for every byte, 43 × 10
	// = 430 gas at the default.
	perByte := state.Auth.Params.TxSizeCostPerByte
	assert.Equal(t, int64(len(oldBz)-len(newBz))*perByte, oldRes.GasUsed-newRes.GasUsed)
}
