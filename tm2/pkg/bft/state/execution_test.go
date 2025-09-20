package state_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/async"
	"github.com/gnolang/gno/tm2/pkg/bft/abci/example/kvstore"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/bft/mempool/mock"
	"github.com/gnolang/gno/tm2/pkg/bft/proxy"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
)

var (
	chainID      = "execution_chain"
	testPartSize = 65536
	nTxsPerBlock = 10
)

func TestApplyBlock(t *testing.T) {
	t.Parallel()

	cc := proxy.NewLocalClientCreator(kvstore.NewKVStoreApplication())
	proxyApp := appconn.NewAppConns(cc)
	err := proxyApp.Start()
	require.Nil(t, err)
	defer proxyApp.Stop()

	state, stateDB, _ := makeState(1, 1)

	blockExec := sm.NewBlockExecutor(stateDB, log.NewTestingLogger(t), proxyApp.Consensus(), mock.Mempool{})
	evsw := events.NewEventSwitch()
	blockExec.SetEventSwitch(evsw)

	block := makeBlock(state, 1)
	blockID := types.BlockID{Hash: block.Hash(), PartsHeader: block.MakePartSet(testPartSize).Header()}

	state, err = blockExec.ApplyBlock(state, blockID, block)
	require.Nil(t, err)

	// TODO check state and mempool
	_ = state
}

// TestBeginBlockValidators ensures we send absent validators list.
func TestBeginBlockValidators(t *testing.T) {
	t.Parallel()

	app := &testApp{}
	cc := proxy.NewLocalClientCreator(app)
	proxyApp := appconn.NewAppConns(cc)
	err := proxyApp.Start()
	require.Nil(t, err)
	defer proxyApp.Stop()

	state, stateDB, _ := makeState(2, 2)

	prevHash := state.LastBlockID.Hash
	prevParts := types.PartSetHeader{}
	prevBlockID := types.BlockID{Hash: prevHash, PartsHeader: prevParts}

	now := tmtime.Now()
	commitSig0 := (&types.Vote{ValidatorIndex: 0, Timestamp: now, Type: types.PrecommitType}).CommitSig()
	commitSig1 := (&types.Vote{ValidatorIndex: 1, Timestamp: now}).CommitSig()

	testCases := []struct {
		desc                     string
		lastCommitPrecommits     []*types.CommitSig
		expectedAbsentValidators []int
	}{
		{"none absent", []*types.CommitSig{commitSig0, commitSig1}, []int{}},
		{"one absent", []*types.CommitSig{commitSig0, nil}, []int{1}},
		{"multiple absent", []*types.CommitSig{nil, nil}, []int{0, 1}},
	}

	for _, tc := range testCases {
		lastCommit := types.NewCommit(prevBlockID, tc.lastCommitPrecommits)

		// block for height 2
		block, _ := state.MakeBlock(2, makeTxs(2), lastCommit, state.Validators.GetProposer().Address)

		_, err = sm.ExecCommitBlock(proxyApp.Consensus(), block, log.NewTestingLogger(t), stateDB)
		require.Nil(t, err, tc.desc)

		// -> app receives a list of validators with a bool indicating if they signed
		ctr := 0
		for i, v := range app.CommitVotes {
			if ctr < len(tc.expectedAbsentValidators) &&
				tc.expectedAbsentValidators[ctr] == i {
				assert.False(t, v.SignedLastBlock)
				ctr++
			} else {
				assert.True(t, v.SignedLastBlock)
			}
		}
	}
}

func TestValidateValidatorUpdates(t *testing.T) {
	t.Parallel()

	pubkey1 := ed25519.GenPrivKey().PubKey()
	pubkey2 := ed25519.GenPrivKey().PubKey()

	secpKey := secp256k1.GenPrivKey().PubKey()

	defaultValidatorParams := abci.ValidatorParams{PubKeyTypeURLs: []string{amino.GetTypeURL(ed25519.PubKeyEd25519{})}}

	testCases := []struct {
		name string

		abciUpdates     []abci.ValidatorUpdate
		validatorParams abci.ValidatorParams

		shouldErr bool
	}{
		{
			"adding a validator is OK",

			[]abci.ValidatorUpdate{{PubKey: (pubkey2), Power: 20}},
			defaultValidatorParams,

			false,
		},
		{
			"updating a validator is OK",

			[]abci.ValidatorUpdate{{PubKey: (pubkey1), Power: 20}},
			defaultValidatorParams,

			false,
		},
		{
			"removing a validator is OK",

			[]abci.ValidatorUpdate{{PubKey: (pubkey2), Power: 0}},
			defaultValidatorParams,

			false,
		},
		{
			"adding a validator with negative power results in error",

			[]abci.ValidatorUpdate{{PubKey: (pubkey2), Power: -100}},
			defaultValidatorParams,

			true,
		},
		{
			"adding a validator with pubkey thats not in validator params results in error",

			[]abci.ValidatorUpdate{{PubKey: (secpKey), Power: -100}},
			defaultValidatorParams,

			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := sm.ValidateValidatorUpdates(tc.abciUpdates, tc.validatorParams)
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateValidators(t *testing.T) {
	t.Parallel()

	pubkey1 := ed25519.GenPrivKey().PubKey()
	val1 := types.NewValidator(pubkey1, 10)
	pubkey2 := ed25519.GenPrivKey().PubKey()
	val2 := types.NewValidator(pubkey2, 20)

	testCases := []struct {
		name string

		currentSet  *types.ValidatorSet
		abciUpdates []abci.ValidatorUpdate

		resultingSet *types.ValidatorSet
		shouldErr    bool
	}{
		{
			"adding a validator is OK",

			types.NewValidatorSet([]*types.Validator{val1}),
			[]abci.ValidatorUpdate{{PubKey: (pubkey2), Power: 20}},

			types.NewValidatorSet([]*types.Validator{val1, val2}),
			false,
		},
		{
			"updating a validator is OK",

			types.NewValidatorSet([]*types.Validator{val1}),
			[]abci.ValidatorUpdate{{PubKey: (pubkey1), Power: 20}},

			types.NewValidatorSet([]*types.Validator{types.NewValidator(pubkey1, 20)}),
			false,
		},
		{
			"removing a validator is OK",

			types.NewValidatorSet([]*types.Validator{val1, val2}),
			[]abci.ValidatorUpdate{{PubKey: (pubkey2), Power: 0}},

			types.NewValidatorSet([]*types.Validator{val1}),
			false,
		},
		{
			"removing a non-existing validator results in error",

			types.NewValidatorSet([]*types.Validator{val1}),
			[]abci.ValidatorUpdate{{PubKey: (pubkey2), Power: 0}},

			types.NewValidatorSet([]*types.Validator{val1}),
			true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.currentSet.UpdateWithABCIValidatorUpdates(tc.abciUpdates)
			if tc.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.Equal(t, tc.resultingSet.Size(), tc.currentSet.Size())

				assert.Equal(t, tc.resultingSet.TotalVotingPower(), tc.currentSet.TotalVotingPower())

				assert.Equal(t, tc.resultingSet.Validators[0].Address, tc.currentSet.Validators[0].Address)
				if tc.resultingSet.Size() > 1 {
					assert.Equal(t, tc.resultingSet.Validators[1].Address, tc.currentSet.Validators[1].Address)
				}
			}
		})
	}
}

// TestEndBlockValidatorUpdates ensures we update validator set and send an event.
func TestEndBlockValidatorUpdates(t *testing.T) {
	t.Parallel()

	app := &testApp{}
	cc := proxy.NewLocalClientCreator(app)
	proxyApp := appconn.NewAppConns(cc)
	err := proxyApp.Start()
	require.Nil(t, err)
	defer proxyApp.Stop()

	state, stateDB, _ := makeState(1, 1)

	blockExec := sm.NewBlockExecutor(stateDB, log.NewTestingLogger(t), proxyApp.Consensus(), mock.Mempool{})

	evsw := events.NewEventSwitch()
	err = evsw.Start()
	require.NoError(t, err)
	defer evsw.Stop()
	blockExec.SetEventSwitch(evsw)

	updatesSub := events.Subscribe(evsw, "TestEndBlockValidatorUpdates")
	require.NoError(t, err)

	block := makeBlock(state, 1)
	blockID := types.BlockID{Hash: block.Hash(), PartsHeader: block.MakePartSet(testPartSize).Header()}

	pubkey := ed25519.GenPrivKey().PubKey()
	app.ValidatorUpdates = []abci.ValidatorUpdate{
		{PubKey: (pubkey), Power: 10},
	}

	// Run in goroutine.
	done := async.Routine(func() {
		state, err := blockExec.ApplyBlock(state, blockID, block)
		require.Nil(t, err)

		// test new validator was added to NextValidators
		if assert.Equal(t, state.Validators.Size()+1, state.NextValidators.Size()) {
			idx, _ := state.NextValidators.GetByAddress(pubkey.Address())
			if idx < 0 {
				t.Fatalf("can't find address %v in the set %v", pubkey.Address(), state.NextValidators)
			}
		}
	})

	// test we threw an event
LOOP:
	for {
		select {
		case msg := <-updatesSub:
			switch event := msg.(type) {
			case types.EventValidatorSetUpdates:
				if assert.NotEmpty(t, event.ValidatorUpdates) {
					assert.Equal(t, pubkey, event.ValidatorUpdates[0].PubKey)
					assert.EqualValues(t, 10, event.ValidatorUpdates[0].Power)
					break LOOP
				}
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Did not receive EventValidatorSetUpdates within 1 sec.")
		}
	}

	<-done
}

// TestEndBlockValidatorUpdatesResultingInEmptySet checks that processing validator updates that
// would result in empty set causes no panic, an error is raised and NextValidators is not updated
func TestEndBlockValidatorUpdatesResultingInEmptySet(t *testing.T) {
	t.Parallel()

	app := &testApp{}
	cc := proxy.NewLocalClientCreator(app)
	proxyApp := appconn.NewAppConns(cc)
	err := proxyApp.Start()
	require.Nil(t, err)
	defer proxyApp.Stop()

	state, stateDB, _ := makeState(1, 1)
	blockExec := sm.NewBlockExecutor(stateDB, log.NewTestingLogger(t), proxyApp.Consensus(), mock.Mempool{})

	block := makeBlock(state, 1)
	blockID := types.BlockID{Hash: block.Hash(), PartsHeader: block.MakePartSet(testPartSize).Header()}

	// Remove the only validator
	app.ValidatorUpdates = []abci.ValidatorUpdate{
		{PubKey: (state.Validators.Validators[0].PubKey), Power: 0},
	}

	assert.NotPanics(t, func() { state, err = blockExec.ApplyBlock(state, blockID, block) })
	assert.NotNil(t, err)
	assert.NotEmpty(t, state.NextValidators.Validators)
}
