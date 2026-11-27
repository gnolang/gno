package gnoland

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A transaction's `any` fields decode into any registered type before its
// signature or fee is checked, so a number that expands while decoding costs
// the sender nothing. CheckTx must refuse such a transaction at decode.
func TestCheckTxRefusesExpandingBigdec(t *testing.T) {
	t.Parallel()

	const (
		placeholder = "12345678" // Same length, so no length prefix changes.
		count       = 1_000      // A full-size transaction holds about 22,000.
	)

	list := make([]gnolang.TypedValue, count)
	for i := range list {
		list[i] = gnolang.TypedValue{V: gnolang.BigdecValue{V: big.NewRat(12345678, 1)}}
	}
	bz, err := amino.Marshal(std.Tx{
		Msgs: []std.Msg{vm.MsgAddPackage{
			Creator: crypto.Address{1},
			Package: &std.MemPackage{
				Name:  "x",
				Path:  "gno.land/r/x/x",
				Files: []*std.MemFile{{Name: "x.gno", Body: "package x\n"}},
				Type:  &gnolang.ArrayValue{List: list},
			},
		}},
		Fee: std.NewFee(1, std.NewCoin("ugnot", 1)),
	})
	require.NoError(t, err)
	require.Less(t, len(bz), 1_000_000)
	require.Equal(t, count, bytes.Count(bz, []byte(placeholder)))

	// Without the placeholder swap the transaction decodes, so what is refused
	// below is the exponent and not the shape of the transaction.
	var tx std.Tx
	require.NoError(t, amino.Unmarshal(bz, &tx))

	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)
	res := bapp.InitChain(abci.RequestInitChain{
		ChainID:  "bigdec-test",
		AppState: gnoGenesisState(t),
		ConsensusParams: &abci.ConsensusParams{
			Block:     defaultBlockParams(),
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
		},
	})
	require.True(t, res.IsOK(), "InitChain: %v", res.Error)
	bapp.Commit()

	for _, attack := range []string{"1e999999", "1p999999"} {
		attackBz := bytes.ReplaceAll(bz, []byte(placeholder), []byte(attack))
		res := bapp.CheckTx(abci.RequestCheckTx{Tx: attackBz})
		assert.IsType(t, std.TxDecodeError{}, res.Error, "%s: %s", attack, res.Log)
	}
}
