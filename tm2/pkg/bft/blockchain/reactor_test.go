package blockchain

import (
	"context"
	"log/slog"
	"os"
	"sort"
	"testing"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/mempool/mock"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/errors"
	p2pTesting "github.com/gnolang/gno/tm2/pkg/internal/p2p"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var config *cfg.Config

func randGenesisDoc(numValidators int, randPower bool, minPower int64) (*types.GenesisDoc, []types.PrivValidator) {
	validators := make([]types.GenesisValidator, numValidators)
	privValidators := make([]types.PrivValidator, numValidators)
	for i := range numValidators {
		val, privVal := types.RandValidator(randPower, minPower)
		validators[i] = types.GenesisValidator{
			PubKey: val.PubKey,
			Power:  val.VotingPower,
		}
		privValidators[i] = privVal
	}
	sort.Sort(types.PrivValidatorsByAddress(privValidators))

	return &types.GenesisDoc{
		GenesisTime: tmtime.Now(),
		ChainID:     config.ChainID(),
		Validators:  validators,
	}, privValidators
}

type BlockchainReactorPair struct {
	reactor *BlockchainReactor
	app     appconn.AppConns
}

func newBlockchainReactor(logger *slog.Logger, genDoc *types.GenesisDoc, privVals []types.PrivValidator, maxBlockHeight int64) BlockchainReactorPair {
	if len(privVals) != 1 {
		panic("only support one validator")
	}

	app := &testApp{}
	cc := proxy.NewLocalClientCreator(app)
	proxyApp := appconn.NewAppConns(cc)
	err := proxyApp.Start()
	if err != nil {
		panic(errors.Wrap(err, "error start app"))
	}

	blockDB := memdb.NewMemDB()
	stateDB := memdb.NewMemDB()
	blockStore := store.NewBlockStore(blockDB)

	state, err := sm.LoadStateFromDBOrGenesisDoc(stateDB, genDoc)
	if err != nil {
		panic(errors.Wrap(err, "error constructing state from genesis file"))
	}

	// Make the BlockchainReactor itself.
	// NOTE we have to create and commit the blocks first because
	// pool.height is determined from the store.
	fastSync := true
	db := memdb.NewMemDB()
	blockExec := sm.NewBlockExecutor(db, logger, proxyApp.Consensus(), mock.Mempool{})
	sm.SaveState(db, state)

	// let's add some blocks in
	for blockHeight := int64(1); blockHeight <= maxBlockHeight; blockHeight++ {
		lastCommit := types.NewCommit(types.BlockID{}, nil)
		if blockHeight > 1 {
			lastBlockMeta := blockStore.LoadBlockMeta(blockHeight - 1)
			lastBlock := blockStore.LoadBlock(blockHeight - 1)

			vote, err := types.MakeVote(lastBlock.Header.Height, lastBlockMeta.BlockID, state.Validators, privVals[0], lastBlock.Header.ChainID)
			if err != nil {
				panic(err)
			}
			voteCommitSig := vote.CommitSig()
			lastCommit = types.NewCommit(lastBlockMeta.BlockID, []*types.CommitSig{voteCommitSig})
		}

		thisBlock := makeBlock(blockHeight, state, lastCommit)

		thisParts := thisBlock.MakePartSet(types.BlockPartSizeBytes)
		blockID := types.BlockID{Hash: thisBlock.Hash(), PartsHeader: thisParts.Header()}

		state, err = blockExec.ApplyBlock(state, blockID, thisBlock)
		if err != nil {
			panic(errors.Wrap(err, "error apply block"))
		}

		blockStore.SaveBlock(thisBlock, thisParts, lastCommit)
	}

	bcReactor := NewBlockchainReactor(state.Copy(), blockExec, blockStore, fastSync, nil)
	bcReactor.SetLogger(logger.With("module", "blockchain"))

	return BlockchainReactorPair{bcReactor, proxyApp}
}

func TestNoBlockResponse(t *testing.T) {
	t.Parallel()

	config, _ = cfg.ResetTestRoot("blockchain_reactor_test")
	defer os.RemoveAll(config.RootDir)
	genDoc, privVals := randGenesisDoc(1, false, 30)

	maxBlockHeight := int64(65)

	var (
		reactorPairs = make([]BlockchainReactorPair, 2)
		options      = make(map[int][]p2p.SwitchOption)
	)

	for i := range reactorPairs {
		height := int64(0)
		if i == 0 {
			height = maxBlockHeight
		}

		reactorPairs[i] = newBlockchainReactor(log.NewTestingLogger(t), genDoc, privVals, height)

		options[i] = []p2p.SwitchOption{
			p2p.WithReactor("BLOCKCHAIN", reactorPairs[i].reactor),
		}
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	testingCfg := p2pTesting.TestingConfig{
		Count:         2,
		P2PCfg:        config.P2P,
		SwitchOptions: options,
		Channels:      []byte{BlockchainChannel},
	}

	p2pTesting.MakeConnectedPeers(t, ctx, testingCfg)

	defer func() {
		for _, r := range reactorPairs {
			r.reactor.Stop()
			r.app.Stop()
		}
	}()

	tests := []struct {
		height   int64
		existent bool
	}{
		{maxBlockHeight + 2, false},
		{10, true},
		{1, true},
		{100, false},
	}

	for !reactorPairs[1].reactor.pool.IsCaughtUp() {
		time.Sleep(10 * time.Millisecond)
	}

	assert.Equal(t, maxBlockHeight, reactorPairs[0].reactor.store.Height())

	for _, tt := range tests {
		block := reactorPairs[1].reactor.store.LoadBlock(tt.height)
		if tt.existent {
			assert.True(t, block != nil)
		} else {
			assert.True(t, block == nil)
		}
	}
}

// NOTE: This is too hard to test without
// an easy way to add test peer to switch
// or without significant refactoring of the module.
// Alternatively we could actually dial a TCP conn but
// that seems extreme.
func TestFlappyBadBlockStopsPeer(t *testing.T) {
	t.Parallel()

	testutils.FilterStability(t, testutils.Flappy)

	config, _ = cfg.ResetTestRoot("blockchain_reactor_test")
	defer os.RemoveAll(config.RootDir)
	genDoc, privVals := randGenesisDoc(1, false, 30)

	maxBlockHeight := int64(148)

	otherChain := newBlockchainReactor(log.NewNoopLogger(), genDoc, privVals, maxBlockHeight)
	defer func() {
		otherChain.reactor.Stop()
		otherChain.app.Stop()
	}()

	var (
		reactorPairs = make([]BlockchainReactorPair, 4)
		options      = make(map[int][]p2p.SwitchOption)
	)

	for i := range reactorPairs {
		height := int64(0)
		if i == 0 {
			height = maxBlockHeight
		}

		reactorPairs[i] = newBlockchainReactor(log.NewNoopLogger(), genDoc, privVals, height)

		options[i] = []p2p.SwitchOption{
			p2p.WithReactor("BLOCKCHAIN", reactorPairs[i].reactor),
		}
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	testingCfg := p2pTesting.TestingConfig{
		Count:         4,
		P2PCfg:        config.P2P,
		SwitchOptions: options,
		Channels:      []byte{BlockchainChannel},
	}

	_, transports := p2pTesting.MakeConnectedPeers(t, ctx, testingCfg)

	defer func() {
		for _, r := range reactorPairs {
			r.reactor.Stop()
			r.app.Stop()
		}
	}()

	for !reactorPairs[3].reactor.pool.IsCaughtUp() {
		time.Sleep(1 * time.Second)
	}

	// at this time, reactors[0-3] is the newest
	assert.Equal(t, 3, len(reactorPairs[1].reactor.Switch.Peers().List()))

	// mark reactorPairs[3] is an invalid peer
	reactorPairs[3].reactor.store = otherChain.reactor.store

	lastReactorPair := newBlockchainReactor(log.NewNoopLogger(), genDoc, privVals, 0)
	reactorPairs = append(reactorPairs, lastReactorPair)

	persistentPeers := make([]*p2pTypes.NetAddress, 0, len(transports))

	for _, tr := range transports {
		addr := tr.NetAddress()
		persistentPeers = append(persistentPeers, &addr)
	}

	for i, opt := range options {
		opt = append(opt, p2p.WithPersistentPeers(persistentPeers))

		options[i] = opt
	}

	ctx, cancelFn = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	testingCfg = p2pTesting.TestingConfig{
		Count:         1,
		P2PCfg:        config.P2P,
		SwitchOptions: options,
		Channels:      []byte{BlockchainChannel},
	}

	p2pTesting.MakeConnectedPeers(t, ctx, testingCfg)

	for !lastReactorPair.reactor.pool.IsCaughtUp() && len(lastReactorPair.reactor.Switch.Peers().List()) != 0 {
		time.Sleep(1 * time.Second)
	}

	assert.True(t, len(lastReactorPair.reactor.Switch.Peers().List()) < len(reactorPairs)-1)
}

func TestBcBlockRequestMessageValidateBasic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName      string
		requestHeight int64
		expectErr     bool
	}{
		{"Valid Request Message", 0, false},
		{"Valid Request Message", 1, false},
		{"Invalid Request Message", -1, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			request := bcBlockRequestMessage{Height: tc.requestHeight}
			assert.Equal(t, tc.expectErr, request.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}
}

func TestBcNoBlockResponseMessageValidateBasic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName          string
		nonResponseHeight int64
		expectErr         bool
	}{
		{"Valid Non-Response Message", 0, false},
		{"Valid Non-Response Message", 1, false},
		{"Invalid Non-Response Message", -1, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			nonResponse := bcNoBlockResponseMessage{Height: tc.nonResponseHeight}
			assert.Equal(t, tc.expectErr, nonResponse.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}
}

func TestBcStatusRequestMessageValidateBasic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName      string
		requestHeight int64
		expectErr     bool
	}{
		{"Valid Request Message", 0, false},
		{"Valid Request Message", 1, false},
		{"Invalid Request Message", -1, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			request := bcStatusRequestMessage{Height: tc.requestHeight}
			assert.Equal(t, tc.expectErr, request.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}
}

func TestBcStatusResponseMessageValidateBasic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName       string
		responseHeight int64
		expectErr      bool
	}{
		{"Valid Response Message", 0, false},
		{"Valid Response Message", 1, false},
		{"Invalid Response Message", -1, true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testName, func(t *testing.T) {
			t.Parallel()

			response := bcStatusResponseMessage{Height: tc.responseHeight}
			assert.Equal(t, tc.expectErr, response.ValidateBasic() != nil, "Validate Basic had an unexpected result")
		})
	}
}

func TestRestore(t *testing.T) {
	t.Parallel()

	config, _ = cfg.ResetTestRoot("blockchain_reactor_test")
	defer os.RemoveAll(config.RootDir)
	genDoc, privVals := randGenesisDoc(1, false, 30)

	logger := log.NewNoopLogger()

	reactor := newBlockchainReactor(logger, genDoc, privVals, 0)

	stateDB := memdb.NewMemDB()
	state, err := sm.LoadStateFromDBOrGenesisDoc(stateDB, genDoc)
	require.NoError(t, err)

	app := &testApp{}
	cc := proxy.NewLocalClientCreator(app)
	proxyApp := appconn.NewAppConns(cc)
	require.NoError(t, proxyApp.Start())

	// we generate blocks using another executor and then restore using the test reactor that has it's own executor
	db := memdb.NewMemDB()
	blockExec := sm.NewBlockExecutor(db, logger, proxyApp.Consensus(), mock.Mempool{})
	sm.SaveState(db, state)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		lastBlock     *types.Block
		lastBlockMeta *types.BlockMeta
		blockHeight   int64 = 1
	)
	generateBlock := func() *types.Block {
		t.Helper()

		lastCommit := types.NewCommit(types.BlockID{}, nil)
		if blockHeight > 1 {
			vote, err := types.MakeVote(lastBlock.Header.Height, lastBlockMeta.BlockID, state.Validators, privVals[0], lastBlock.Header.ChainID)
			require.NoError(t, err)
			voteCommitSig := vote.CommitSig()
			lastCommit = types.NewCommit(lastBlockMeta.BlockID, []*types.CommitSig{voteCommitSig})
		}

		thisBlock := makeBlock(blockHeight, state, lastCommit)

		thisParts := thisBlock.MakePartSet(types.BlockPartSizeBytes)
		blockID := types.BlockID{Hash: thisBlock.Hash(), PartsHeader: thisParts.Header()}

		state, err = blockExec.ApplyBlock(state, blockID, thisBlock)
		require.NoError(t, err)

		lastBlock = thisBlock
		lastBlockMeta = &types.BlockMeta{BlockID: blockID, Header: lastBlock.Header}

		blockHeight++

		return thisBlock
	}

	numBlocks := 50

	err = reactor.reactor.Restore(ctx, func(yield func(block *types.Block) error) error {
		for range numBlocks {
			block := generateBlock()

			err := yield(block)
			require.NoError(t, err)

			if blockHeight > 2 {
				require.NotNil(t, reactor.reactor.store.LoadBlock(blockHeight-2))
			}

			require.Equal(t, int64(blockHeight-2), reactor.reactor.store.Height())
		}
		return nil
	})
	require.NoError(t, err)
}

// ----------------------------------------------
// utility funcs

func makeTxs(height int64) (txs []types.Tx) {
	for i := range 10 {
		txs = append(txs, types.Tx([]byte{byte(height), byte(i)}))
	}
	return txs
}

func makeBlock(height int64, state sm.State, lastCommit *types.Commit) *types.Block {
	block, _ := state.MakeBlock(height, makeTxs(height), lastCommit, state.Validators.GetProposer().Address)
	return block
}

type testApp struct {
	abci.BaseApplication
}

var _ abci.Application = (*testApp)(nil)

func (app *testApp) Info(req abci.RequestInfo) (resInfo abci.ResponseInfo) {
	return abci.ResponseInfo{}
}

func (app *testApp) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return abci.ResponseBeginBlock{}
}

func (app *testApp) EndBlock(req abci.RequestEndBlock) abci.ResponseEndBlock {
	return abci.ResponseEndBlock{}
}

func (app *testApp) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	return abci.ResponseDeliverTx{ResponseBase: abci.ResponseBase{Events: []abci.Event{}}}
}

func (app *testApp) CheckTx(req abci.RequestCheckTx) abci.ResponseCheckTx {
	return abci.ResponseCheckTx{}
}

func (app *testApp) Commit() abci.ResponseCommit {
	return abci.ResponseCommit{}
}

func (app *testApp) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	return
}
