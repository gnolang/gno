package gnoclient

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/log"
)

func TestSimulateBurstDuringCommit(t *testing.T) {
	t.Parallel()

	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(rootdir)

	meta := loadpkgs(t, rootdir, "gno.land/r/tests/vm/deep/very/deep")
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	signer := newInMemorySigner(t, "tendermint_test")
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/tests/vm/deep/very/deep",
		Func:    "RenderCrossing",
		Args:    []string{"burst"},
		Send:    nil,
	}

	simCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      21000000,
		AccountNumber:  0,
		SequenceNumber: 0,
	}

	simTx, err := NewCallTx(simCfg, msg)
	require.NoError(t, err)
	signedSimTx, err := client.SignTx(*simTx, 0, 0)
	require.NoError(t, err)

	account, _, err := client.QueryAccount(caller.GetAddress())
	require.NoError(t, err)
	startSeq := account.Sequence
	acctNum := account.AccountNumber

	const numTxs = 100
	signedTxBytes := make([][]byte, numTxs)
	for i := range numTxs {
		seq := startSeq + uint64(i)
		cfg := BaseTxCfg{
			GasFee:         ugnot.ValueString(2100000),
			GasWanted:      21000000,
			AccountNumber:  acctNum,
			SequenceNumber: seq,
			Memo:           fmt.Sprintf("burst-%d", i),
		}
		tx, err := NewCallTx(cfg, msg)
		require.NoError(t, err)
		signed, err := client.SignTx(*tx, acctNum, seq)
		require.NoError(t, err)
		bz, err := amino.Marshal(signed)
		require.NoError(t, err)
		signedTxBytes[i] = bz
	}

	var (
		simulateErrors atomic.Int64
		simulateTotal  atomic.Int64
	)

	done := make(chan struct{})
	var wg sync.WaitGroup

	for i := range 8 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					simulateTotal.Add(1)
					_, err := client.Simulate(signedSimTx)
					if err != nil {
						simulateErrors.Add(1)
					}
				}
			}
		}(i)
	}

	for wave := 0; wave < 10; wave++ {
		for j := 0; j < 10; j++ {
			idx := wave*10 + j
			rpcClient.BroadcastTxAsync(context.Background(), signedTxBytes[idx])
		}
		time.Sleep(200 * time.Millisecond)
	}

	time.Sleep(10 * time.Second)
	close(done)
	wg.Wait()

	total := simulateTotal.Load()
	errors := simulateErrors.Load()
	t.Logf("Simulate: total=%d errors=%d (%.1f%% error rate)", total, errors,
		float64(errors)/float64(total)*100)

	assert.Zero(t, errors, "simulate had errors during concurrent block commits")
}
