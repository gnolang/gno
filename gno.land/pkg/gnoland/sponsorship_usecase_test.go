package gnoland

// Does the REAL sponsorship use case fit in a conservative credit window?
//
// The motivating pattern is a paymaster: the realm collects payment in a token,
// then calls PayGas so the user needs no gnot. That is the whole point of
// "the realm decides during execution". If it does not fit under a small
// window, the recommendation to activate at a small window is wrong and the
// off-chain co-signer model is the better trust boundary.
//
// This measures the gas that pattern actually consumes end to end: a GRC20-style
// ledger transfer (storage reads + writes on an avl tree) followed by PayGas.
//
//	go test ./pkg/gnoland/ -run TestSponsorshipUseCaseFitsWindow -v

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// paymasterRealm is the realistic shape: take payment in a token, then sponsor.
// The ledger is a plain avl tree, which is what a GRC20 balance ledger is, so
// the storage reads and writes are representative rather than synthetic.
const paymasterRealm = `package paymaster

import (
	"chain/runtime"
	"gno.land/p/nt/avl/v0"
)

var balances avl.Tree

func init() {
	// Pre-seed so the transfer path hits real tree reads and writes.
	for i := 0; i < 64; i++ {
		balances.Set(string(rune('a'+i%26))+string(rune('a'+i/26)), int64(1000000))
	}
}

// Sponsor collects payment in the realm's own token ledger, then pays the gas.
// This is the "cheap checks, then early PayGas" shape.
func Sponsor(cur realm, payer string) string {
	// 1. collect payment: read + write the ledger (the "approve/transfer" work)
	v := balances.Get(payer)
	if v == nil {
		panic("no balance")
	}
	bal := v.(int64)
	if bal < 1000 {
		panic("insufficient token balance")
	}
	balances.Set(payer, bal-1000)

	treasury := balances.Get("aa")
	balances.Set("aa", treasury.(int64)+1000)

	// 2. only now commit to paying the user's gas
	runtime.PayGas(5000000)
	return "sponsored"
}
`

// deployExampleTx builds a genesis deploy tx for a package taken from the
// examples tree, so the use case exercises the real avl implementation rather
// than a stand-in.
func deployExampleTx(tb testing.TB, deployer crypto.Address, dir, pkgPath string) std.Tx {
	tb.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(tb, err)
	var files []*std.MemFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".gno") || strings.HasSuffix(e.Name(), "_test.gno") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, e.Name()))
		require.NoError(tb, err)
		files = append(files, &std.MemFile{Name: e.Name(), Body: string(body)})
	}
	files = append(files, &std.MemFile{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)})
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
	return std.Tx{
		Msgs:       []std.Msg{vm.NewMsgAddPackage(deployer, pkgPath, files)},
		Fee:        std.Fee{GasWanted: 5e8, GasFee: std.Coin{Amount: 1e6, Denom: "ugnot"}},
		Signatures: []std.Signature{{}},
	}
}

// paygasOnlyRealm is the absolute floor: sponsorship machinery and nothing else.
const paygasOnlyRealm = `package paymaster

import "chain/runtime"

func Sponsor(cur realm, payer string) string {
	runtime.PayGas(5000000)
	return "sponsored"
}
`

// mapLedgerRealm is a lean paymaster: a package-level map as the ledger, no
// external imports. Collect payment, then sponsor.
const mapLedgerRealm = `package paymaster

import "chain/runtime"

var balances = map[string]int64{"ab": 1000000, "aa": 0}

func Sponsor(cur realm, payer string) string {
	bal := balances[payer]
	if bal < 1000 {
		panic("insufficient token balance")
	}
	balances[payer] = bal - 1000
	balances["aa"] += 1000

	runtime.PayGas(5000000)
	return "sponsored"
}
`

// TestSponsorshipUseCaseFitsWindow reports the gas the paymaster pattern costs
// and the smallest credit window under which it succeeds.
func TestSponsorshipUseCaseFitsWindow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		src     string
		withAVL bool
	}{
		{"paygas-only (floor)", paygasOnlyRealm, false},
		{"map ledger + paygas", mapLedgerRealm, false},
		{"avl ledger + paygas", paymasterRealm, true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			const window = 30_000_000 // generous, so we measure cost not the cap

			opts := TestAppOptions(memdb.NewMemDB())
			opts.AllowZeroFeeTxs = true
			app, err := NewAppWithOptions(opts)
			require.NoError(t, err)
			bapp := app.(*sdk.BaseApp)

			deployer := crypto.AddressFromPreimage([]byte("uc_deployer"))
			priv := ed25519.GenPrivKey()
			addr := priv.PubKey().Address()

			st := DefaultGenState()
			st.Balances = []Balance{
				{Address: deployer, Amount: []std.Coin{{Amount: 1e15, Denom: "ugnot"}}},
				{Address: addr, Amount: []std.Coin{{Amount: 1e12, Denom: "ugnot"}}},
				{Address: gnolang.DerivePkgCryptoAddr("gno.land/r/uc/paymaster"), Amount: []std.Coin{{Amount: 1e14, Denom: "ugnot"}}},
			}
			ex := filepath.Join(gnoenv.RootDir(), "examples", "gno.land", "p", "nt")
			if tc.withAVL {
				st.Txs = append(st.Txs,
					TxWithMetadata{Tx: deployExampleTx(t, deployer, filepath.Join(ex, "ufmt", "v0"), "gno.land/p/nt/ufmt/v0")},
					TxWithMetadata{Tx: deployExampleTx(t, deployer, filepath.Join(ex, "avl", "v0"), "gno.land/p/nt/avl/v0")},
				)
			}
			st.Txs = append(st.Txs,
				TxWithMetadata{Tx: mustDeployTx(deployer, "gno.land/r/uc/paymaster", "paymaster.gno", tc.src)})

			bp := defaultBlockParams()
			bp.MaxGasCreditPerTx = window
			r := bapp.InitChain(abci.RequestInitChain{
				Time: time.Now(), ChainID: "uc",
				ConsensusParams: &abci.ConsensusParams{Block: bp}, AppState: st,
			})
			require.True(t, r.IsOK(), "InitChain: %v", r)

			bapp.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{ChainID: "uc", Height: 1}})
			bapp.EndBlock(abci.RequestEndBlock{})
			bapp.Commit()

			q := bapp.Query(abci.RequestQuery{Path: "auth/accounts/" + addr.String()})
			require.True(t, q.IsOK())
			var acc GnoAccount
			require.NoError(t, amino.UnmarshalJSON(q.Data, &acc))

			tx := std.Tx{
				Msgs: []std.Msg{vm.NewMsgCall(addr, nil, "gno.land/r/uc/paymaster", "Sponsor", []string{"ab"})},
				Fee:  std.Fee{GasWanted: window, GasFee: std.Coin{Denom: "ugnot", Amount: 0}},
			}
			sb, err := tx.GetSignBytes("uc", acc.GetAccountNumber(), 0)
			require.NoError(t, err)
			sig, err := priv.Sign(sb)
			require.NoError(t, err)
			tx.Signatures = []std.Signature{{PubKey: priv.PubKey(), Signature: sig}}

			res := bapp.CheckTx(abci.RequestCheckTx{Tx: amino.MustMarshal(tx)})
			require.True(t, res.IsOK(), "should be admitted at a 30M window: %v | %.200s", res.Error, res.Log)
			t.Logf("%-22s gasUsed=%8d  => needs a window of at least ~%dM", tc.name, res.GasUsed, (res.GasUsed+999_999)/1_000_000)
		})
	}
}
