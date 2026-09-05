package gnoland

// Concurrency validation for the VM query path.
//
// TestParallelQueries_NWaySimulate covers .app/simulate, which routes to the
// bank handler and never enters the GnoVM. vm/qeval and vm/qrender go through
// handleQueryCustom into gno.Machine, preprocess, and the process-wide type
// graph reachable from the shared cacheNodes map — the machinery the lock-free
// query connection's invariant actually names, and the machinery nothing else
// in the tree exercises concurrently. Production already fans these out:
// gnoweb issues several VM queries per page render from an errgroup.
//
// The realm below is shaped to reach every lazily-memoised cache on a shared
// graph in one query:
//
//   - FuncType.bound / FuncType.typeid, via a concrete-to-interface conversion
//     preprocessed fresh per query (VerifyImplementedBy -> BoundType -> TypeID);
//   - DeclaredType.pkgID and StructType.pkgID, via allocation construction-time
//     checks while the body runs;
//   - DeclaredType.methodIndex, which builds only past methodIndexThreshold (8),
//     so Wide carries more than eight methods;
//   - StaticBlock.nameIndex, which builds only past nameIndexThreshold (32), so
//     the package declares more than thirty-two top-level names;
//   - StructType.comparable and the effective field/method counts, via map keys
//     and interface satisfaction.
//
// Run with -race. Without it this only asserts that concurrent VM queries agree.

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// parallelVMRealm builds a realm with more than nameIndexThreshold (32)
// top-level names and a declared type with more than methodIndexThreshold (8)
// methods, so both lazy indexes are past their build gates.
func parallelVMRealm() string {
	var b strings.Builder
	b.WriteString(`package vmrace

type Speaker interface {
	Speak() string
	Name() string
}

type Dog struct{ n string }

func (d Dog) Speak() string { return "woof" }
func (d Dog) Name() string  { return d.n }

type Cat struct{ n string }

func (c Cat) Speak() string { return "meow" }
func (c Cat) Name() string  { return c.n }

// Wide crosses methodIndexThreshold so DeclaredType.methodIndex is built.
type Wide struct{ v int }

func (w Wide) M0() int { return w.v }
func (w Wide) M1() int { return w.v + 1 }
func (w Wide) M2() int { return w.v + 2 }
func (w Wide) M3() int { return w.v + 3 }
func (w Wide) M4() int { return w.v + 4 }
func (w Wide) M5() int { return w.v + 5 }
func (w Wide) M6() int { return w.v + 6 }
func (w Wide) M7() int { return w.v + 7 }
func (w Wide) M8() int { return w.v + 8 }
func (w Wide) M9() int { return w.v + 9 }
func (w Wide) Speak() string { return "wide" }
func (w Wide) Name() string  { return "wide" }

func NewDog() Dog   { return Dog{n: "rex"} }
func NewCat() Cat   { return Cat{n: "tom"} }
func NewWide() Wide { return Wide{v: 7} }

// Describe forces a concrete-to-interface conversion at every call site that
// is preprocessed fresh, which is what a query expression is.
func Describe(s Speaker) string { return s.Name() + " says " + s.Speak() }

// Tally uses a struct as a map key, which fills StructType.comparable, and
// builds an anonymous struct type inside a body.
func Tally() int {
	type pair struct {
		a string
		b int
	}
	counts := map[pair]int{}
	counts[pair{a: "x", b: 1}]++
	counts[pair{a: "y", b: 2}] += 2
	total := 0
	for _, v := range counts {
		total += v
	}
	return total
}

func Render(path string) string {
	return Describe(NewDog()) + "|" + Describe(NewCat()) + "|" + Describe(NewWide())
}
`)
	// Pad past nameIndexThreshold (32) so StaticBlock.nameIndex is built.
	for i := range 40 {
		fmt.Fprintf(&b, "\nfunc Pad%d() int { return %d }\n", i, i)
	}
	return b.String()
}

// buildParallelVMApp returns an app with the vmrace realm deployed and one real
// block committed, so Simulate's pre-first-block fallback cannot be what a
// failure here is measuring.
func buildParallelVMApp(t *testing.T) *sdk.BaseApp {
	t.Helper()

	const pkgPath = "gno.land/r/vmrace"
	deployer := crypto.AddressFromPreimage([]byte("test1"))

	appState := DefaultGenState()
	appState.Balances = []Balance{
		{Address: deployer, Amount: []std.Coin{{Amount: 1e15, Denom: "ugnot"}}},
	}
	appState.Txs = []TxWithMetadata{
		{Tx: std.Tx{
			Msgs: []std.Msg{vm.NewMsgAddPackage(deployer, pkgPath, []*std.MemFile{
				{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
				{Name: "vmrace.gno", Body: parallelVMRealm()},
			})},
			Fee:        std.Fee{GasWanted: 1e9, GasFee: std.Coin{Amount: 1e8, Denom: "ugnot"}},
			Signatures: []std.Signature{{}},
		}},
	}

	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)
	bapp := app.(*sdk.BaseApp)

	resp := bapp.InitChain(abci.RequestInitChain{
		Time:            time.Now(),
		ChainID:         "dev",
		ConsensusParams: &abci.ConsensusParams{Block: &abci.BlockParams{MaxGas: 1e10, MaxTxBytes: 1e7, MaxDataBytes: 1e7}},
		AppState:        appState,
	})
	require.True(t, resp.IsOK(), "InitChain: %v", resp)
	bapp.Commit()

	bapp.BeginBlock(abci.RequestBeginBlock{
		Header: &bft.Header{ChainID: "dev", Height: 1, Time: time.Now()},
	})
	bapp.EndBlock(abci.RequestEndBlock{})
	bapp.Commit()
	require.GreaterOrEqual(t, bapp.LastBlockHeight(), int64(1))

	return bapp
}

// vmQueries are the expressions the queriers rotate through. Each reaches a
// different corner of the shared graph.
var vmQueries = []struct {
	path string
	data string
}{
	{"vm/qeval", "gno.land/r/vmrace.Describe(NewDog())"},
	{"vm/qeval", "gno.land/r/vmrace.Describe(NewCat())"},
	{"vm/qeval", "gno.land/r/vmrace.Describe(NewWide())"},
	{"vm/qeval", "gno.land/r/vmrace.NewWide().M9()"},
	{"vm/qeval", "gno.land/r/vmrace.Tally()"},
	{"vm/qeval", "gno.land/r/vmrace.Pad17()"},
	{"vm/qrender", "gno.land/r/vmrace:"},
}

// TestParallelVMQueries fires N goroutines through the read-only connection,
// each rotating over every expression, and requires that they all agree.
func TestParallelVMQueries(t *testing.T) {
	if testing.Short() {
		t.Skip("concurrent VM query stress")
	}

	const (
		queriers = 8
		rounds   = 4
	)

	bapp := buildParallelVMApp(t)
	creator := proxy.NewLocalClientCreator(bapp)
	query, err := creator.NewReadOnlyABCIClient()
	require.NoError(t, err)
	require.NoError(t, query.Start())
	defer query.Stop()

	// ORDER MATTERS. Every cache this test is about is filled on FIRST touch,
	// so a serial baseline taken before the barrier would fill the whole shared
	// graph single-threaded and leave the concurrent phase reading warm, already
	// memoised fields — the test would pass on racy code. The concurrent phase
	// therefore runs first, against a cold graph, and the serial baseline is
	// captured afterwards purely as the value oracle.
	var (
		start sync.WaitGroup
		wg    sync.WaitGroup
	)
	start.Add(1)
	errs := make([]error, queriers)
	got := make([][]string, queriers)

	for g := range queriers {
		got[g] = make([]string, len(vmQueries))
		wg.Go(func() {
			start.Wait()
			for range rounds {
				for i, q := range vmQueries {
					res, err := query.QuerySync(abci.RequestQuery{Path: q.path, Data: []byte(q.data)})
					if err != nil {
						errs[g] = fmt.Errorf("%s %s: %w", q.path, q.data, err)
						return
					}
					if res.IsErr() {
						errs[g] = fmt.Errorf("%s %s: %w (log %q)", q.path, q.data, res.Error, res.Log)
						return
					}
					got[g][i] = string(res.Data)
				}
			}
		})
	}

	start.Done()
	wg.Wait()

	for g := range queriers {
		require.NoError(t, errs[g], "querier %d", g)
	}

	// Serial oracle: the same expressions with nothing else in flight. Every
	// concurrent reading must equal it.
	for i, q := range vmQueries {
		res, err := query.QuerySync(abci.RequestQuery{Path: q.path, Data: []byte(q.data)})
		require.NoError(t, err, "serial %s %s", q.path, q.data)
		require.Falsef(t, res.IsErr(), "serial %s %s: %v (log %q)", q.path, q.data, res.Error, res.Log)
		want := string(res.Data)
		for g := range queriers {
			require.Equalf(t, want, got[g][i],
				"querier %d: %s %s returned %q under concurrency but %q serially",
				g, q.path, q.data, got[g][i], want)
		}
	}
}
