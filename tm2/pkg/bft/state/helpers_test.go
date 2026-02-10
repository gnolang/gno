package state_test

import (
	"bytes"
	"fmt"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

type paramsChangeTestCase struct {
	height int64
	params abci.ConsensusParams
}

func newTestApp() appconn.AppConns {
	app := &testApp{}
	cc := proxy.NewLocalClientCreator(app)
	return appconn.NewAppConns(cc)
}

func makeAndCommitGoodBlock(
	state sm.State,
	height int64,
	lastCommit *types.Commit,
	proposerAddr crypto.Address,
	blockExec *sm.BlockExecutor,
	privVals map[string]types.PrivValidator,
) (sm.State, types.BlockID, *types.Commit, error) {
	// A good block passes
	state, blockID, err := makeAndApplyGoodBlock(state, height, lastCommit, proposerAddr, blockExec)
	if err != nil {
		return state, types.BlockID{}, nil, err
	}

	// Simulate a lastCommit for this block from all validators for the next height
	commit, err := makeValidCommit(height, blockID, state.Validators, privVals)
	if err != nil {
		return state, types.BlockID{}, nil, err
	}
	return state, blockID, commit, nil
}

func makeAndApplyGoodBlock(state sm.State, height int64, lastCommit *types.Commit, proposerAddr crypto.Address,
	blockExec *sm.BlockExecutor,
) (sm.State, types.BlockID, error) {
	block, _ := state.MakeBlock(height, makeTxs(height), lastCommit, proposerAddr)
	if err := state.ValidateBlock(block); err != nil {
		return state, types.BlockID{}, err
	}
	blockID := types.BlockID{Hash: block.Hash(), PartsHeader: types.PartSetHeader{}}
	state, err := blockExec.ApplyBlock(state, blockID, block)
	if err != nil {
		return state, types.BlockID{}, err
	}
	return state, blockID, nil
}

func makeValidCommit(height int64, blockID types.BlockID, vals *types.ValidatorSet, privVals map[string]types.PrivValidator) (*types.Commit, error) {
	sigs := make([]*types.CommitSig, 0)
	for i := range vals.Size() {
		_, val := vals.GetByIndex(i)
		vote, err := types.MakeVote(height, blockID, vals, privVals[val.Address.String()], chainID)
		if err != nil {
			return nil, err
		}
		sigs = append(sigs, vote.CommitSig())
	}
	return types.NewCommit(blockID, sigs), nil
}

// make some bogus txs
func makeTxs(height int64) (txs []types.Tx) {
	for i := range nTxsPerBlock {
		txs = append(txs, types.Tx([]byte{byte(height), byte(i)}))
	}
	return txs
}

func makeState(nVals, height int) (sm.State, dbm.DB, map[string]types.PrivValidator) {
	vals := make([]types.GenesisValidator, nVals)
	privVals := make(map[string]types.PrivValidator, nVals)
	for i := range nVals {
		secret := fmt.Appendf(nil, "test%d", i)
		pk := ed25519.GenPrivKeyFromSecret(secret)
		valAddr := pk.PubKey().Address()
		vals[i] = types.GenesisValidator{
			Address: valAddr,
			PubKey:  pk.PubKey(),
			Power:   1000,
			Name:    fmt.Sprintf("test%d", i),
		}
		privVals[valAddr.String()] = types.NewMockPVWithPrivKey(pk)
	}
	s, _ := sm.MakeGenesisState(&types.GenesisDoc{
		ChainID:    chainID,
		Validators: vals,
		AppHash:    nil,
	})

	stateDB := memdb.NewMemDB()
	sm.SaveState(stateDB, s)

	for i := 1; i < height; i++ {
		s.LastBlockHeight++
		s.LastValidators = s.Validators.Copy()
		sm.SaveState(stateDB, s)
	}
	return s, stateDB, privVals
}

func makeBlock(state sm.State, height int64) *types.Block {
	block, _ := state.MakeBlock(height, makeTxs(state.LastBlockHeight), new(types.Commit), state.Validators.GetProposer().Address)
	return block
}

func genValSet(size int) *types.ValidatorSet {
	vals := make([]*types.Validator, size)
	for i := range size {
		vals[i] = types.NewValidator(ed25519.GenPrivKey().PubKey(), 10)
	}
	return types.NewValidatorSet(vals)
}

func makeConsensusParams( // XXX search and replace
	maxTxBytes, maxDataBytes, maxBlockBytes, maxGas int64,
	timeIotaMS int64,
) abci.ConsensusParams {
	return abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxTxBytes:    maxTxBytes,
			MaxDataBytes:  maxDataBytes,
			MaxBlockBytes: maxBlockBytes,
			MaxGas:        maxGas,
			TimeIotaMS:    timeIotaMS,
		},
	}
}

func makeHeaderPartsResponsesValPubKeyChange(state sm.State, pubkey crypto.PubKey) (types.Header, types.BlockID, *sm.ABCIResponses) {
	block := makeBlock(state, state.LastBlockHeight+1)
	abciResponses := &sm.ABCIResponses{
		EndBlock: abci.ResponseEndBlock{ValidatorUpdates: nil},
	}

	// If the pubkey is new, remove the old and add the new.
	_, val := state.NextValidators.GetByIndex(0)
	if !bytes.Equal(pubkey.Bytes(), val.PubKey.Bytes()) {
		abciResponses.EndBlock = abci.ResponseEndBlock{
			ValidatorUpdates: []abci.ValidatorUpdate{
				{Address: val.Address, PubKey: val.PubKey, Power: 0},
				{Address: pubkey.Address(), PubKey: pubkey, Power: 10},
			},
		}
	}

	return block.Header, types.BlockID{Hash: block.Hash(), PartsHeader: types.PartSetHeader{}}, abciResponses
}

func makeHeaderPartsResponsesValPowerChange(state sm.State, power int64) (types.Header, types.BlockID, *sm.ABCIResponses) {
	block := makeBlock(state, state.LastBlockHeight+1)
	abciResponses := &sm.ABCIResponses{
		EndBlock: abci.ResponseEndBlock{ValidatorUpdates: nil},
	}

	// If the pubkey is new, remove the old and add the new.
	_, val := state.NextValidators.GetByIndex(0)
	if val.VotingPower != power {
		abciResponses.EndBlock = abci.ResponseEndBlock{
			ValidatorUpdates: []abci.ValidatorUpdate{
				{Address: val.Address, PubKey: val.PubKey, Power: power},
			},
		}
	}

	return block.Header, types.BlockID{Hash: block.Hash(), PartsHeader: types.PartSetHeader{}}, abciResponses
}

func makeHeaderPartsResponsesParams(state sm.State, params abci.ConsensusParams) (types.Header, types.BlockID, *sm.ABCIResponses) {
	block := makeBlock(state, state.LastBlockHeight+1)
	abciResponses := &sm.ABCIResponses{
		EndBlock: abci.ResponseEndBlock{ConsensusParams: &params},
	}
	return block.Header, types.BlockID{Hash: block.Hash(), PartsHeader: types.PartSetHeader{}}, abciResponses
}

// ----------------------------------------------------------------------------

type testApp struct {
	abci.BaseApplication

	CommitVotes      []abci.VoteInfo
	ValidatorUpdates []abci.ValidatorUpdate
}

var _ abci.Application = (*testApp)(nil)

func (app *testApp) Info(req abci.RequestInfo) (resInfo abci.ResponseInfo) {
	return abci.ResponseInfo{}
}

func (app *testApp) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	app.CommitVotes = req.LastCommitInfo.Votes
	return abci.ResponseBeginBlock{}
}

func (app *testApp) EndBlock(req abci.RequestEndBlock) abci.ResponseEndBlock {
	return abci.ResponseEndBlock{ValidatorUpdates: app.ValidatorUpdates}
}

func (app *testApp) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	return abci.ResponseDeliverTx{}
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
