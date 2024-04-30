package state

import (
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/bft/fail"
	mempl "github.com/gnolang/gno/tm2/pkg/bft/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	typesver "github.com/gnolang/gno/tm2/pkg/bft/types/version"
	tmver "github.com/gnolang/gno/tm2/pkg/bft/version"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/events"
)

// -----------------------------------------------------------------------------
// BlockExecutor handles block execution and state updates.
// It exposes ApplyBlock(), which validates & executes the block, updates state w/ ABCI responses,
// then commits and updates the mempool atomically, then saves state.

// BlockExecutor provides the context and accessories for properly executing a block.
type BlockExecutor struct {
	// save state, validators, consensus params, abci responses here
	db dbm.DB

	// execute the app against this
	proxyApp appconn.Consensus

	// events
	evsw events.EventSwitch

	// manage the mempool lock during commit
	// and update both with block results after commit.
	mempool mempl.Mempool

	logger *slog.Logger
}

type BlockExecutorOption func(executor *BlockExecutor)

// NewBlockExecutor returns a new BlockExecutor with a NopEventBus.
// Call SetEventBus to provide one.
func NewBlockExecutor(db dbm.DB, logger *slog.Logger, proxyApp appconn.Consensus, mempool mempl.Mempool, options ...BlockExecutorOption) *BlockExecutor {
	res := &BlockExecutor{
		db:       db,
		proxyApp: proxyApp,
		evsw:     events.NilEventSwitch(),
		mempool:  mempool,
		logger:   logger,
	}

	for _, option := range options {
		option(res)
	}

	return res
}

func (blockExec *BlockExecutor) DB() dbm.DB {
	return blockExec.db
}

func (blockExec *BlockExecutor) SetEventSwitch(evsw events.EventSwitch) {
	blockExec.evsw = evsw
}

// CreateProposalBlock calls state.MakeBlock with txs from the mempool.
func (blockExec *BlockExecutor) CreateProposalBlock(
	height int64,
	state State, commit *types.Commit,
	proposerAddr crypto.Address,
) (*types.Block, *types.PartSet) {
	maxDataBytes := state.ConsensusParams.Block.MaxDataBytes
	maxGas := state.ConsensusParams.Block.MaxGas

	txs := blockExec.mempool.ReapMaxBytesMaxGas(maxDataBytes, maxGas)

	return state.MakeBlock(height, txs, commit, proposerAddr)
}

// ValidateBlock validates the given block against the given state.
// If the block is invalid, it returns an error.
// Validation does not mutate state, but does require historical information from the stateDB
func (blockExec *BlockExecutor) ValidateBlock(state State, block *types.Block) error {
	return validateBlock(blockExec.db, state, block)
}

// ApplyBlock validates the block against the state, executes it against the app,
// fires the relevant events, commits the app, and saves the new state and responses.
// It's the only function that needs to be called
// from outside this package to process and commit an entire block.
// It takes a blockID to avoid recomputing the parts hash.
func (blockExec *BlockExecutor) ApplyBlock(state State, blockID types.BlockID, block *types.Block) (State, error) {
	if err := blockExec.ValidateBlock(state, block); err != nil {
		return state, InvalidBlockError(err)
	}

	abciResponses, err := execBlockOnProxyApp(blockExec.logger, blockExec.proxyApp, block, blockExec.db)
	if err != nil {
		return state, ProxyAppConnError(err)
	}

	fail.Fail() // XXX

	// Save the results before we commit.
	saveABCIResponses(blockExec.db, block.Height, abciResponses)

	// Save the transaction results
	for index, tx := range block.Txs {
		saveTxResultIndex(
			blockExec.db,
			tx.Hash(),
			TxResultIndex{
				BlockNum: block.Height,
				TxIndex:  uint32(index),
			},
		)
	}

	fail.Fail() // XXX

	// validate the validator updates and convert to tendermint types
	abciValUpdates := abciResponses.EndBlock.ValidatorUpdates
	err = validateValidatorUpdates(abciValUpdates, *state.ConsensusParams.Validator)
	if err != nil {
		return state, fmt.Errorf("Error in validator updates: %w", err)
	}
	if len(abciValUpdates) > 0 {
		blockExec.logger.Info("Updates to validators", "updates", abciValUpdates)
	}

	// Update the state with the block and responses.
	state, err = updateState(state, blockID, &block.Header, abciResponses)
	if err != nil {
		return state, fmt.Errorf("Commit failed for application: %w", err)
	}

	// Lock mempool, commit app state, update mempoool.
	appHash, err := blockExec.Commit(state, block, abciResponses.DeliverTxs)
	if err != nil {
		return state, fmt.Errorf("Commit failed for application: %w", err)
	}

	fail.Fail() // XXX

	// Update the app hash and save the state.
	state.AppHash = appHash
	SaveState(blockExec.db, state)

	fail.Fail() // XXX

	// Events are fired after everything else.
	// NOTE: if we crash between Commit and Save, events wont be fired during replay
	fireEvents(blockExec.evsw, block, abciResponses)

	return state, nil
}

// Commit locks the mempool, runs the ABCI Commit message, and updates the
// mempool.
// It returns the result of calling abci.Commit (the AppHash), and an error.
// The Mempool must be locked during commit and update because state is
// typically reset on Commit and old txs must be replayed against committed
// state before new txs are run in the mempool, lest they be invalid.
func (blockExec *BlockExecutor) Commit(
	state State,
	block *types.Block,
	deliverTxResponses []abci.ResponseDeliverTx,
) ([]byte, error) {
	blockExec.mempool.Lock()
	defer blockExec.mempool.Unlock()

	// while mempool is Locked, flush to ensure all async requests have completed
	// in the ABCI app before Commit.
	err := blockExec.mempool.FlushAppConn()
	if err != nil {
		blockExec.logger.Error("Client error during mempool.FlushAppConn", "err", err)
		return nil, err
	}

	// Commit block, get hash back
	res, err := blockExec.proxyApp.CommitSync()
	if err != nil {
		blockExec.logger.Error(
			"Client error during proxyAppConn.CommitSync",
			"err", err,
		)
		return nil, err
	}
	// ResponseCommit has no error code - just data

	blockExec.logger.Info(
		"Committed state",
		"height", block.Height,
		"txs", block.NumTxs,
		"appHash", fmt.Sprintf("%X", res.Data),
	)

	// Update mempool.
	err = blockExec.mempool.Update(
		block.Height,
		block.Txs,
		deliverTxResponses,
		TxPreCheck(state),
		state.ConsensusParams.Block.MaxTxBytes,
	)

	return res.Data, err
}

// ---------------------------------------------------------
// Helper functions for executing blocks and updating state

// Executes block's transactions on proxyAppConn.
// Returns a list of transaction results and updates to the validator set
func execBlockOnProxyApp(
	logger *slog.Logger,
	proxyAppConn appconn.Consensus,
	block *types.Block,
	stateDB dbm.DB,
) (*ABCIResponses, error) {
	validTxs, invalidTxs := 0, 0

	txIndex := 0
	abciResponses := NewABCIResponses(block)

	// Execute transactions and get hash.
	proxyCb := func(req abci.Request, res abci.Response) {
		if res, ok := res.(abci.ResponseDeliverTx); ok {
			// TODO: make use of res.Log
			// TODO: make use of this info
			// Blocks may include invalid txs.
			if res.Error == nil {
				validTxs++
			} else {
				logger.Debug("Invalid tx", "error", res.Error, "log", res.Log)
				invalidTxs++
			}
			abciResponses.DeliverTxs[txIndex] = res
			txIndex++
		}
	}
	proxyAppConn.SetResponseCallback(proxyCb)

	commitInfo := getBeginBlockLastCommitInfo(block, stateDB)

	// Begin block
	var err error
	abciResponses.BeginBlock, err = proxyAppConn.BeginBlockSync(abci.RequestBeginBlock{
		Hash:           block.Hash(),
		Header:         block.Header.Copy(),
		LastCommitInfo: &commitInfo,
	})
	if err != nil {
		logger.Error("Error in proxyAppConn.BeginBlock", "err", err)
		return nil, err
	}

	// Run txs of block.
	for _, tx := range block.Txs {
		proxyAppConn.DeliverTxAsync(abci.RequestDeliverTx{Tx: tx})
		if err := proxyAppConn.Error(); err != nil {
			return nil, err
		}
	}

	// End block.
	abciResponses.EndBlock, err = proxyAppConn.EndBlockSync(abci.RequestEndBlock{Height: block.Height})
	if err != nil {
		logger.Error("Error in proxyAppConn.EndBlock", "err", err)
		return nil, err
	}

	logger.Info("Executed block", "height", block.Height, "validTxs", validTxs, "invalidTxs", invalidTxs)

	return abciResponses, nil
}

func getBeginBlockLastCommitInfo(block *types.Block, stateDB dbm.DB) abci.LastCommitInfo {
	voteInfos := make([]abci.VoteInfo, block.LastCommit.Size())
	var lastValSet *types.ValidatorSet
	var err error
	if block.Height > 1 {
		lastValSet, err = LoadValidators(stateDB, block.Height-1)
		if err != nil {
			panic(err) // shouldn't happen
		}

		// Sanity check that commit length matches validator set size -
		// only applies after first block

		precommitLen := block.LastCommit.Size()
		valSetLen := len(lastValSet.Validators)
		if precommitLen != valSetLen {
			// sanity check
			panic(fmt.Sprintf("precommit length (%d) doesn't match valset length (%d) at height %d\n\n%v\n\n%v",
				precommitLen, valSetLen, block.Height, block.LastCommit.Precommits, lastValSet.Validators))
		}
	} else {
		lastValSet = types.NewValidatorSet(nil)
	}

	for i, val := range lastValSet.Validators {
		var vote *types.CommitSig
		if i < len(block.LastCommit.Precommits) {
			vote = block.LastCommit.Precommits[i]
		}
		voteInfo := abci.VoteInfo{
			Address:         val.Address,
			Power:           val.VotingPower,
			SignedLastBlock: vote != nil,
		}
		voteInfos[i] = voteInfo
	}

	commitInfo := abci.LastCommitInfo{
		Round: int32(block.LastCommit.Round()),
		Votes: voteInfos,
	}
	return commitInfo
}

func validateValidatorUpdates(abciUpdates []abci.ValidatorUpdate,
	params abci.ValidatorParams,
) error {
	for _, valUpdate := range abciUpdates {
		if valUpdate.Power < 0 {
			return fmt.Errorf("voting power can't be negative %v", valUpdate)
		} else if valUpdate.Power == 0 {
			// continue, since this is deleting the validator, and thus there is no
			// pubkey to check
			continue
		}

		// Check if validator's pubkey matches an ABCI type in the consensus params
		pubkeyTypeURL := amino.GetTypeURL(valUpdate.PubKey)
		if !params.IsValidPubKeyTypeURL(pubkeyTypeURL) {
			return fmt.Errorf("validator %v is using pubkey %s, which is unsupported for consensus",
				valUpdate, pubkeyTypeURL)
		}
	}
	return nil
}

// updateState returns a new State updated according to the header and responses.
func updateState(
	state State,
	blockID types.BlockID,
	header *types.Header,
	abciResponses *ABCIResponses,
) (State, error) {
	// Copy the valset so we can apply changes from EndBlock
	// and update s.LastValidators and s.Validators.
	nValSet := state.NextValidators.Copy()

	// Update the validator set with the latest abciResponses.
	lastHeightValsChanged := state.LastHeightValidatorsChanged
	if u := abciResponses.EndBlock.ValidatorUpdates; len(u) > 0 {
		err := nValSet.UpdateWithABCIValidatorUpdates(u)
		if err != nil {
			return state, fmt.Errorf("Error changing validator set: %w", err)
		}
		// Change results from this height but only applies to the next next height.
		lastHeightValsChanged = header.Height + 1 + 1
	}

	// Update validator proposer priority and set state variables.
	nValSet.IncrementProposerPriority(1)

	// Update the params with the latest abciResponses.
	nextParams := state.ConsensusParams
	lastHeightParamsChanged := state.LastHeightConsensusParamsChanged
	if abciResponses.EndBlock.ConsensusParams != nil {
		// NOTE: must not mutate s.ConsensusParams
		nextParams = state.ConsensusParams.Update(*abciResponses.EndBlock.ConsensusParams)
		err := types.ValidateConsensusParams(nextParams)
		if err != nil {
			return state, fmt.Errorf("Error updating consensus params: %w", err)
		}
		// Change results from this height but only applies to the next height.
		lastHeightParamsChanged = header.Height + 1
	}

	// NOTE: the AppHash has not been populated.
	// It will be filled on state.Save.
	return State{
		SoftwareVersion:                  tmver.Version,
		BlockVersion:                     typesver.BlockVersion,
		AppVersion:                       state.AppVersion, // TODO
		ChainID:                          state.ChainID,
		LastBlockHeight:                  header.Height,
		LastBlockTotalTx:                 state.LastBlockTotalTx + header.NumTxs,
		LastBlockID:                      blockID,
		LastBlockTime:                    header.Time,
		NextValidators:                   nValSet,
		Validators:                       state.NextValidators.Copy(),
		LastValidators:                   state.Validators.Copy(),
		LastHeightValidatorsChanged:      lastHeightValsChanged,
		ConsensusParams:                  nextParams,
		LastHeightConsensusParamsChanged: lastHeightParamsChanged,
		LastResultsHash:                  abciResponses.ResultsHash(),
		AppHash:                          nil,
	}, nil
}

// Fire NewBlock, NewBlockHeader.
// Fire TxEvent for every tx.
// NOTE: if Tendermint crashes before commit, some or all of these events may be published again.
func fireEvents(evsw events.EventSwitch, block *types.Block, abciResponses *ABCIResponses) {
	evsw.FireEvent(types.EventNewBlock{
		Block:            block,
		ResultBeginBlock: abciResponses.BeginBlock,
		ResultEndBlock:   abciResponses.EndBlock,
	})
	evsw.FireEvent(types.EventNewBlockHeader{
		Header:           block.Header,
		ResultBeginBlock: abciResponses.BeginBlock,
		ResultEndBlock:   abciResponses.EndBlock,
	})

	for i, tx := range block.Data.Txs {
		evsw.FireEvent(types.EventTx{Result: types.TxResult{
			Height:   block.Height,
			Index:    uint32(i),
			Tx:       tx,
			Response: (abciResponses.DeliverTxs[i]),
		}})
	}

	if u := abciResponses.EndBlock.ValidatorUpdates; len(u) > 0 {
		evsw.FireEvent(
			types.EventValidatorSetUpdates{ValidatorUpdates: u})
	}
}

// ----------------------------------------------------------------------------------------------------
// Execute block without state. TODO: eliminate

// ExecCommitBlock executes and commits a block on the proxyApp without validating or mutating the state.
// It returns the application root hash (result of abci.Commit).
func ExecCommitBlock(
	appConnConsensus appconn.Consensus,
	block *types.Block,
	logger *slog.Logger,
	stateDB dbm.DB,
) ([]byte, error) {
	_, err := execBlockOnProxyApp(logger, appConnConsensus, block, stateDB)
	if err != nil {
		logger.Error("Error executing block on proxy app", "height", block.Height, "err", err)
		return nil, err
	}
	// Commit block, get hash back
	res, err := appConnConsensus.CommitSync()
	if err != nil {
		logger.Error("Client error during proxyAppConn.CommitSync", "err", res)
		return nil, err
	}
	// ResponseCommit has no error or log, just data
	return res.Data, nil
}
