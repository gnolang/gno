package consensus

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/errors"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	mempl "github.com/gnolang/gno/tm2/pkg/bft/mempool"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// for testing
func assertMempool(txn txNotifier) mempl.Mempool {
	return txn.(mempl.Mempool)
}

func TestMempoolNoProgressUntilTxsAvailable(t *testing.T) {
	t.Parallel()

	config, _ := ResetConfig("consensus_mempool_no_progress_until_txs_available")
	defer os.RemoveAll(config.RootDir)
	config.Consensus.CreateEmptyBlocks = false
	state, privVals := randGenesisState(1, false, 10)
	app := NewCounterApplication()
	cs := newConsensusStateWithConfig(config, state, privVals[0], app)
	assertMempool(cs.txNotifier).EnableTxsAvailable()
	height, round := cs.Height, cs.Round
	newBlockCh := subscribe(cs.evsw, types.EventNewBlock{})
	startFrom(cs, height, round)
	defer func() {
		cs.Stop()
		cs.Wait()
		_ = app.Close()
	}()

	ensureNewEventOnChannel(newBlockCh) // first block gets committed
	ensureNoNewEventOnChannel(newBlockCh)
	deliverTxsRange(cs, 0, 1)
	ensureNewEventOnChannel(newBlockCh) // commit txs
	ensureNewEventOnChannel(newBlockCh) // commit updated app hash
	ensureNoNewEventOnChannel(newBlockCh)
}

func TestMempoolProgressAfterCreateEmptyBlocksInterval(t *testing.T) {
	config, _ := ResetConfig("consensus_mempool_progress_after_create_empty_blocks_interval")
	defer os.RemoveAll(config.RootDir)
	config.Consensus.CreateEmptyBlocksInterval = ensureTimeout
	state, privVals := randGenesisState(1, false, 10)
	app := NewCounterApplication()
	cs := newConsensusStateWithConfig(config, state, privVals[0], app)
	assertMempool(cs.txNotifier).EnableTxsAvailable()
	height, round := cs.Height, cs.Round
	newBlockCh := subscribe(cs.evsw, types.EventNewBlock{})
	startFrom(cs, height, round)
	defer func() {
		cs.Stop()
		cs.Wait()
		_ = app.Close()
	}()

	ensureNewEventOnChannel(newBlockCh)   // first block gets committed
	ensureNoNewEventOnChannel(newBlockCh) // then we dont make a block ...
	ensureNewEventOnChannel(newBlockCh)   // until the CreateEmptyBlocksInterval has passed
}

func TestMempoolProgressInHigherRound(t *testing.T) {
	t.Parallel()

	config, _ := ResetConfig("consensus_mempool_progress_in_higher_round")
	defer os.RemoveAll(config.RootDir)
	config.Consensus.CreateEmptyBlocks = false
	state, privVals := randGenesisState(1, false, 10)
	app := NewCounterApplication()
	cs := newConsensusStateWithConfig(config, state, privVals[0], app)
	assertMempool(cs.txNotifier).EnableTxsAvailable()
	height, round := cs.Height, cs.Round
	newBlockCh := subscribe(cs.evsw, types.EventNewBlock{})
	newStepCh := subscribe(cs.evsw, cstypes.EventNewRoundStep{})
	timeoutCh := subscribe(cs.evsw, cstypes.EventTimeoutPropose{})
	cs.setProposal = func(proposal *types.Proposal) error {
		if cs.Height == 2 && cs.Round == 0 {
			// dont set the proposal in round 0 so we timeout and
			// go to next round
			cs.Logger.Info("Ignoring set proposal at height 2, round 0")
			return nil
		}
		return cs.defaultSetProposal(proposal)
	}
	startFrom(cs, height, round)
	defer func() {
		cs.Stop()
		cs.Wait()
		_ = app.Close()
	}()

	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepPropose)   // first round at first height
	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepPrevote)   // ...
	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepPrecommit) // ...
	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepCommit)    // ...
	ensureNewEventOnChannel(newBlockCh)                                      // first block gets committed

	height++ // moving to the next height
	round = 0

	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepNewHeight) // new height
	deliverTxsRange(cs, 0, 1)                                                // we deliver txs, but dont set a proposal so we get the next round
	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepPropose)   // first round at next height

	ensureNewTimeout(timeoutCh, height, round, cs.config.TimeoutPropose.Nanoseconds())

	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepPrevote)       // ...
	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepPrecommit)     // ...
	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepPrecommitWait) // ...

	round++                                                                  // moving to the next round
	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepPropose)   // wait for the next round
	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepPrevote)   // ...
	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepPrecommit) // ...
	ensureNewRoundStep(newStepCh, height, round, cstypes.RoundStepCommit)    // ...
	ensureNewEventOnChannel(newBlockCh)                                      // now we can commit the block
}

func deliverTxsRange(cs *ConsensusState, start, end int) {
	// Deliver some txs.
	for i := start; i < end; i++ {
		txBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(txBytes, uint64(i))
		err := assertMempool(cs.txNotifier).CheckTx(txBytes, nil)
		if err != nil {
			panic(fmt.Sprintf("Error after CheckTx: %v", err))
		}
	}
}

func TestMempoolTxConcurrentWithCommit(t *testing.T) {
	t.Parallel()

	state, privVals := randGenesisState(1, false, 10)
	blockDB := memdb.NewMemDB()
	app := NewCounterApplication()
	cs := newConsensusStateWithConfigAndBlockStore(config, state, privVals[0], app, blockDB)
	sm.SaveState(blockDB, state)
	height, round := cs.Height, cs.Round
	newBlockCh := subscribe(cs.evsw, types.EventNewBlock{})

	NTxs := 3000
	go deliverTxsRange(cs, 0, NTxs)

	startFrom(cs, height, round)
	defer func() {
		cs.Stop()
		cs.Wait()
		_ = app.Close()
	}()

	for nTxs := 0; nTxs < NTxs; {
		ticker := time.NewTicker(time.Second * 30)
		select {
		case msg := <-newBlockCh:
			blockEvent := msg.(types.EventNewBlock)
			nTxs += int(blockEvent.Block.Header.NumTxs)
		case <-ticker.C:
			panic("Timed out waiting to commit blocks with transactions")
		}
	}
}

func TestMempoolRmBadTx(t *testing.T) {
	t.Parallel()

	state, privVals := randGenesisState(1, false, 10)
	app := NewCounterApplication()
	blockDB := memdb.NewMemDB()
	cs := newConsensusStateWithConfigAndBlockStore(config, state, privVals[0], app, blockDB)
	sm.SaveState(blockDB, state)

	// increment the counter by 1
	txBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(txBytes, uint64(0))

	resDeliver := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
	assert.False(t, resDeliver.IsErr(), fmt.Sprintf("expected no error. got %v", resDeliver))

	resCommit := app.Commit()
	assert.True(t, len(resCommit.Data) > 0)

	emptyMempoolCh := make(chan struct{})
	checkTxRespCh := make(chan struct{})
	go func() {
		// Try to send the tx through the mempool.
		// CheckTx should not err, but the app should return an abci Error.
		// and the tx should get removed from the pool
		err := assertMempool(cs.txNotifier).CheckTx(txBytes, func(r abci.Response) {
			if _, ok := r.(abci.ResponseCheckTx).Error.(errors.BadNonceError); !ok {
				t.Errorf("expected checktx to return bad nonce, got %v", r)
				return
			}
			checkTxRespCh <- struct{}{}
		})
		if err != nil {
			t.Errorf("Error after CheckTx: %v", err)
			return
		}

		// check for the tx
		for {
			txs := assertMempool(cs.txNotifier).ReapMaxBytesMaxGas(int64(len(txBytes)), -1)
			if len(txs) == 0 {
				emptyMempoolCh <- struct{}{}
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Wait until the tx returns
	ticker := time.After(time.Second * 5)
	select {
	case <-checkTxRespCh:
		// success
	case <-ticker:
		t.Errorf("Timed out waiting for tx to return")
		return
	}

	// Wait until the tx is removed
	ticker = time.After(time.Second * 5)
	select {
	case <-emptyMempoolCh:
		// success
	case <-ticker:
		t.Errorf("Timed out waiting for tx to be removed")
		return
	}
}

// CounterApplication that maintains a mempool state and resets it upon commit
type CounterApplication struct {
	abci.BaseApplication

	txCount        int
	mempoolTxCount int
}

func NewCounterApplication() *CounterApplication {
	return &CounterApplication{}
}

func (app *CounterApplication) Info(req abci.RequestInfo) (res abci.ResponseInfo) {
	res.Data = fmt.Appendf(nil, "txs:%v", app.txCount)
	return
}

func (app *CounterApplication) DeliverTx(req abci.RequestDeliverTx) (res abci.ResponseDeliverTx) {
	txValue := txAsUint64(req.Tx)
	if txValue != uint64(app.txCount) {
		res.Error = errors.BadNonceError{}
		res.Log = fmt.Sprintf("Invalid nonce. Expected %v, got %v", app.txCount, txValue)
		return
	}
	app.txCount++
	return
}

func (app *CounterApplication) CheckTx(req abci.RequestCheckTx) (res abci.ResponseCheckTx) {
	txValue := txAsUint64(req.Tx)
	if txValue != uint64(app.mempoolTxCount) {
		res.Error = errors.BadNonceError{}
		res.Log = fmt.Sprintf("Invalid nonce. Expected %v, got %v", app.mempoolTxCount, txValue)
		return
	}
	app.mempoolTxCount++
	return
}

func txAsUint64(tx []byte) uint64 {
	tx8 := make([]byte, 8)
	copy(tx8[len(tx8)-len(tx):], tx)
	return binary.BigEndian.Uint64(tx8)
}

func (app *CounterApplication) Commit() (res abci.ResponseCommit) {
	app.mempoolTxCount = app.txCount
	if app.txCount == 0 {
		return abci.ResponseCommit{}
	}
	hash := make([]byte, 8)
	binary.BigEndian.PutUint64(hash, uint64(app.txCount))
	res.Data = hash
	return
}
