package gnoland

// Several nodes booting in one process, which is what gno.land/pkg/integration
// does. Every InitChain loads the stdlibs from the process-global cached store,
// and CopyFromCachedStore hands each node the SAME BlockNode and Type pointers,
// so every node's transactionStore.Write runs the sealer over one shared graph.
//
// That makes this the only test in the tree with more than one goroutine
// publishing at once, which is the case sealing is NOT written for: the sealer
// assumes the publishing goroutine owns the graph. It stays correct here only
// because every filler it calls is check-then-set, so re-sealing an already
// sealed graph writes nothing. Delete `ct.methodIndex == nil &&` from
// seal.go's *DeclaredType case and this dies with
// `fatal error: concurrent map writes`, or reports a data race under -race.
//
// gnovm/pkg/gnolang's TestSealSkipsBuiltMethodIndex pins the same guard as a
// unit test; this one pins that the guard is what a real boot depends on.

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk"
)

func TestConcurrentInitChain(t *testing.T) {
	if testing.Short() {
		t.Skip("boots 24 nodes in one process")
	}

	// Above any core count this runs on, so the boots overlap rather than
	// queueing behind GOMAXPROCS.
	const nodes = 24

	var wg, gate sync.WaitGroup
	gate.Add(1)
	for i := range nodes {
		wg.Go(func() {
			gate.Wait() // release every boot at once
			app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
			require.NoError(t, err)
			bapp := app.(*sdk.BaseApp)
			resp := bapp.InitChain(abci.RequestInitChain{
				ChainID: "dev",
				Time:    time.Now(),
				ConsensusParams: &abci.ConsensusParams{
					Block: &abci.BlockParams{MaxGas: 1e10, MaxTxBytes: 1e7, MaxDataBytes: 1e7},
				},
				AppState: DefaultGenState(),
			})
			require.Truef(t, resp.IsOK(), "node %d InitChain: %v", i, resp)
			bapp.Commit()
		})
	}
	gate.Done()
	wg.Wait()
}
