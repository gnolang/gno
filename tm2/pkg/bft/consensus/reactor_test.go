package consensus

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/bitarray"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	"github.com/gnolang/gno/tm2/pkg/events"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

// ----------------------------------------------
// in-process testnets

func startConsensusNet(
	t *testing.T,
	css []*ConsensusState,
	n int,
) ([]*ConsensusReactor, []<-chan events.Event, []events.EventSwitch, []*p2p.MultiplexSwitch) {
	t.Helper()

	reactors := make([]*ConsensusReactor, n)
	blocksSubs := make([]<-chan events.Event, 0)
	eventSwitches := make([]events.EventSwitch, n)
	p2pSwitches := ([]*p2p.MultiplexSwitch)(nil)
	options := make(map[int][]p2p.SwitchOption)
	for i := range n {
		/*logger, err := tmflags.ParseLogLevel("consensus:info,*:error", logger, "info")
		if err != nil {	t.Fatal(err)}*/
		reactors[i] = NewConsensusReactor(css[i], true) // so we dont start the consensus states
		reactors[i].SetLogger(css[i].Logger)

		options[i] = []p2p.SwitchOption{
			p2p.WithReactor("CONSENSUS", reactors[i]),
		}

		// evsw is already started with the cs
		eventSwitches[i] = css[i].evsw
		reactors[i].SetEventSwitch(eventSwitches[i])

		blocksSub := subscribe(eventSwitches[i], types.EventNewBlock{})
		blocksSubs = append(blocksSubs, blocksSub)

		if css[i].state.LastBlockHeight == 0 { // simulate handle initChain in handshake
			sm.SaveState(css[i].blockExec.DB(), css[i].state)
		}
	}
	// make connected switches and start all reactors
	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	testingCfg := p2pTesting.TestingConfig{
		P2PCfg:        config.P2P,
		Count:         n,
		SwitchOptions: options,
		Channels: []byte{
			StateChannel,
			DataChannel,
			VoteChannel,
			VoteSetBitsChannel,
		},
	}

	p2pSwitches, _ = p2pTesting.MakeConnectedPeers(t, ctx, testingCfg)

	// now that everyone is connected,  start the state machines
	// If we started the state machines before everyone was connected,
	// we'd block when the cs fires NewBlockEvent and the peers are trying to start their reactors
	// TODO: is this still true with new pubsub?
	for i := range n {
		s := reactors[i].conS.GetState()
		reactors[i].SwitchToConsensus(s, 0)
	}
	return reactors, blocksSubs, eventSwitches, p2pSwitches
}

func stopConsensusNet(
	logger *slog.Logger,
	reactors []*ConsensusReactor,
	eventSwitches []events.EventSwitch,
	p2pSwitches []*p2p.MultiplexSwitch,
) {
	logger.Info("stopConsensusNet", "n", len(reactors))
	for i := range reactors {
		logger.Info("stopConsensusNet: Stopping ConsensusReactor", "i", i)
	}
	for i, b := range eventSwitches {
		logger.Info("stopConsensusNet: Stopping evsw", "i", i)
		b.Stop()
	}
	for i, p := range p2pSwitches {
		logger.Info("stopConsensusNet: Stopping p2p switch", "i", i)
		p.Stop()
	}
	logger.Info("stopConsensusNet: DONE", "n", len(reactors))
}

// Ensure a testnet makes blocks
func TestReactorBasic(t *testing.T) {
	t.Parallel()

	N := 4
	css, cleanup := randConsensusNet(N, "consensus_reactor_test", newMockTickerFunc(true), newCounter)
	defer cleanup()
	reactors, blocksSubs, eventSwitches, p2pSwitches := startConsensusNet(t, css, N)
	defer stopConsensusNet(log.NewTestingLogger(t), reactors, eventSwitches, p2pSwitches)
	// wait till everyone makes the first new block
	timeoutWaitGroup(t, N, func(j int) {
		<-blocksSubs[j]
	}, css)
}

// ------------------------------------

// Ensure a testnet makes blocks when there are txs
func TestReactorCreatesBlockWhenEmptyBlocksFalse(t *testing.T) {
	t.Parallel()

	N := 4
	css, cleanup := randConsensusNet(N, "consensus_reactor_test", newMockTickerFunc(true), newCounter,
		func(c *cfg.Config) {
			c.Consensus.CreateEmptyBlocks = false
		})
	defer cleanup()
	reactors, blocksSubs, eventSwitches, p2pSwitches := startConsensusNet(t, css, N)
	defer stopConsensusNet(log.NewTestingLogger(t), reactors, eventSwitches, p2pSwitches)

	// send a tx
	if err := assertMempool(css[3].txNotifier).CheckTx([]byte{1, 2, 3}, nil); err != nil {
		t.Error(err)
	}

	// wait till everyone makes the first new block
	timeoutWaitGroup(t, N, func(j int) {
		<-blocksSubs[j]
	}, css)
}

func TestReactorReceiveDoesNotPanicIfAddPeerHasntBeenCalledYet(t *testing.T) {
	t.Parallel()

	N := 1
	css, cleanup := randConsensusNet(N, "consensus_reactor_test", newMockTickerFunc(true), newCounter)
	defer cleanup()
	reactors, _, eventSwitches, p2pSwitches := startConsensusNet(t, css, N)
	defer stopConsensusNet(log.NewTestingLogger(t), reactors, eventSwitches, p2pSwitches)

	var (
		reactor = reactors[0]
		peer    = p2pTesting.NewPeer(t)
		msg     = amino.MustMarshalAny(&HasVoteMessage{Height: 1, Round: 1, Index: 1, Type: types.PrevoteType})
	)

	reactor.InitPeer(peer)

	// simulate switch calling Receive before AddPeer
	assert.NotPanics(t, func() {
		reactor.Receive(StateChannel, peer, msg)
		reactor.AddPeer(peer)
	})
}

func TestReactorReceivePanicsIfInitPeerHasntBeenCalledYet(t *testing.T) {
	t.Parallel()

	N := 1
	css, cleanup := randConsensusNet(N, "consensus_reactor_test", newMockTickerFunc(true), newCounter)
	defer cleanup()
	reactors, _, eventSwitches, p2pSwitches := startConsensusNet(t, css, N)
	defer stopConsensusNet(log.NewTestingLogger(t), reactors, eventSwitches, p2pSwitches)

	var (
		reactor = reactors[0]
		peer    = p2pTesting.NewPeer(t)
		msg     = amino.MustMarshalAny(&HasVoteMessage{Height: 1, Round: 1, Index: 1, Type: types.PrevoteType})
	)

	// we should call InitPeer here

	// simulate switch calling Receive before AddPeer
	assert.Panics(t, func() {
		reactor.Receive(StateChannel, peer, msg)
	})
}

// Test we record stats about votes and block parts from other peers.
func TestFlappyReactorRecordsVotesAndBlockParts(t *testing.T) {
	t.Parallel()

	testutils.FilterStability(t, testutils.Flappy)

	N := 4
	css, cleanup := randConsensusNet(N, "consensus_reactor_test", newMockTickerFunc(true), newCounter)
	defer cleanup()
	reactors, blocksSubs, eventSwitches, p2pSwitches := startConsensusNet(t, css, N)
	defer stopConsensusNet(log.NewTestingLogger(t), reactors, eventSwitches, p2pSwitches)

	// wait till everyone makes the first new block
	timeoutWaitGroup(t, N, func(j int) {
		<-blocksSubs[j]
	}, css)

	// Get peer
	peer := reactors[1].Switch.Peers().List()[0]
	// Get peer state
	ps := peer.Get(types.PeerStateKey).(*PeerState)

	assert.Equal(t, true, ps.VotesSent() > 0, "number of votes sent should have increased")
	assert.Equal(t, true, ps.BlockPartsSent() > 0, "number of votes sent should have increased")
}

// -------------------------------------------------------------
// ensure we can make blocks despite cycling a validator set

func TestReactorVotingPowerChange(t *testing.T) {
	t.Parallel()

	nVals := 4
	logger := log.NewTestingLogger(t)
	css, cleanup := randConsensusNet(nVals, "consensus_voting_power_changes_test", newMockTickerFunc(true), newPersistentKVStore)
	defer cleanup()

	reactors, blocksSubs, eventSwitches, p2pSwitches := startConsensusNet(t, css, nVals)
	defer stopConsensusNet(logger, reactors, eventSwitches, p2pSwitches)

	// map of active validators
	activeVals := make(map[string]struct{})
	for i := range nVals {
		addr := css[i].privValidator.PubKey().Address()
		activeVals[addr.String()] = struct{}{}
	}

	// wait till everyone makes block 1
	timeoutWaitGroup(t, nVals, func(j int) {
		<-blocksSubs[j]
	}, css)

	// ---------------------------------------------------------------------------
	logger.Debug("---------------------------- Testing changing the voting power of one validator a few times")

	val1PubKey := css[0].privValidator.PubKey()
	updateValTx := kvstore.MakeValSetChangeTx(val1PubKey, 25)
	previousTotalVotingPower := css[0].GetRoundState().LastValidators.TotalVotingPower()

	waitForAndValidateBlock(t, nVals, activeVals, blocksSubs, css, updateValTx)
	waitForAndValidateBlockWithTx(t, nVals, activeVals, blocksSubs, css, updateValTx)
	waitForAndValidateBlock(t, nVals, activeVals, blocksSubs, css)
	waitForAndValidateBlock(t, nVals, activeVals, blocksSubs, css)

	if css[0].GetRoundState().LastValidators.TotalVotingPower() == previousTotalVotingPower {
		t.Fatalf("expected voting power to change (before: %d, after: %d)", previousTotalVotingPower, css[0].GetRoundState().LastValidators.TotalVotingPower())
	}

	updateValTx = kvstore.MakeValSetChangeTx(val1PubKey, 2)
	previousTotalVotingPower = css[0].GetRoundState().LastValidators.TotalVotingPower()

	waitForAndValidateBlock(t, nVals, activeVals, blocksSubs, css, updateValTx)
	waitForAndValidateBlockWithTx(t, nVals, activeVals, blocksSubs, css, updateValTx)
	waitForAndValidateBlock(t, nVals, activeVals, blocksSubs, css)
	waitForAndValidateBlock(t, nVals, activeVals, blocksSubs, css)

	if css[0].GetRoundState().LastValidators.TotalVotingPower() == previousTotalVotingPower {
		t.Fatalf("expected voting power to change (before: %d, after: %d)", previousTotalVotingPower, css[0].GetRoundState().LastValidators.TotalVotingPower())
	}

	updateValTx = kvstore.MakeValSetChangeTx(val1PubKey, 26)
	previousTotalVotingPower = css[0].GetRoundState().LastValidators.TotalVotingPower()

	waitForAndValidateBlock(t, nVals, activeVals, blocksSubs, css, updateValTx)
	waitForAndValidateBlockWithTx(t, nVals, activeVals, blocksSubs, css, updateValTx)
	waitForAndValidateBlock(t, nVals, activeVals, blocksSubs, css)
	waitForAndValidateBlock(t, nVals, activeVals, blocksSubs, css)

	if css[0].GetRoundState().LastValidators.TotalVotingPower() == previousTotalVotingPower {
		t.Fatalf("expected voting power to change (before: %d, after: %d)", previousTotalVotingPower, css[0].GetRoundState().LastValidators.TotalVotingPower())
	}
}

func TestReactorValidatorSetChanges(t *testing.T) {
	t.Parallel()

	nPeers := 7
	nVals := 4
	css, _, _, cleanup := randConsensusNetWithPeers(nVals, nPeers, "consensus_val_set_changes_test", newMockTickerFunc(true), newPersistentKVStoreWithPath)
	defer cleanup()

	logger := log.NewTestingLogger(t)

	reactors, blocksSubs, eventSwitches, p2pSwitches := startConsensusNet(t, css, nPeers)
	defer stopConsensusNet(logger, reactors, eventSwitches, p2pSwitches)

	// map of active validators
	activeVals := make(map[string]struct{})
	for i := range nVals {
		addr := css[i].privValidator.PubKey().Address()
		activeVals[addr.String()] = struct{}{}
	}

	// wait till everyone makes block 1
	timeoutWaitGroup(t, nPeers, func(j int) {
		<-blocksSubs[j]
	}, css)

	// ---------------------------------------------------------------------------
	logger.Info("---------------------------- Testing adding one validator")

	newValPubKey1 := css[nVals].privValidator.PubKey()
	newValTx1 := kvstore.MakeValSetChangeTx(newValPubKey1, testMinPower)

	// wait till everyone makes block 2
	// ensure the commit includes all validators
	// send newValTx to change vals in block 3
	waitForAndValidateBlock(t, nPeers, activeVals, blocksSubs, css, newValTx1)

	// wait till everyone makes block 3.
	// it includes the commit for block 2, which is by the original validator set
	waitForAndValidateBlockWithTx(t, nPeers, activeVals, blocksSubs, css, newValTx1)

	// wait till everyone makes block 4.
	// it includes the commit for block 3, which is by the original validator set
	waitForAndValidateBlock(t, nPeers, activeVals, blocksSubs, css)

	// the commits for block 4 should be with the updated validator set
	activeVals[newValPubKey1.Address().String()] = struct{}{}

	// wait till everyone makes block 5
	// it includes the commit for block 4, which should have the updated validator set
	waitForBlockWithUpdatedValsAndValidateIt(t, nPeers, activeVals, blocksSubs, css)

	// ---------------------------------------------------------------------------
	logger.Info("---------------------------- Testing changing the voting power of one validator")

	updateValPubKey1 := css[nVals].privValidator.PubKey()
	updateValTx1 := kvstore.MakeValSetChangeTx(updateValPubKey1, 25)
	previousTotalVotingPower := css[nVals].GetRoundState().LastValidators.TotalVotingPower()

	waitForAndValidateBlock(t, nPeers, activeVals, blocksSubs, css, updateValTx1)
	waitForAndValidateBlockWithTx(t, nPeers, activeVals, blocksSubs, css, updateValTx1)
	waitForAndValidateBlock(t, nPeers, activeVals, blocksSubs, css)
	waitForBlockWithUpdatedValsAndValidateIt(t, nPeers, activeVals, blocksSubs, css)

	if css[nVals].GetRoundState().LastValidators.TotalVotingPower() == previousTotalVotingPower {
		t.Errorf("expected voting power to change (before: %d, after: %d)", previousTotalVotingPower, css[nVals].GetRoundState().LastValidators.TotalVotingPower())
	}

	// ---------------------------------------------------------------------------
	logger.Info("---------------------------- Testing adding two validators at once")

	newValPubKey2 := css[nVals+1].privValidator.PubKey()
	newValTx2 := kvstore.MakeValSetChangeTx(newValPubKey2, testMinPower)

	newValPubKey3 := css[nVals+2].privValidator.PubKey()
	newValTx3 := kvstore.MakeValSetChangeTx(newValPubKey3, testMinPower)

	waitForAndValidateBlock(t, nPeers, activeVals, blocksSubs, css, newValTx2, newValTx3)
	waitForAndValidateBlockWithTx(t, nPeers, activeVals, blocksSubs, css, newValTx2, newValTx3)
	waitForAndValidateBlock(t, nPeers, activeVals, blocksSubs, css)
	activeVals[newValPubKey2.Address().String()] = struct{}{}
	activeVals[newValPubKey3.Address().String()] = struct{}{}
	waitForBlockWithUpdatedValsAndValidateIt(t, nPeers, activeVals, blocksSubs, css)

	// ---------------------------------------------------------------------------
	logger.Info("---------------------------- Testing removing two validators at once")

	removeValidatorTx2 := kvstore.MakeValSetChangeTx(newValPubKey2, 0)
	removeValidatorTx3 := kvstore.MakeValSetChangeTx(newValPubKey3, 0)

	waitForAndValidateBlock(t, nPeers, activeVals, blocksSubs, css, removeValidatorTx2, removeValidatorTx3)
	waitForAndValidateBlockWithTx(t, nPeers, activeVals, blocksSubs, css, removeValidatorTx2, removeValidatorTx3)
	waitForAndValidateBlock(t, nPeers, activeVals, blocksSubs, css)
	delete(activeVals, newValPubKey2.Address().String())
	delete(activeVals, newValPubKey3.Address().String())
	waitForBlockWithUpdatedValsAndValidateIt(t, nPeers, activeVals, blocksSubs, css)
}

// Check we can make blocks with skip_timeout_commit=false
func TestReactorWithTimeoutCommit(t *testing.T) {
	t.Parallel()

	N := 4
	css, cleanup := randConsensusNet(N, "consensus_reactor_with_timeout_commit_test", newMockTickerFunc(false), newCounter)
	defer cleanup()
	// override default SkipTimeoutCommit == true for tests
	for i := range N {
		css[i].config.SkipTimeoutCommit = false
	}

	reactors, blocksSubs, eventSwitches, p2pSwitches := startConsensusNet(t, css, N-1)
	defer stopConsensusNet(log.NewTestingLogger(t), reactors, eventSwitches, p2pSwitches)

	// wait till everyone makes the first new block
	timeoutWaitGroup(t, N-1, func(j int) {
		<-blocksSubs[j]
	}, css)
}

func waitForAndValidateBlock(
	t *testing.T,
	n int,
	activeVals map[string]struct{},
	blocksSubs []<-chan events.Event,
	css []*ConsensusState,
	txs ...[]byte,
) {
	t.Helper()

	timeoutWaitGroup(t, n, func(j int) {
		css[j].Logger.Debug("waitForAndValidateBlock")
		msg := <-blocksSubs[j]
		newBlock := msg.(types.EventNewBlock).Block
		css[j].Logger.Debug("waitForAndValidateBlock: Got block", "height", newBlock.Height)
		err := validateBlock(newBlock, activeVals)
		assert.Nil(t, err)
		for _, tx := range txs {
			err := assertMempool(css[j].txNotifier).CheckTx(tx, nil)
			assert.Nil(t, err)
		}
	}, css)
}

func waitForAndValidateBlockWithTx(
	t *testing.T,
	n int,
	activeVals map[string]struct{},
	blocksSubs []<-chan events.Event,
	css []*ConsensusState,
	txs ...[]byte,
) {
	t.Helper()

	timeoutWaitGroup(t, n, func(j int) {
		ntxs := 0
	BLOCK_TX_LOOP:
		for {
			css[j].Logger.Debug("waitForAndValidateBlockWithTx", "ntxs", ntxs)
			msg := <-blocksSubs[j]
			newBlock := msg.(types.EventNewBlock).Block
			css[j].Logger.Debug("waitForAndValidateBlockWithTx: Got block", "height", newBlock.Height)
			err := validateBlock(newBlock, activeVals)
			assert.Nil(t, err)

			// check that txs match the txs we're waiting for.
			// note they could be spread over multiple blocks,
			// but they should be in order.
			for _, tx := range newBlock.Data.Txs {
				assert.EqualValues(t, txs[ntxs], tx)
				ntxs++
			}

			if ntxs == len(txs) {
				break BLOCK_TX_LOOP
			}
		}
	}, css)
}

func waitForBlockWithUpdatedValsAndValidateIt(
	t *testing.T,
	n int,
	updatedVals map[string]struct{},
	blocksSubs []<-chan events.Event,
	css []*ConsensusState,
) {
	t.Helper()

	timeoutWaitGroup(t, n, func(j int) {
		var newBlock *types.Block
	LOOP:
		for {
			css[j].Logger.Debug("waitForBlockWithUpdatedValsAndValidateIt")
			msg := <-blocksSubs[j]
			newBlock = msg.(types.EventNewBlock).Block
			if newBlock.LastCommit.Size() == len(updatedVals) {
				css[j].Logger.Debug("waitForBlockWithUpdatedValsAndValidateIt: Got block", "height", newBlock.Height)
				break LOOP
			} else {
				css[j].Logger.Debug("waitForBlockWithUpdatedValsAndValidateIt: Got block with no new validators. Skipping", "height", newBlock.Height)
			}
		}

		err := validateBlock(newBlock, updatedVals)
		assert.Nil(t, err)
	}, css)
}

// expects high synchrony!
func validateBlock(block *types.Block, activeVals map[string]struct{}) error {
	if block.LastCommit.Size() != len(activeVals) {
		return fmt.Errorf("Commit size doesn't match number of active validators. Got %d, expected %d", block.LastCommit.Size(), len(activeVals))
	}

	for _, vote := range block.LastCommit.Precommits {
		if _, ok := activeVals[vote.ValidatorAddress.String()]; !ok {
			return fmt.Errorf("Found vote for unactive validator %X", vote.ValidatorAddress)
		}
	}
	return nil
}

func timeoutWaitGroup(t *testing.T, n int, f func(int), css []*ConsensusState) {
	t.Helper()

	wg := new(sync.WaitGroup)
	wg.Add(n)
	for i := range n {
		go func(j int) {
			f(j)
			wg.Done()
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// we're running many nodes in-process, possibly in in a virtual machine,
	// and spewing debug messages - making a block could take a while,
	timeout := time.Second * 300

	select {
	case <-done:
	case <-time.After(timeout):
		for i, cs := range css {
			t.Log("#################")
			t.Log("Validator", i)
			t.Log(cs.GetRoundState())
			t.Log("")
		}
		osm.PrintAllGoroutines()
		panic("Timed out waiting for all validators to commit a block")
	}
}

// -------------------------------------------------------------
// Ensure basic validation of structs is functioning

func TestNewRoundStepMessageValidateBasic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName               string
		messageHeight          int64
		messageRound           int
		messageStep            cstypes.RoundStepType
		messageLastCommitRound int
		expectErr              bool
	}{
		{"Valid Message", 0, 0, 0x01, 1, false},
		{"Invalid Message", -1, 0, 0x01, 1, true},
		{"Invalid Message", 0, -1, 0x01, 1, true},
		{"Invalid Message", 0, 0, 0x00, 1, true},
		{"Invalid Message", 0, 0, 0x00, 0, true},
		{"Invalid Message", 1, 0, 0x01, 0, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			message := NewRoundStepMessage{
				Height:          tc.messageHeight,
				Round:           tc.messageRound,
				Step:            tc.messageStep,
				LastCommitRound: tc.messageLastCommitRound,
			}

			assert.Equal(t, tc.expectErr, message.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}
}

func TestNewValidBlockMessageValidateBasic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		malleateFn func(*NewValidBlockMessage)
		expErr     string
	}{
		{func(msg *NewValidBlockMessage) {}, ""},
		{func(msg *NewValidBlockMessage) { msg.Height = -1 }, "Negative Height"},
		{func(msg *NewValidBlockMessage) { msg.Round = -1 }, "Negative Round"},
		{
			func(msg *NewValidBlockMessage) { msg.BlockPartsHeader.Total = 2 },
			"BlockParts bit array size 1 not equal to BlockPartsHeader.Total 2",
		},
		{
			func(msg *NewValidBlockMessage) {
				msg.BlockPartsHeader.Total = 0
				msg.BlockParts = bitarray.NewBitArray(0)
			},
			"Empty BlockParts",
		},
		{
			func(msg *NewValidBlockMessage) { msg.BlockParts = bitarray.NewBitArray(types.MaxBlockPartsCount + 1) },
			"BlockParts bit array size 1602 not equal to BlockPartsHeader.Total 1",
		},
		{
			func(msg *NewValidBlockMessage) { msg.BlockParts.Elems = nil },
			"mismatch between specified number of bits 1, and number of elements 0, expected 1 element",
		},
		{
			func(msg *NewValidBlockMessage) {
				msg.BlockParts.Bits = 500
				msg.BlockPartsHeader.Total = 500 // header total should match bitarray size so ba validation is reached
			},
			"mismatch between specified number of bits 500, and number of elements 1, expected 8 elements",
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()

			msg := &NewValidBlockMessage{
				Height: 1,
				Round:  0,
				BlockPartsHeader: types.PartSetHeader{
					Total: 1,
				},
				BlockParts: bitarray.NewBitArray(1),
			}

			tc.malleateFn(msg)
			err := msg.ValidateBasic()
			if tc.expErr != "" && assert.Error(t, err) {
				assert.Contains(t, err.Error(), tc.expErr)
			}
		})
	}
}

func TestProposalPOLMessageValidateBasic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		malleateFn func(*ProposalPOLMessage)
		expErr     string
	}{
		{func(msg *ProposalPOLMessage) {}, ""},
		{func(msg *ProposalPOLMessage) { msg.Height = -1 }, "Negative Height"},
		{func(msg *ProposalPOLMessage) { msg.ProposalPOLRound = -1 }, "Negative ProposalPOLRound"},
		{func(msg *ProposalPOLMessage) { msg.ProposalPOL = bitarray.NewBitArray(0) }, "Empty ProposalPOL bit array"},
		{
			func(msg *ProposalPOLMessage) { msg.ProposalPOL = bitarray.NewBitArray(types.MaxVotesCount + 1) },
			"ProposalPOL bit array is too big: 10001, max: 10000",
		},
		{
			func(msg *ProposalPOLMessage) { msg.ProposalPOL.Elems = nil },
			"mismatch between specified number of bits 1, and number of elements 0, expected 1 elements",
		},
		{
			func(msg *ProposalPOLMessage) { msg.ProposalPOL.Bits = 500 },
			"mismatch between specified number of bits 500, and number of elements 1, expected 8 elements",
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()

			msg := &ProposalPOLMessage{
				Height:           1,
				ProposalPOLRound: 1,
				ProposalPOL:      bitarray.NewBitArray(1),
			}

			tc.malleateFn(msg)
			err := msg.ValidateBasic()
			if tc.expErr != "" && assert.Error(t, err) {
				assert.Contains(t, err.Error(), tc.expErr)
			}
		})
	}
}

func TestBlockPartMessageValidateBasic(t *testing.T) {
	t.Parallel()

	testPart := new(types.Part)
	testPart.Proof.LeafHash = tmhash.Sum([]byte("leaf"))
	testCases := []struct {
		testName      string
		messageHeight int64
		messageRound  int
		messagePart   *types.Part
		expectErr     bool
	}{
		{"Valid Message", 0, 0, testPart, false},
		{"Invalid Message", -1, 0, testPart, true},
		{"Invalid Message", 0, -1, testPart, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			message := BlockPartMessage{
				Height: tc.messageHeight,
				Round:  tc.messageRound,
				Part:   tc.messagePart,
			}

			assert.Equal(t, tc.expectErr, message.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}

	message := BlockPartMessage{Height: 0, Round: 0, Part: new(types.Part)}
	message.Part.Index = -1

	assert.Equal(t, true, message.ValidateBasic() != nil, "Validate Basic had an unexpected result")
}

func TestHasVoteMessageValidateBasic(t *testing.T) {
	t.Parallel()

	const (
		validSignedMsgType   types.SignedMsgType = 0x01
		invalidSignedMsgType types.SignedMsgType = 0x03
	)

	testCases := []struct {
		testName      string
		messageHeight int64
		messageRound  int
		messageType   types.SignedMsgType
		messageIndex  int
		expectErr     bool
	}{
		{"Valid Message", 0, 0, validSignedMsgType, 0, false},
		{"Invalid Message", -1, 0, validSignedMsgType, 0, true},
		{"Invalid Message", 0, -1, validSignedMsgType, 0, true},
		{"Invalid Message", 0, 0, invalidSignedMsgType, 0, true},
		{"Invalid Message", 0, 0, validSignedMsgType, -1, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			message := HasVoteMessage{
				Height: tc.messageHeight,
				Round:  tc.messageRound,
				Type:   tc.messageType,
				Index:  tc.messageIndex,
			}

			assert.Equal(t, tc.expectErr, message.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}
}

func TestVoteSetMaj23MessageValidateBasic(t *testing.T) {
	t.Parallel()

	const (
		validSignedMsgType   types.SignedMsgType = 0x01
		invalidSignedMsgType types.SignedMsgType = 0x03
	)

	validBlockID := types.BlockID{}
	invalidBlockID := types.BlockID{
		Hash: []byte{},
		PartsHeader: types.PartSetHeader{
			Total: -1,
			Hash:  []byte{},
		},
	}

	testCases := []struct {
		testName       string
		messageHeight  int64
		messageRound   int
		messageType    types.SignedMsgType
		messageBlockID types.BlockID
		expectErr      bool
	}{
		{"Valid Message", 0, 0, validSignedMsgType, validBlockID, false},
		{"Invalid Message", -1, 0, validSignedMsgType, validBlockID, true},
		{"Invalid Message", 0, -1, validSignedMsgType, validBlockID, true},
		{"Invalid Message", 0, 0, invalidSignedMsgType, validBlockID, true},
		{"Invalid Message", 0, 0, validSignedMsgType, invalidBlockID, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			message := VoteSetMaj23Message{
				Height:  tc.messageHeight,
				Round:   tc.messageRound,
				Type:    tc.messageType,
				BlockID: tc.messageBlockID,
			}

			assert.Equal(t, tc.expectErr, message.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}
}

func TestVoteSetBitsMessageValidateBasic(t *testing.T) {
	t.Parallel()

	testCases := []struct { //nolint: maligned
		malleateFn func(*VoteSetBitsMessage)
		expErr     string
	}{
		{func(msg *VoteSetBitsMessage) {}, ""},
		{func(msg *VoteSetBitsMessage) { msg.Height = -1 }, "Negative Height"},
		{func(msg *VoteSetBitsMessage) { msg.Round = -1 }, "Negative Round"},
		{func(msg *VoteSetBitsMessage) { msg.Type = 0x03 }, "Invalid Type"},
		{func(msg *VoteSetBitsMessage) {
			msg.BlockID = types.BlockID{
				Hash: []byte{},
				PartsHeader: types.PartSetHeader{
					Total: -1,
					Hash:  []byte{},
				},
			}
		}, "wrong BlockID: wrong PartsHeader: Negative Total"},
		{
			func(msg *VoteSetBitsMessage) { msg.Votes = bitarray.NewBitArray(types.MaxVotesCount + 1) },
			"votes bit array is too big: 10001, max: 10000",
		},
		{
			func(msg *VoteSetBitsMessage) { msg.Votes.Elems = nil },
			"mismatch between specified number of bits 1, and number of elements 0, expected 1 elements",
		},
		{
			func(msg *VoteSetBitsMessage) { msg.Votes.Bits = 500 },
			"mismatch between specified number of bits 500, and number of elements 1, expected 8 elements",
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()

			msg := &VoteSetBitsMessage{
				Height:  1,
				Round:   0,
				Type:    0x01,
				Votes:   bitarray.NewBitArray(1),
				BlockID: types.BlockID{},
			}

			tc.malleateFn(msg)
			err := msg.ValidateBasic()
			if tc.expErr != "" && assert.Error(t, err) {
				assert.Contains(t, err.Error(), tc.expErr)
			}
		})
	}
}
