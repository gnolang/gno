package gnoland

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// BenchmarkDeliverTxBankSend measures one DeliverTx of a bank send through the
// full gno.land app: decode, the ante handler and the bank message handler.
// Every iteration delivers a different pre-signed tx (consecutive sequences)
// in the same block.
func BenchmarkDeliverTxBankSend(b *testing.B) {
	to := crypto.AddressFromPreimage([]byte("to"))
	benchDeliverTx(b, nil, func(from crypto.Address) std.Msg {
		return bank.NewMsgSend(from, to, std.NewCoins(std.NewCoin("ugnot", 1)))
	})
}

// BenchmarkDeliverTxMsgCall is BenchmarkDeliverTxBankSend with a realm call,
// whose message handler runs the VM.
func BenchmarkDeliverTxMsgCall(b *testing.B) {
	const pkgPath = "gno.land/r/bench/counter"
	files := []*std.MemFile{
		{Name: "counter.gno", Body: "package counter\n\nvar n int\n\nfunc Incr(cur realm) int {\n\tn++\n\treturn n\n}\n"},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}
	benchDeliverTx(b, func(from crypto.Address) std.Tx {
		return std.Tx{
			Msgs:       []std.Msg{vm.NewMsgAddPackage(from, pkgPath, files)},
			Fee:        std.NewFee(100_000_000, std.NewCoin("ugnot", 100_000_000)),
			Signatures: []std.Signature{{}},
		}
	}, func(from crypto.Address) std.Msg {
		return vm.NewMsgCall(from, nil, pkgPath, "Incr", nil)
	})
}

// benchDeliverTx commits height 1, with genesisTx in genesis when it is not
// nil, then delivers one pre-signed tx carrying msg per iteration at height 2.
func benchDeliverTx(b *testing.B, genesisTx func(crypto.Address) std.Tx, msg func(crypto.Address) std.Msg) {
	b.Helper()

	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	if err != nil {
		b.Fatal(err)
	}
	key := secp256k1.GenPrivKey()
	from := key.PubKey().Address()
	state := DefaultGenState()
	state.Balances = []Balance{{Address: from, Amount: std.NewCoins(std.NewCoin("ugnot", 1_000_000_000_000))}}
	if genesisTx != nil {
		state.Txs = []TxWithMetadata{{Tx: genesisTx(from)}}
	}
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

	txs := make([][]byte, b.N)
	for i := range txs {
		tx := std.Tx{
			Msgs: []std.Msg{msg(from)},
			Fee:  std.NewFee(10_000_000, std.NewCoin("ugnot", 10_000_000)),
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

	app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: "bench", Height: 2}})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if res := app.DeliverTx(abci.RequestDeliverTx{Tx: txs[i]}); !res.IsOK() {
			b.Fatal(res.Log)
		}
	}
}
