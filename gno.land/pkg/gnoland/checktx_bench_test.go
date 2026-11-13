package gnoland

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// BenchmarkCheckTxRecheck measures one mempool recheck of a pending
// transaction through the full gno.land app: decode, the gno.land ante
// wrapper, the auth ante handler. Every iteration checks a different
// pre-signed tx (consecutive sequences) against the same committed height.
func BenchmarkCheckTxRecheck(b *testing.B) {
	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	if err != nil {
		b.Fatal(err)
	}
	key := secp256k1.GenPrivKey()
	from := key.PubKey().Address()
	state := DefaultGenState()
	state.Balances = []Balance{{Address: from, Amount: std.NewCoins(std.NewCoin("ugnot", 1_000_000_000_000))}}
	if res := app.InitChain(abci.RequestInitChain{
		ChainID: "bench",
		ConsensusParams: &abci.ConsensusParams{Block: &abci.BlockParams{
			MaxTxBytes: 1_000_000, MaxDataBytes: 2_000_000, MaxGas: 100_000_000_000, TimeIotaMS: 100,
		}},
		AppState: state,
	}); !res.IsOK() {
		b.Fatal(res.Error)
	}
	app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: "bench", Height: 1}})
	app.EndBlock(abci.RequestEndBlock{Height: 1})
	app.Commit()

	to := crypto.AddressFromPreimage([]byte("to"))
	txs := make([][]byte, b.N)
	for i := range txs {
		tx := std.Tx{
			Msgs: []std.Msg{bank.NewMsgSend(from, to, std.NewCoins(std.NewCoin("ugnot", 1)))},
			Fee:  std.NewFee(1_000_000, std.NewCoin("ugnot", 1_000_000)),
		}
		sb, err := tx.GetSignBytes("bench", 0, uint64(i))
		if err != nil {
			b.Fatal(err)
		}
		sig, err := key.Sign(sb)
		if err != nil {
			b.Fatal(err)
		}
		tx.Signatures = []std.Signature{{PubKey: key.PubKey(), Signature: sig}}
		txs[i] = amino.MustMarshal(tx)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if res := app.CheckTx(abci.RequestCheckTx{Tx: txs[i], Type: abci.CheckTxTypeRecheck}); !res.IsOK() {
			b.Fatal(res.Log)
		}
	}
}
