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
	"fmt"
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

	// Enough boots that two of them land in the same unfilled cache often
	// enough to matter. The barrier is what makes them overlap, at any core
	// count.
	const nodes = 24

	// Failures are collected rather than asserted in place: testing.T.FailNow,
	// which require reaches, has to run on the test's own goroutine.
	errs := make([]error, nodes)

	var wg, gate sync.WaitGroup
	gate.Add(1)
	for i := range nodes {
		wg.Go(func() {
			gate.Wait() // release every boot at once
			app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
			if err != nil {
				errs[i] = fmt.Errorf("node %d NewAppWithOptions: %w", i, err)
				return
			}
			bapp := app.(*sdk.BaseApp)
			resp := bapp.InitChain(abci.RequestInitChain{
				ChainID: "dev",
				Time:    time.Now(),
				ConsensusParams: &abci.ConsensusParams{
					Block: &abci.BlockParams{MaxGas: 1e10, MaxTxBytes: 1e7, MaxDataBytes: 1e7},
				},
				AppState: DefaultGenState(),
			})
			if !resp.IsOK() {
				errs[i] = fmt.Errorf("node %d InitChain: %v", i, resp)
				return
			}
			bapp.Commit()
		})
	}
	gate.Done()
	wg.Wait()

	for _, err := range errs {
		require.NoError(t, err)
	}
}
