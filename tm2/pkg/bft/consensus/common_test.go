package consensus

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/counter"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	mempl "github.com/gnolang/gno/tm2/pkg/bft/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/privval"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/store"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

const (
	testSubscriber = "test-client"
)

// A cleanupFunc cleans up any config / test files created for a particular
// test.
type cleanupFunc func()

// genesis, chain_id, priv_val
var config *cfg.Config // NOTE: must be reset for each _test.go file
var (
	consensusReplayConfig *cfg.Config
	ensureTimeout         = time.Millisecond * 20000
)

func ensureDir(dir string, mode os.FileMode) {
	if err := osm.EnsureDir(dir, mode); err != nil {
		panic(err)
	}
}

func ResetConfig(name string) (*cfg.Config, string) {
	return cfg.ResetTestRoot(name)
}

// -------------------------------------------------------------------------------
// validator stub (a kvstore consensus peer we control)

type validatorStub struct {
	Index  int // Validator index. NOTE: we don't assume validator set changes.
	Height int64
	Round  int
	types.PrivValidator
}

var testMinPower int64 = 10

func NewValidatorStub(privValidator types.PrivValidator, valIndex int) *validatorStub {
	return &validatorStub{
		Index:         valIndex,
		PrivValidator: privValidator,
	}
}

func (vs *validatorStub) signVote(voteType types.SignedMsgType, hash []byte, header types.PartSetHeader) (*types.Vote, error) {
	addr := vs.PrivValidator.GetPubKey().Address()
	vote := &types.Vote{
		ValidatorIndex:   vs.Index,
		ValidatorAddress: addr,
		Height:           vs.Height,
		Round:            vs.Round,
		Timestamp:        tmtime.Now(),
		Type:             voteType,
		BlockID:          types.BlockID{Hash: hash, PartsHeader: header},
	}
	err := vs.PrivValidator.SignVote(config.ChainID(), vote)
	return vote, err
}

// Sign vote for type/hash/header
func signVote(vs *validatorStub, voteType types.SignedMsgType, hash []byte, header types.PartSetHeader) *types.Vote {
	v, err := vs.signVote(voteType, hash, header)
	if err != nil {
		panic(fmt.Errorf("failed to sign vote: %w", err))
	}
	return v
}

func signVotes(voteType types.SignedMsgType, hash []byte, header types.PartSetHeader, vss ...*validatorStub) []*types.Vote {
	votes := make([]*types.Vote, len(vss))
	for i, vs := range vss {
		votes[i] = signVote(vs, voteType, hash, header)
	}
	return votes
}

func incrementHeight(vss ...*validatorStub) {
	for _, vs := range vss {
		vs.Height++
	}
}

func incrementRound(vss ...*validatorStub) {
	for _, vs := range vss {
		vs.Round++
	}
}

type ValidatorStubsByAddress []*validatorStub

func (vss ValidatorStubsByAddress) Len() int {
	return len(vss)
}

func (vss ValidatorStubsByAddress) Less(i, j int) bool {
	return vss[i].GetPubKey().Address().Compare(vss[j].GetPubKey().Address()) == -1
}

func (vss ValidatorStubsByAddress) Swap(i, j int) {
	it := vss[i]
	vss[i] = vss[j]
	vss[i].Index = i
	vss[j] = it
	vss[j].Index = j
}

// -------------------------------------------------------------------------------
// Functions for transitioning the consensus state

func startFrom(cs *ConsensusState, height int64, round int) {
	go func() {
		cs.enterNewRound(height, round)
		cs.StartWithoutWALCatchup()
	}()
}

// Create proposal block from cs but sign it with vs.
// NOTE: assumes cs already locked via mutex (perhaps via debugger).
func decideProposal(cs *ConsensusState, vs *validatorStub, height int64, round int) (proposal *types.Proposal, block *types.Block) {
	block, blockParts := cs.createProposalBlock()
	validRound := cs.ValidRound
	chainID := cs.state.ChainID
	if block == nil {
		panic("Failed to createProposalBlock. Did you forget to add commit for previous block?")
	}

	// Make proposal
	polRound, propBlockID := validRound, types.BlockID{Hash: block.Hash(), PartsHeader: blockParts.Header()}
	proposal = types.NewProposal(height, round, polRound, propBlockID)
	if err := vs.SignProposal(chainID, proposal); err != nil {
		panic(err)
	}
	return
}

func addVotes(to *ConsensusState, votes ...*types.Vote) {
	for _, vote := range votes {
		to.peerMsgQueue <- msgInfo{Msg: &VoteMessage{vote}}
	}
}

func signAddVotes(to *ConsensusState, voteType types.SignedMsgType, hash []byte, header types.PartSetHeader, vss ...*validatorStub) {
	votes := signVotes(voteType, hash, header, vss...)
	addVotes(to, votes...)
}

func validatePrevote(cs *ConsensusState, round int, privVal *validatorStub, blockHash []byte) {
	prevotes := cs.Votes.Prevotes(round)
	address := privVal.GetPubKey().Address()
	var vote *types.Vote
	if vote = prevotes.GetByAddress(address); vote == nil {
		panic("Failed to find prevote from validator")
	}
	if blockHash == nil {
		if vote.BlockID.Hash != nil {
			panic(fmt.Sprintf("Expected prevote to be for nil, got %X", vote.BlockID.Hash))
		}
	} else {
		if !bytes.Equal(vote.BlockID.Hash, blockHash) {
			panic(fmt.Sprintf("Expected prevote to be for %X, got %X", blockHash, vote.BlockID.Hash))
		}
	}
}

func validateLastPrecommit(cs *ConsensusState, privVal *validatorStub, blockHash []byte) {
	votes := cs.LastCommit
	address := privVal.GetPubKey().Address()
	var vote *types.Vote
	if vote = votes.GetByAddress(address); vote == nil {
		panic("Failed to find precommit from validator")
	}
	if !bytes.Equal(vote.BlockID.Hash, blockHash) {
		panic(fmt.Sprintf("Expected precommit to be for %X, got %X", blockHash, vote.BlockID.Hash))
	}
}

func validatePrecommit(_ *testing.T, cs *ConsensusState, thisRound, lockRound int, privVal *validatorStub, votedBlockHash, lockedBlockHash []byte) {
	precommits := cs.Votes.Precommits(thisRound)
	address := privVal.GetPubKey().Address()
	var vote *types.Vote
	if vote = precommits.GetByAddress(address); vote == nil {
		panic("Failed to find precommit from validator")
	}

	if votedBlockHash == nil {
		if vote.BlockID.Hash != nil {
			panic("Expected precommit to be for nil")
		}
	} else {
		if !bytes.Equal(vote.BlockID.Hash, votedBlockHash) {
			panic("Expected precommit to be for proposal block")
		}
	}

	if lockedBlockHash == nil {
		if cs.LockedRound != lockRound || cs.LockedBlock != nil {
			panic(fmt.Sprintf("Expected to be locked on nil at round %d. Got locked at round %d with block %v", lockRound, cs.LockedRound, cs.LockedBlock))
		}
	} else {
		if cs.LockedRound != lockRound || !bytes.Equal(cs.LockedBlock.Hash(), lockedBlockHash) {
			panic(fmt.Sprintf("Expected block to be locked on round %d, got %d. Got locked block %X, expected %X", lockRound, cs.LockedRound, cs.LockedBlock.Hash(), lockedBlockHash))
		}
	}
}

func validatePrevoteAndPrecommit(t *testing.T, cs *ConsensusState, thisRound, lockRound int, privVal *validatorStub, votedBlockHash, lockedBlockHash []byte) {
	t.Helper()

	// verify the prevote
	validatePrevote(cs, thisRound, privVal, votedBlockHash)
	// verify precommit
	validatePrecommit(t, cs, thisRound, lockRound, privVal, votedBlockHash, lockedBlockHash)
}

func subscribeToVoter(cs *ConsensusState, addr crypto.Address) <-chan events.Event {
	return events.SubscribeFiltered(cs.evsw, testSubscriber, func(event events.Event) bool {
		if vote, ok := event.(types.EventVote); ok {
			if vote.Vote.ValidatorAddress == addr {
				return true
			}
		}
		return false
	})
}

// -------------------------------------------------------------------------------
// consensus states

func newConsensusState(state sm.State, pv types.PrivValidator, app abci.Application) *ConsensusState {
	config, _ := cfg.ResetTestRoot("consensus_state_test")
	return newConsensusStateWithConfig(config, state, pv, app)
}

func newConsensusStateWithConfig(thisConfig *cfg.Config, state sm.State, pv types.PrivValidator, app abci.Application) *ConsensusState {
	blockDB := memdb.NewMemDB()
	return newConsensusStateWithConfigAndBlockStore(thisConfig, state, pv, app, blockDB)
}

func newConsensusStateWithConfigAndBlockStore(thisConfig *cfg.Config, state sm.State, pv types.PrivValidator, app abci.Application, blockDB dbm.DB) *ConsensusState {
	// Get BlockStore
	blockStore := store.NewBlockStore(blockDB)

	// one for mempool, one for consensus
	mtx := new(sync.Mutex)
	proxyAppConnMem := abcicli.NewLocalClient(mtx, app)
	proxyAppConnCon := abcicli.NewLocalClient(mtx, app)

	// Make Mempool
	mempool := mempl.NewCListMempool(thisConfig.Mempool, proxyAppConnMem, 0, state.ConsensusParams.Block.MaxTxBytes)
	mempool.SetLogger(log.NewNoopLogger().With("module", "mempool"))
	if thisConfig.Consensus.WaitForTxs() {
		mempool.EnableTxsAvailable()
	}

	// Make ConsensusState
	stateDB := blockDB
	sm.SaveState(stateDB, state) // for save height 1's validators info
	blockExec := sm.NewBlockExecutor(stateDB, log.NewNoopLogger(), proxyAppConnCon, mempool)
	cs := NewConsensusState(thisConfig.Consensus, state, blockExec, blockStore, mempool)
	cs.SetLogger(log.NewNoopLogger().With("module", "consensus"))
	cs.SetPrivValidator(pv)

	evsw := events.NewEventSwitch()
	evsw.SetLogger(log.NewNoopLogger().With("module", "events"))
	evsw.Start()
	cs.SetEventSwitch(evsw)
	return cs
}

func loadPrivValidator(config *cfg.Config) *privval.FilePV {
	privValidatorKeyFile := config.PrivValidatorKeyFile()
	ensureDir(filepath.Dir(privValidatorKeyFile), 0o700)
	privValidatorStateFile := config.PrivValidatorStateFile()
	privValidator := privval.LoadOrGenFilePV(privValidatorKeyFile, privValidatorStateFile)
	privValidator.Reset()
	return privValidator
}

func randConsensusState(nValidators int) (*ConsensusState, []*validatorStub) {
	// Get State
	state, privVals := randGenesisState(nValidators, false, 10)

	vss := make([]*validatorStub, nValidators)

	cs := newConsensusState(state, privVals[0], counter.NewCounterApplication(true))

	for i := 0; i < nValidators; i++ {
		vss[i] = NewValidatorStub(privVals[i], i)
	}
	// since cs1 starts at 1
	incrementHeight(vss[1:]...)

	return cs, vss
}

// -------------------------------------------------------------------------------

func ensureNoNewEvent(ch <-chan events.Event, timeout time.Duration,
	errorMessage string,
) {
	select {
	case <-time.After(timeout):
		break
	case <-ch:
		panic(errorMessage)
	}
}

func ensureNoNewEventOnChannel(ch <-chan events.Event) {
	ensureNoNewEvent(
		ch,
		ensureTimeout,
		"We should be stuck waiting, not receiving new event on the channel")
}

func ensureNoNewRoundStep(stepCh <-chan events.Event) {
	ensureNoNewEvent(
		stepCh,
		ensureTimeout,
		"We should be stuck waiting, not receiving NewRoundStep event")
}

func ensureNoNewUnlock(unlockCh <-chan events.Event) {
	ensureNoNewEvent(
		unlockCh,
		ensureTimeout,
		"We should be stuck waiting, not receiving Unlock event")
}

func ensureNoNewTimeout(stepCh <-chan events.Event, timeout int64) {
	timeoutDuration := time.Duration(timeout*10) * time.Nanosecond
	ensureNoNewEvent(
		stepCh,
		timeoutDuration,
		"We should be stuck waiting, not receiving NewTimeout event")
}

func ensureNewEvent(ch <-chan events.Event, height int64, round int, timeout time.Duration, errorMessage string) {
	select {
	case <-time.After(timeout):
		osm.PrintAllGoroutines()
		panic(errorMessage)
	case msg := <-ch:
		csevent, ok := msg.(cstypes.ConsensusEvent)
		if !ok {
			panic(fmt.Sprintf("expected a ConsensusEvent, got %T. Wrong subscription channel?",
				msg))
		}
		if csevent.GetHRS().Height != height {
			panic(fmt.Sprintf("expected height %v, got %v", height, csevent.GetHRS().Height))
		}
		if csevent.GetHRS().Round != round {
			panic(fmt.Sprintf("expected round %v, got %v", round, csevent.GetHRS().Round))
		}
		// TODO: We could check also for a step at this point!
	}
}

func ensureNewRound(roundCh <-chan events.Event, height int64, round int) {
	select {
	case <-time.After(ensureTimeout):
		panic("Timeout expired while waiting for NewRound event")
	case msg := <-roundCh:
		newRoundEvent, ok := msg.(cstypes.EventNewRound)
		if !ok {
			panic(fmt.Sprintf("expected a EventNewRound, got %T. Wrong subscription channel?",
				msg))
		}
		if newRoundEvent.Height != height {
			panic(fmt.Sprintf("expected height %v, got %v", height, newRoundEvent.Height))
		}
		if newRoundEvent.Round != round {
			panic(fmt.Sprintf("expected round %v, got %v", round, newRoundEvent.Round))
		}
	}
}

func ensureNewRoundStep(stepCh <-chan events.Event, height int64, round int, step cstypes.RoundStepType) {
	select {
	case <-time.After(ensureTimeout):
		panic("Timeout expired while waiting for NewRoundStep event")
	case msg := <-stepCh:
		newStepEvent, ok := msg.(cstypes.EventNewRoundStep)
		if !ok {
			panic(fmt.Sprintf("expected a EventNewRound, got %T. Wrong subscription channel?",
				msg))
		}
		if newStepEvent.Height != height {
			panic(fmt.Sprintf("expected height %v, got %v", height, newStepEvent.Height))
		}
		if newStepEvent.Round != round {
			panic(fmt.Sprintf("expected round %v, got %v", round, newStepEvent.Round))
		}
		if newStepEvent.Step != step {
			panic(fmt.Sprintf("expected step %v, got %v", step, newStepEvent.Step))
		}
	}
}

func ensureNewTimeout(timeoutCh <-chan events.Event, height int64, round int, timeout int64) {
	timeoutDuration := (time.Duration(timeout))*time.Nanosecond + ensureTimeout
	ensureNewEvent(timeoutCh, height, round, timeoutDuration,
		"Timeout expired while waiting for NewTimeout event")
}

func ensureNewProposal(proposalCh <-chan events.Event, height int64, round int) {
	select {
	case <-time.After(ensureTimeout):
		panic("Timeout expired while waiting for NewProposal event")
	case msg := <-proposalCh:
		proposalEvent, ok := msg.(cstypes.EventCompleteProposal)
		if !ok {
			panic(fmt.Sprintf("expected a EventCompleteProposal, got %T. Wrong subscription channel?",
				msg))
		}
		if proposalEvent.Height != height {
			panic(fmt.Sprintf("expected height %v, got %v", height, proposalEvent.Height))
		}
		if proposalEvent.Round != round {
			panic(fmt.Sprintf("expected round %v, got %v", round, proposalEvent.Round))
		}
	}
}

func ensureNewValidBlock(validBlockCh <-chan events.Event, height int64, round int) {
	ensureNewEvent(validBlockCh, height, round, ensureTimeout,
		"Timeout expired while waiting for NewValidBlock event")
}

func ensureNewBlock(blockCh <-chan events.Event, height int64) {
	select {
	case <-time.After(ensureTimeout):
		panic("Timeout expired while waiting for NewBlock event")
	case msg := <-blockCh:
		blockEvent, ok := msg.(types.EventNewBlock)
		if !ok {
			panic(fmt.Sprintf("expected a EventNewBlock, got %T. Wrong subscription channel?",
				msg))
		}
		if blockEvent.Block.Height != height {
			panic(fmt.Sprintf("expected height %v, got %v", height, blockEvent.Block.Height))
		}
	}
}

func ensureNewBlockHeader(blockCh <-chan events.Event, height int64, blockHash []byte) {
	select {
	case <-time.After(ensureTimeout):
		panic("Timeout expired while waiting for NewBlockHeader event")
	case msg := <-blockCh:
		blockHeaderEvent, ok := msg.(types.EventNewBlockHeader)
		if !ok {
			panic(fmt.Sprintf("expected a EventNewBlockHeader, got %T. Wrong subscription channel?",
				msg))
		}
		if blockHeaderEvent.Header.Height != height {
			panic(fmt.Sprintf("expected height %v, got %v", height, blockHeaderEvent.Header.Height))
		}
		if !bytes.Equal(blockHeaderEvent.Header.Hash(), blockHash) {
			panic(fmt.Sprintf("expected header %X, got %X", blockHash, blockHeaderEvent.Header.Hash()))
		}
	}
}

func ensureNewUnlock(unlockCh <-chan events.Event, height int64, round int) {
	ensureNewEvent(unlockCh, height, round, ensureTimeout,
		"Timeout expired while waiting for NewUnlock event")
}

func ensureProposal(proposalCh <-chan events.Event, height int64, round int, propID types.BlockID) {
	select {
	case <-time.After(ensureTimeout):
		panic("Timeout expired while waiting for NewProposal event")
	case msg := <-proposalCh:
		proposalEvent, ok := msg.(cstypes.EventCompleteProposal)
		if !ok {
			panic(fmt.Sprintf("expected a EventCompleteProposal, got %T. Wrong subscription channel?",
				msg))
		}
		if proposalEvent.Height != height {
			panic(fmt.Sprintf("expected height %v, got %v", height, proposalEvent.Height))
		}
		if proposalEvent.Round != round {
			panic(fmt.Sprintf("expected round %v, got %v", round, proposalEvent.Round))
		}
		if !proposalEvent.BlockID.Equals(propID) {
			panic("Proposed block does not match expected block")
		}
	}
}

func ensurePrecommit(voteCh <-chan events.Event, height int64, round int) {
	ensureVote(voteCh, height, round, types.PrecommitType)
}

func ensurePrevote(voteCh <-chan events.Event, height int64, round int) {
	ensureVote(voteCh, height, round, types.PrevoteType)
}

func ensureVote(voteCh <-chan events.Event, height int64, round int,
	voteType types.SignedMsgType,
) {
	select {
	case <-time.After(ensureTimeout):
		panic("Timeout expired while waiting for NewVote event")
	case msg := <-voteCh:
		voteEvent, ok := msg.(types.EventVote)
		if !ok {
			panic(fmt.Sprintf("expected a EventVote, got %T. Wrong subscription channel?",
				msg))
		}
		vote := voteEvent.Vote
		if vote.Height != height {
			panic(fmt.Sprintf("expected height %v, got %v", height, vote.Height))
		}
		if vote.Round != round {
			panic(fmt.Sprintf("expected round %v, got %v", round, vote.Round))
		}
		if vote.Type != voteType {
			panic(fmt.Sprintf("expected type %v, got %v", voteType, vote.Type))
		}
	}
}

func ensureNewEventOnChannel(ch <-chan events.Event) {
	select {
	case <-time.After(ensureTimeout):
		panic("Timeout expired while waiting for new activity on the channel")
	case <-ch:
	}
}

// -------------------------------------------------------------------------------
// consensus nets

func randConsensusNet(nValidators int, testName string, tickerFunc func() TimeoutTicker,
	appFunc func() abci.Application, configOpts ...func(*cfg.Config),
) ([]*ConsensusState, cleanupFunc) {
	genDoc, privVals := randGenesisDoc(nValidators, false, 30)
	css := make([]*ConsensusState, nValidators)
	apps := make([]abci.Application, nValidators)
	logger := log.NewNoopLogger()
	configRootDirs := make([]string, 0, nValidators)
	for i := 0; i < nValidators; i++ {
		stateDB := memdb.NewMemDB() // each state needs its own db
		state, _ := sm.LoadStateFromDBOrGenesisDoc(stateDB, genDoc)
		thisConfig, _ := ResetConfig(fmt.Sprintf("%s_%d", testName, i))
		configRootDirs = append(configRootDirs, thisConfig.RootDir)
		for _, opt := range configOpts {
			opt(thisConfig)
		}
		ensureDir(filepath.Dir(thisConfig.Consensus.WalFile()), 0o700) // dir for wal
		app := appFunc()
		vals := state.Validators.ABCIValidatorUpdates()
		app.InitChain(abci.RequestInitChain{Validators: vals})

		css[i] = newConsensusStateWithConfigAndBlockStore(thisConfig, state, privVals[i], app, stateDB)
		css[i].SetTimeoutTicker(tickerFunc())
		css[i].SetLogger(logger.With("validator", i, "module", "consensus"))
		apps[i] = app
	}
	return css, func() {
		for _, dir := range configRootDirs {
			os.RemoveAll(dir)
		}
		for _, cs := range css {
			cs.Stop()
			cs.Wait()
		}
		for _, app := range apps {
			app.Close()
		}
	}
}

// nPeers = nValidators + nNotValidator
func randConsensusNetWithPeers(nValidators, nPeers int, testName string, tickerFunc func() TimeoutTicker, appFunc func(string) abci.Application) ([]*ConsensusState, *types.GenesisDoc, *cfg.Config, cleanupFunc) {
	genDoc, privVals := randGenesisDoc(nValidators, false, testMinPower)
	css := make([]*ConsensusState, nPeers)
	apps := make([]abci.Application, nPeers)
	logger := log.NewNoopLogger()
	var peer0Config *cfg.Config
	configRootDirs := make([]string, 0, nPeers)
	for i := 0; i < nPeers; i++ {
		stateDB := memdb.NewMemDB() // each state needs its own db
		state, _ := sm.LoadStateFromDBOrGenesisDoc(stateDB, genDoc)
		thisConfig, _ := ResetConfig(fmt.Sprintf("%s_%d", testName, i))
		configRootDirs = append(configRootDirs, thisConfig.RootDir)
		ensureDir(filepath.Dir(thisConfig.Consensus.WalFile()), 0o700) // dir for wal
		if i == 0 {
			peer0Config = thisConfig
		}
		var privVal types.PrivValidator
		if i < nValidators {
			privVal = privVals[i]
		} else {
			tempKeyFile, err := os.CreateTemp("", "priv_validator_key_")
			if err != nil {
				panic(err)
			}
			tempStateFile, err := os.CreateTemp("", "priv_validator_state_")
			if err != nil {
				panic(err)
			}

			privVal = privval.GenFilePV(tempKeyFile.Name(), tempStateFile.Name())
		}

		app := appFunc(path.Join(config.DBDir(), fmt.Sprintf("%s_%d", testName, i)))
		vals := state.Validators.ABCIValidatorUpdates()
		if _, ok := app.(*kvstore.PersistentKVStoreApplication); ok {
			state.AppVersion = kvstore.AppVersion
			// simulate handshake, receive app version. If don't do this, replay test will fail
		}
		app.InitChain(abci.RequestInitChain{Validators: vals})
		// sm.SaveState(stateDB,state)	//height 1's validatorsInfo already saved in LoadStateFromDBOrGenesisDoc above

		css[i] = newConsensusStateWithConfig(thisConfig, state, privVal, app)
		css[i].SetTimeoutTicker(tickerFunc())
		css[i].SetLogger(logger.With("validator", i, "module", "consensus"))
		apps[i] = app
	}
	return css, genDoc, peer0Config, func() {
		for _, dir := range configRootDirs {
			os.RemoveAll(dir)
		}
		for _, cs := range css {
			cs.Stop()
			cs.Wait()
		}
		for _, app := range apps {
			app.Close()
		}
	}
}

// -------------------------------------------------------------------------------
// genesis

func randGenesisDoc(numValidators int, randPower bool, minPower int64) (*types.GenesisDoc, []types.PrivValidator) {
	validators := make([]types.GenesisValidator, numValidators)
	privValidators := make([]types.PrivValidator, numValidators)
	for i := 0; i < numValidators; i++ {
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

func randGenesisState(numValidators int, randPower bool, minPower int64) (sm.State, []types.PrivValidator) {
	genDoc, privValidators := randGenesisDoc(numValidators, randPower, minPower)
	s0, _ := sm.MakeGenesisState(genDoc)
	return s0, privValidators
}

// ------------------------------------
// mock ticker

func newMockTickerFunc(onlyOnce bool) func() TimeoutTicker {
	return func() TimeoutTicker {
		return &mockTicker{
			c:        make(chan timeoutInfo, 10),
			onlyOnce: onlyOnce,
		}
	}
}

// mock ticker only fires on RoundStepNewHeight
// and only once if onlyOnce=true
type mockTicker struct {
	c chan timeoutInfo

	mtx      sync.Mutex
	onlyOnce bool
	fired    bool
}

func (m *mockTicker) Start() error {
	return nil
}

func (m *mockTicker) Stop() error {
	return nil
}

func (m *mockTicker) ScheduleTimeout(ti timeoutInfo) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if m.onlyOnce && m.fired {
		return
	}
	if ti.Step == cstypes.RoundStepNewHeight {
		m.c <- ti
		m.fired = true
	}
}

func (m *mockTicker) Chan() <-chan timeoutInfo {
	return m.c
}

func (*mockTicker) SetLogger(_ *slog.Logger) {}

// ------------------------------------

func newCounter() abci.Application {
	return counter.NewCounterApplication(true)
}

func newPersistentKVStore() abci.Application {
	dir, err := os.MkdirTemp("", "persistent-kvstore")
	if err != nil {
		panic(err)
	}
	return kvstore.NewPersistentKVStoreApplication(dir)
}

func newPersistentKVStoreWithPath(dbDir string) abci.Application {
	return kvstore.NewPersistentKVStoreApplication(dbDir)
}
