package consensus

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	"github.com/gnolang/gno/tm2/pkg/bft/mempool/mock"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	walm "github.com/gnolang/gno/tm2/pkg/bft/wal"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// Functionality to replay blocks and messages on recovery from a crash.
// There are two general failure scenarios:
//
//  1. failure during consensus
//  2. failure while applying the block
//
// The former is handled by the WAL, the latter by the proxyApp Handshake on
// restart, which ultimately hands off the work to the WAL.

// -----------------------------------------
// 1. Recover from failure during consensus
// (by replaying messages from the WAL)
// -----------------------------------------

// Unmarshal and apply a single message to the consensus state as if it were
// received in receiveRoutine.  Lines that start with "#" are ignored.
// NOTE: receiveRoutine should not be running.
func (cs *ConsensusState) readReplayMessage(msg *walm.TimedWALMessage, meta *walm.MetaMessage, newStepSub <-chan events.Event) error {
	// Skip meta messages which exist for demarcating boundaries.
	if meta != nil {
		return nil
	}

	// for logging
	switch m := msg.Msg.(type) {
	case newRoundStepInfo:
		cs.Logger.Info("Replay: New Step", "height", m.Height, "round", m.Round, "step", m.Step)
		// these are playback checks
		ticker := time.After(time.Second * 2)
		if newStepSub != nil {
			select {
			case stepMsg, ok := <-newStepSub:
				if !ok {
					return fmt.Errorf("failed to read off newStepSub. newStepSub was cancelled")
				}
				m2 := stepMsg.(cstypes.EventNewRoundStep)
				if m.Height != m2.Height || m.Round != m2.Round || m.Step != m2.Step {
					return fmt.Errorf("RoundState mismatch. Got %v; Expected %v", m2, m)
				}
			case <-ticker:
				return fmt.Errorf("failed to read off newStepSub")
			}
		}
	case msgInfo:
		peerID := m.PeerID
		if peerID == "" {
			peerID = "local"
		}
		switch msg := m.Msg.(type) {
		case *ProposalMessage:
			p := msg.Proposal
			cs.Logger.Info("Replay: Proposal", "height", p.Height, "round", p.Round, "header",
				p.BlockID.PartsHeader, "pol", p.POLRound, "peer", peerID)
		case *BlockPartMessage:
			cs.Logger.Info("Replay: BlockPart", "height", msg.Height, "round", msg.Round, "peer", peerID)
		case *VoteMessage:
			v := msg.Vote
			cs.Logger.Info("Replay: Vote", "height", v.Height, "round", v.Round, "type", v.Type,
				"blockID", v.BlockID, "peer", peerID)
		}

		cs.handleMsg(m)
	case timeoutInfo:
		cs.Logger.Info("Replay: Timeout", "height", m.Height, "round", m.Round, "step", m.Step, "dur", m.Duration)
		cs.handleTimeout(m, cs.RoundState)
	default:
		return fmt.Errorf("replay: Unknown TimedWALMessage type: %v", reflect.TypeOf(msg.Msg))
	}
	return nil
}

// Replay only those messages since the last block.  `timeoutRoutine` should
// run concurrently to read off tickChan.
func (cs *ConsensusState) catchupReplay(csHeight int64) error {
	// Set replayMode to true so we don't log signing errors.
	cs.replayMode = true
	defer func() { cs.replayMode = false }()

	// Ensure that MetaMessage.Height = height+1 doesn't exist.
	// NOTE: This is just a sanity check. As far as we know things work fine
	// without it, and Handshake could reuse ConsensusState if it weren't for
	// this check (since we can crash after writing #{"h"} (meta height).).
	//
	// Ignore data corruption errors since this is a sanity check.
	gr, found, err := cs.wal.SearchForHeight(csHeight+1, &walm.WALSearchOptions{IgnoreDataCorruptionErrors: true})
	if err != nil {
		return err
	}
	if gr != nil {
		if err := gr.Close(); err != nil {
			return err
		}
	}
	if found {
		return fmt.Errorf("WAL should not contain #ENDHEIGHT %d", csHeight)
	}

	// Search for last height marker.
	//
	// Ignore data corruption errors in previous heights because we only care about last height
	gr, found, err = cs.wal.SearchForHeight(csHeight, &walm.WALSearchOptions{IgnoreDataCorruptionErrors: true})
	if errors.Is(err, io.EOF) {
		cs.Logger.Error("Replay: wal.group.Search returned EOF", "#ENDHEIGHT", csHeight-1)
	} else if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("cannot replay height %d. WAL does not contain #ENDHEIGHT for %d", csHeight, csHeight-1)
	}
	defer gr.Close() //nolint: errcheck

	cs.Logger.Info("Catchup by replaying consensus messages", "height", csHeight)

	dec := walm.NewWALReader(gr, maxMsgSize)

LOOP:
	for {
		msg, meta, err := dec.ReadMessage()
		switch {
		case errors.Is(err, io.EOF):
			break LOOP
		case walm.IsDataCorruptionError(err):
			cs.Logger.Error("data has been corrupted in last height of consensus WAL", "err", err, "height", csHeight)
			return err
		case err != nil:
			return err
		}
		// NOTE: since the priv key is set when the msgs are received
		// it will attempt to eg double sign but we can just ignore it
		// since the votes will be replayed and we'll get to the next step
		if err := cs.readReplayMessage(msg, meta, nil); err != nil {
			return err
		}
	}
	cs.Logger.Info("Replay: Done")
	return nil
}

// --------------------------------------------------------------------------------

// Parses marker lines of the form:
// #ENDHEIGHT: 12345
/*
func makeHeightSearchFunc(height int64) auto.SearchFunc {
	return func(line string) (int, error) {
		line = strings.TrimRight(line, "\n")
		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			return -1, errors.New("Line did not have 2 parts")
		}
		i, err := strconv.Atoi(parts[1])
		if err != nil {
			return -1, errors.New("Failed to parse INFO: " + err.Error())
		}
		if height < i {
			return 1, nil
		} else if height == i {
			return 0, nil
		} else {
			return -1, nil
		}
	}
}*/

// ---------------------------------------------------
// 2. Recover from failure while applying the block.
// (by handshaking with the app to figure out where
// we were last, and using the WAL to recover there.)
// ---------------------------------------------------

type Handshaker struct {
	stateDB      dbm.DB
	initialState sm.State
	store        sm.BlockStore
	evsw         events.EventSwitch
	genDoc       *types.GenesisDoc
	logger       *slog.Logger

	nBlocks int // number of blocks applied to the state
}

func NewHandshaker(stateDB dbm.DB, state sm.State,
	store sm.BlockStore, genDoc *types.GenesisDoc,
) *Handshaker {
	return &Handshaker{
		stateDB:      stateDB,
		initialState: state,
		store:        store,
		evsw:         events.NilEventSwitch(),
		genDoc:       genDoc,
		logger:       log.NewNoopLogger(),
		nBlocks:      0,
	}
}

func (h *Handshaker) SetLogger(l *slog.Logger) {
	h.logger = l
}

// SetEventSwitch - sets the event bus for publishing block related events.
// If not called, it defaults to types.NopEventSwitch.
func (h *Handshaker) SetEventSwitch(evsw events.EventSwitch) {
	h.evsw = evsw
}

// NBlocks returns the number of blocks applied to the state.
func (h *Handshaker) NBlocks() int {
	return h.nBlocks
}

// TODO: retry the handshake/replay if it fails ?
func (h *Handshaker) Handshake(proxyApp appconn.AppConns) error {
	// Handshake is done via ABCI Info on the query conn.
	res, err := proxyApp.Query().InfoSync(abci.RequestInfo{})
	if err != nil {
		return fmt.Errorf("error calling Info: %w", err)
	}

	blockHeight := res.LastBlockHeight
	if blockHeight < 0 {
		return fmt.Errorf("got a negative last block height (%d) from the app", blockHeight)
	}
	appHash := res.LastBlockAppHash

	h.logger.Info("ABCI Handshake App Info",
		"height", blockHeight,
		"hash", fmt.Sprintf("%X", appHash),
		"abci-version", res.ABCIVersion,
		"app-version", res.AppVersion,
	)

	// Set AppVersion on the state.
	if h.initialState.AppVersion != res.AppVersion {
		h.initialState.AppVersion = res.AppVersion
		sm.SaveState(h.stateDB, h.initialState)
	}

	// Replay blocks up to the latest in the blockstore.
	_, err = h.ReplayBlocks(h.initialState, appHash, blockHeight, proxyApp)
	if err != nil {
		return fmt.Errorf("error on replay: %w", err)
	}

	h.logger.Info("Completed ABCI Handshake - Tendermint and App are synced",
		"appHeight", blockHeight, "appHash", fmt.Sprintf("%X", appHash))

	// TODO: (on restart) replay mempool

	return nil
}

// ReplayBlocks replays all blocks since appBlockHeight and ensures the result
// matches the current state.
// Returns the final AppHash or an error.
func (h *Handshaker) ReplayBlocks(
	state sm.State,
	appHash []byte,
	appBlockHeight int64,
	proxyApp appconn.AppConns,
) ([]byte, error) {
	storeBlockHeight := h.store.Height()
	stateBlockHeight := state.LastBlockHeight
	h.logger.Info("ABCI Replay Blocks", "appHeight", appBlockHeight, "storeHeight", storeBlockHeight, "stateHeight", stateBlockHeight)

	// If appBlockHeight == 0 it means that we are at genesis and hence should send InitChain.
	if appBlockHeight == 0 {
		validators := make([]*types.Validator, len(h.genDoc.Validators))
		for i, val := range h.genDoc.Validators {
			validators[i] = types.NewValidator(val.PubKey, val.Power)
		}
		validatorSet := types.NewValidatorSet(validators)
		nextVals := validatorSet.ABCIValidatorUpdates()
		csParams := h.genDoc.ConsensusParams
		req := abci.RequestInitChain{
			Time:            h.genDoc.GenesisTime,
			ChainID:         h.genDoc.ChainID,
			ConsensusParams: &csParams,
			Validators:      nextVals,
			AppState:        h.genDoc.AppState,
		}
		res, err := proxyApp.Consensus().InitChainSync(req)
		if err != nil {
			return nil, err
		}

		// Save the results by height
		abciResponse := sm.NewABCIResponsesFromNum(int64(len(res.TxResponses)))
		copy(abciResponse.DeliverTxs, res.TxResponses)
		sm.SaveABCIResponses(h.stateDB, 0, abciResponse)

		// NOTE: we don't save results by tx hash since the transactions are in the AppState opaque type

		if stateBlockHeight == 0 { // we only update state when we are in initial state
			// If the app returned validators or consensus params, update the state.
			if len(res.Validators) > 0 {
				vals := types.NewValidatorSetFromABCIValidatorUpdates(res.Validators)
				state.Validators = vals
				state.NextValidators = vals.Copy()
			} else if len(h.genDoc.Validators) == 0 {
				// If validator set is not set in genesis and still empty after InitChain, exit.
				return nil, fmt.Errorf("validator set is nil in genesis and still empty after InitChain")
			}

			if res.ConsensusParams != nil {
				state.ConsensusParams = state.ConsensusParams.Update(*res.ConsensusParams)
			}
			sm.SaveState(h.stateDB, state)
		}
	}

	// First handle edge cases and constraints on the storeBlockHeight.
	switch {
	case storeBlockHeight == 0:
		assertAppHashEqualsOneFromState(appHash, state)
		return appHash, nil

	case storeBlockHeight < appBlockHeight:
		// the app should never be ahead of the store (but this is under app's control)
		return appHash, sm.AppBlockHeightTooHighError{CoreHeight: storeBlockHeight, AppHeight: appBlockHeight}

	case storeBlockHeight < stateBlockHeight:
		// the state should never be ahead of the store (this is under tendermint's control)
		panic(fmt.Sprintf("StateBlockHeight (%d) > StoreBlockHeight (%d)", stateBlockHeight, storeBlockHeight))

	case storeBlockHeight > stateBlockHeight+1:
		// store should be at most one ahead of the state (this is under tendermint's control)
		panic(fmt.Sprintf("StoreBlockHeight (%d) > StateBlockHeight + 1 (%d)", storeBlockHeight, stateBlockHeight+1))
	}

	var err error
	// Now either store is equal to state, or one ahead.
	// For each, consider all cases of where the app could be, given app <= store
	switch storeBlockHeight {
	case stateBlockHeight:
		// Tendermint ran Commit and saved the state.
		// Either the app is asking for replay, or we're all synced up.
		if appBlockHeight < storeBlockHeight {
			// the app is behind, so replay blocks, but no need to go through WAL (state is already synced to store)
			return h.replayBlocks(state, proxyApp, appBlockHeight, storeBlockHeight, false)
		} else if appBlockHeight == storeBlockHeight {
			// We're good!
			assertAppHashEqualsOneFromState(appHash, state)
			return appHash, nil
		}
	case stateBlockHeight + 1:
		// We saved the block in the store but haven't updated the state,
		// so we'll need to replay a block using the WAL.
		switch {
		case appBlockHeight < stateBlockHeight:
			// the app is further behind than it should be, so replay blocks
			// but leave the last block to go through the WAL
			return h.replayBlocks(state, proxyApp, appBlockHeight, storeBlockHeight, true)

		case appBlockHeight == stateBlockHeight:
			// We haven't run Commit (both the state and app are one block behind),
			// so replayBlock with the real app.
			// NOTE: We could instead use the cs.WAL on cs.Start,
			// but we'd have to allow the WAL to replay a block that wrote it's #ENDHEIGHT
			h.logger.Info("Replay last block using real app")
			state, err = h.replayBlock(state, storeBlockHeight, proxyApp.Consensus())
			return state.AppHash, err

		case appBlockHeight == storeBlockHeight:
			// We ran Commit, but didn't save the state, so replayBlock with mock app.
			abciResponses, err := sm.LoadABCIResponses(h.stateDB, storeBlockHeight)
			if err != nil {
				return nil, err
			}
			mockApp := newMockProxyApp(appHash, abciResponses)
			h.logger.Info("Replay last block using mock app")
			state, err = h.replayBlock(state, storeBlockHeight, mockApp)
			return state.AppHash, err
		}
	}

	panic(fmt.Sprintf("uncovered case! appHeight: %d, storeHeight: %d, stateHeight: %d",
		appBlockHeight, storeBlockHeight, stateBlockHeight))
}

func (h *Handshaker) replayBlocks(state sm.State, proxyApp appconn.AppConns, appBlockHeight, storeBlockHeight int64, mutateState bool) ([]byte, error) {
	// App is further behind than it should be, so we need to replay blocks.
	// We replay all blocks from appBlockHeight+1.
	//
	// Note that we don't have an old version of the state,
	// so we by-pass state validation/mutation using sm.ExecCommitBlock.
	// This also means we won't be saving validator sets if they change during this period.
	// TODO: Load the historical information to fix this and just use state.ApplyBlock
	//
	// If mutateState == true, the final block is replayed with h.replayBlock()

	var appHash []byte
	var err error
	finalBlock := storeBlockHeight
	if mutateState {
		finalBlock--
	}
	for i := appBlockHeight + 1; i <= finalBlock; i++ {
		h.logger.Info("Applying block", "height", i)
		block := h.store.LoadBlock(i)
		if block == nil {
			return nil, fmt.Errorf("block not found for height %d", i)
		}
		// Extra check to ensure the app was not changed in a way it shouldn't have.
		if len(appHash) > 0 {
			assertAppHashEqualsOneFromBlock(appHash, block)
		}

		appHash, err = sm.ExecCommitBlock(proxyApp.Consensus(), block, h.logger, h.stateDB)
		if err != nil {
			return nil, err
		}

		h.nBlocks++
	}

	if mutateState {
		// sync the final block
		state, err = h.replayBlock(state, storeBlockHeight, proxyApp.Consensus())
		if err != nil {
			return nil, err
		}
		appHash = state.AppHash
	}

	assertAppHashEqualsOneFromState(appHash, state)
	return appHash, nil
}

// ApplyBlock on the proxyApp with the last block.
func (h *Handshaker) replayBlock(state sm.State, height int64, proxyApp appconn.Consensus) (sm.State, error) {
	block := h.store.LoadBlock(height)
	if block == nil {
		return sm.State{}, fmt.Errorf("block not found for height %d", height)
	}
	meta := h.store.LoadBlockMeta(height)
	if meta == nil {
		return sm.State{}, fmt.Errorf("block meta not found for height %d", height)
	}

	blockExec := sm.NewBlockExecutor(h.stateDB, h.logger, proxyApp, mock.Mempool{})
	blockExec.SetEventSwitch(h.evsw)

	var err error
	state, err = blockExec.ApplyBlock(state, meta.BlockID, block)
	if err != nil {
		return sm.State{}, err
	}

	h.nBlocks++

	return state, nil
}

func assertAppHashEqualsOneFromBlock(appHash []byte, block *types.Block) {
	if !bytes.Equal(appHash, block.AppHash) {
		panic(fmt.Sprintf(`block.AppHash does not match AppHash after replay. Got %X, expected %X.

Block: %v
`,
			appHash, block.AppHash, block))
	}
}

func assertAppHashEqualsOneFromState(appHash []byte, state sm.State) {
	if !bytes.Equal(appHash, state.AppHash) {
		panic(fmt.Sprintf(`state.AppHash does not match AppHash after replay. Got
%X, expected %X.

State: %v

Did you reset Tendermint without resetting your application's data?`,
			appHash, state.AppHash, state))
	}
}

// --------------------------------------------------------------------------------
// mockProxyApp uses ABCIResponses to give the right results
// Useful because we don't want to call Commit() twice for the same block on the real app.

func newMockProxyApp(appHash []byte, abciResponses *sm.ABCIResponses) appconn.Consensus {
	clientCreator := proxy.NewLocalClientCreator(&mockProxyApp{
		appHash:       appHash,
		abciResponses: abciResponses,
	})
	cli, _ := clientCreator.NewABCIClient()
	err := cli.Start()
	if err != nil {
		panic(err)
	}
	return appconn.Consensus(cli)
}

type mockProxyApp struct {
	abci.BaseApplication

	appHash       []byte
	txCount       int
	abciResponses *sm.ABCIResponses
}

func (mock *mockProxyApp) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	r := mock.abciResponses.DeliverTxs[mock.txCount]
	mock.txCount++
	return r
}

func (mock *mockProxyApp) EndBlock(req abci.RequestEndBlock) abci.ResponseEndBlock {
	mock.txCount = 0
	return mock.abciResponses.EndBlock
}

func (mock *mockProxyApp) Commit() (res abci.ResponseCommit) {
	res.Data = mock.appHash
	return
}
