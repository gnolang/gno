package gnoland

// Benchmarks for the execution budget that 0-fee realm sponsorship exposes at
// transaction admission.
//
// Question: with a validator opted in, how much real CPU can an attacker force
// per CheckTx as a function of Block.MaxGasCreditPerTx?
//
// Methodology notes that matter for trusting the numbers:
//
//   - Real gas is parsed from the OOG log, NOT res.GasUsed. On the ante-abort
//     path runTx's defer reads the OUTER context's meter, which for CheckTx is
//     checkState's InfiniteGasMeter; it accumulates across calls and is not the
//     tx's gas. Asserting res.GasUsed > 0 proves nothing.
//   - Each case asserts WHICH meter tripped. Below roughly 700k gas the tx OOGs
//     inside the ante (SetAccount alone is ~313k), so the VM never runs and the
//     measurement would say nothing about sponsorship. Windows start at 1M.
//   - The ante-only control pre-signs one tx per sequence: a fee-paying CheckTx
//     COMMITS its sequence increment (RunTxModeCheck does MultiWrite), so
//     replaying a single tx would measure a signature mismatch after iteration 0.
//   - PebbleDB, not memdb: memdb's Iterator sorts the entire keyspace on every
//     call, which swamps the Simulate arm with an artifact absent on real nodes.
//
// Run:
//
//	go test -c -o bench.test ./pkg/gnoland/
//	for i in $(seq 1 10); do ./bench.test -test.run '^$' -test.bench Sponsorship \
//	    -test.benchtime 50x -test.count 1 -test.benchmem; done > raw.txt
//	benchstat raw.txt
//
// allocs/op is far more stable than ns/op here; prefer it as the work proxy.

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/db/pebbledb"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// creditWindows sweeps Block.MaxGasCreditPerTx. All are above the ~700k floor
// below which the ante, not the VM, is the binding constraint. 10M is what
// gno.land's own integration harness uses today.
var creditWindows = []int64{1_000_000, 5_000_000, 10_000_000}

// burnerRealm burns the whole budget in a pure-CPU loop with no storage writes,
// so what is measured is VM execution rather than store traffic. The bound is
// high enough that every window under test runs out of gas inside the loop, so
// each attempt consumes exactly its window. It never calls PayGas: this is the
// "burn the window, pay nothing" transaction.
const burnerRealm = `package burner

func Burn(cur realm) string {
	x := 0
	for i := 0; i < 1000000000; i++ {
		x += i
	}
	return "done"
}
`

var (
	gasUsedRe = regexp.MustCompile(`gas used \((\d+)\)`)
	gasOpRe   = regexp.MustCompile(`during operation: (\w+)`)
)

// vmOps are the meters that only trip while the VM is executing user code.
// Anything else (DepthSet from the ante's SetAccount, ReadFlat from package
// load) means the budget was spent before message execution.
var vmOps = map[string]bool{"CPUCycles": true, "memory": true}

func parseOOG(tb testing.TB, log string) (gas int64, op string) {
	tb.Helper()
	m := gasUsedRe.FindStringSubmatch(log)
	require.Len(tb, m, 2, "no `gas used (N)` in log: %.300s", log)
	gas, err := strconv.ParseInt(m[1], 10, 64)
	require.NoError(tb, err)
	o := gasOpRe.FindStringSubmatch(log)
	require.Len(tb, o, 2, "no `during operation:` in log: %.300s", log)
	return gas, o[1]
}

type benchEnv struct {
	app     *sdk.BaseApp
	chainID string
	priv    ed25519.PrivKeyEd25519
	addr    crypto.Address
	accNum  uint64
}

func mustDeployTx(deployer crypto.Address, pkgPath, file, body string) std.Tx {
	files := []*std.MemFile{
		{Name: file, Body: body},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	return std.Tx{
		Msgs:       []std.Msg{vm.NewMsgAddPackage(deployer, pkgPath, files)},
		Fee:        std.Fee{GasWanted: 1e8, GasFee: std.Coin{Amount: 1e6, Denom: "ugnot"}},
		Signatures: []std.Signature{{}},
	}
}

// setupSponsorshipBench builds a gno.land app on PebbleDB with the credit window
// open and the validator opted in, then advances past genesis so signatures are
// verified and the height-0 admission guard is not what gets measured.
func setupSponsorshipBench(tb testing.TB, maxGasCreditPerTx int64) *benchEnv {
	tb.Helper()

	db, err := pebbledb.NewPebbleDB("bench", tb.TempDir())
	require.NoError(tb, err)
	tb.Cleanup(func() { db.Close() })

	opts := TestAppOptions(db)
	opts.AllowZeroFeeTxs = true // validator opt-in
	app, err := NewAppWithOptions(opts)
	require.NoError(tb, err)
	bapp := app.(*sdk.BaseApp)

	priv := ed25519.GenPrivKey()
	addr := priv.PubKey().Address()
	deployer := crypto.AddressFromPreimage([]byte("bench_deployer"))

	appState := DefaultGenState()
	appState.Balances = []Balance{
		{Address: deployer, Amount: []std.Coin{{Amount: 1e15, Denom: "ugnot"}}},
		{Address: addr, Amount: []std.Coin{{Amount: 1e12, Denom: "ugnot"}}},
	}
	appState.Txs = []TxWithMetadata{
		{Tx: mustDeployTx(deployer, "gno.land/r/bench/burner", "burner.gno", burnerRealm)},
	}

	blockParams := defaultBlockParams()
	blockParams.MaxGasCreditPerTx = maxGasCreditPerTx

	chainID := "bench"
	resp := bapp.InitChain(abci.RequestInitChain{
		Time:            time.Now(),
		ChainID:         chainID,
		ConsensusParams: &abci.ConsensusParams{Block: blockParams},
		AppState:        appState,
	})
	require.True(tb, resp.IsOK(), "InitChain: %v", resp)

	bapp.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: chainID, Height: 1}})
	bapp.EndBlock(abci.RequestEndBlock{})
	bapp.Commit()

	// Query the real genesis-assigned account number rather than assuming it.
	q := bapp.Query(abci.RequestQuery{Path: "auth/accounts/" + addr.String()})
	require.True(tb, q.IsOK(), "account query: %v", q)
	var acc GnoAccount
	require.NoError(tb, amino.UnmarshalJSON(q.Data, &acc))

	return &benchEnv{app: bapp, chainID: chainID, priv: priv, addr: addr, accNum: acc.GetAccountNumber()}
}

func (e *benchEnv) signedTx(tb testing.TB, fee std.Fee, seq uint64) []byte {
	tb.Helper()
	tx := std.Tx{
		Msgs: []std.Msg{vm.NewMsgCall(e.addr, nil, "gno.land/r/bench/burner", "Burn", nil)},
		Fee:  fee,
	}
	sb, err := tx.GetSignBytes(e.chainID, e.accNum, seq)
	require.NoError(tb, err)
	sig, err := e.priv.Sign(sb)
	require.NoError(tb, err)
	tx.Signatures = []std.Signature{{PubKey: e.priv.PubKey(), Signature: sig}}
	return amino.MustMarshal(tx)
}

// BenchmarkSponsorshipCheckTxNoPayGas is the adversarial case: a signed 0-fee tx
// burns the whole credit window and never calls PayGas. It is rejected, so no
// sequence is consumed (only result.IsOK() flushes the ante checkpoint) and the
// identical bytes replay indefinitely. ns/op is the CPU bought per free attempt.
func BenchmarkSponsorshipCheckTxNoPayGas(b *testing.B) {
	for _, window := range creditWindows {
		b.Run(fmt.Sprintf("credit=%d", window), func(b *testing.B) {
			env := setupSponsorshipBench(b, window)
			txBytes := env.signedTx(b, std.Fee{
				GasWanted: window,
				GasFee:    std.Coin{Denom: "ugnot", Amount: 0},
			}, 0)

			res := env.app.CheckTx(abci.RequestCheckTx{Tx: txBytes})
			require.False(b, res.IsOK(), "expected rejection")
			gas, op := parseOOG(b, res.Log)
			require.True(b, vmOps[op], "budget spent outside the VM (op=%s, gas=%d)", op, gas)
			require.InEpsilon(b, window, gas, 0.05, "should burn ~the whole window")

			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				env.app.CheckTx(abci.RequestCheckTx{Tx: txBytes})
			}
			b.StopTimer()

			after := env.app.CheckTx(abci.RequestCheckTx{Tx: txBytes})
			gasAfter, opAfter := parseOOG(b, after.Log)
			require.Equal(b, op, opAfter, "behaviour drifted during the run")
			require.Equal(b, gas, gasAfter, "gas drifted during the run")

			b.ReportMetric(float64(gas), "gas")
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/(float64(gas)/1e6), "ns/Mgas")
		})
	}
}

// BenchmarkSponsorshipCheckTxAnteOnly is the control: an ordinary fee-paying
// CheckTx runs the ante and never executes messages. It commits its sequence
// increment, so each iteration needs its own pre-signed sequence.
func BenchmarkSponsorshipCheckTxAnteOnly(b *testing.B) {
	for _, window := range creditWindows {
		b.Run(fmt.Sprintf("gasWanted=%d", window), func(b *testing.B) {
			env := setupSponsorshipBench(b, window)
			txs := make([][]byte, b.N)
			for i := range txs {
				txs[i] = env.signedTx(b, std.Fee{
					GasWanted: window,
					GasFee:    std.Coin{Denom: "ugnot", Amount: 1_000_000},
				}, uint64(i))
			}

			b.ReportAllocs()
			b.ResetTimer()
			ok := 0
			for i := range b.N {
				if env.app.CheckTx(abci.RequestCheckTx{Tx: txs[i]}).IsOK() {
					ok++
				}
			}
			b.StopTimer()
			require.Equal(b, b.N, ok, "every ante-only CheckTx must succeed")
		})
	}
}
