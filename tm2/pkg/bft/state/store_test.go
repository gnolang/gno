package state_test

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

func TestStoreLoadValidators(t *testing.T) {
	t.Parallel()

	stateDB := memdb.NewMemDB()
	val, _ := types.RandValidator(true, 10)
	vals := types.NewValidatorSet([]*types.Validator{val})

	// 1) LoadValidators loads validators using a height where they were last changed
	sm.SaveValidatorsInfo(stateDB, 1, 1, vals)
	sm.SaveValidatorsInfo(stateDB, 2, 1, vals)
	loadedVals, err := sm.LoadValidators(stateDB, 2)
	require.NoError(t, err)
	assert.NotZero(t, loadedVals.Size())

	// 2) LoadValidators loads validators using a checkpoint height

	// TODO(melekes): REMOVE in 0.33 release
	// https://github.com/tendermint/tendermint/issues/3543
	// for releases prior to v0.31.4, it uses last height changed
	valInfo := &sm.ValidatorsInfo{
		LastHeightChanged: sm.ValSetCheckpointInterval,
	}
	stateDB.Set(sm.CalcValidatorsKey(sm.ValSetCheckpointInterval), valInfo.Bytes())
	assert.NotPanics(t, func() {
		sm.SaveValidatorsInfo(stateDB, sm.ValSetCheckpointInterval+1, 1, vals)
		loadedVals, err := sm.LoadValidators(stateDB, sm.ValSetCheckpointInterval+1)
		if err != nil {
			t.Fatal(err)
		}
		if loadedVals.Size() == 0 {
			t.Fatal("Expected validators to be non-empty")
		}
	})
	// ENDREMOVE

	sm.SaveValidatorsInfo(stateDB, sm.ValSetCheckpointInterval, 1, vals)

	loadedVals, err = sm.LoadValidators(stateDB, sm.ValSetCheckpointInterval)
	require.NoError(t, err)
	assert.NotZero(t, loadedVals.Size())
}

// TestValidatorsCacheBoundsReplay checks that ValidatorsCache returns what
// LoadValidators returns, replaying only the rounds above the nearest kept set.
func TestValidatorsCacheBoundsReplay(t *testing.T) {
	t.Parallel()

	const checkpoint = sm.ValSetCheckpointInterval
	stateDB := memdb.NewMemDB()
	vals := make([]*types.Validator, 6)
	for i := range vals {
		vals[i], _ = types.RandValidator(false, int64(10*(i+1)))
	}
	set := types.NewValidatorSet(vals)
	// Priorities spread wider than the window and off centre, so the replay
	// has to rescale them first, as LoadValidators does.
	for i, p := range []int64{9_000, -4_000, 350, 0, -77, 1_234} {
		set.Validators[i].ProposerPriority = p
	}
	sm.SaveValidatorsInfo(stateDB, checkpoint, 1, set)
	sm.SaveValidatorsInfo(stateDB, 2*checkpoint, 1, set)
	// A second segment starts where the set changes rather than at a checkpoint.
	changed := int64(2*checkpoint + 500)
	set2 := set.Copy()
	set2.Validators[0].ProposerPriority = -9_000
	sm.SaveValidatorsInfo(stateDB, changed, changed, set2)

	var cache sm.ValidatorsCache
	for _, tc := range []struct {
		height int64
		rounds int64
	}{
		{checkpoint + checkpoint - 1, checkpoint - 1}, // first load replays the segment
		{checkpoint + checkpoint - 2, 998},
		{checkpoint + 50_001, 1},
		{checkpoint + 1, 1},
		{checkpoint + sm.ValidatorsCacheInterval, sm.ValidatorsCacheInterval},
		{checkpoint + sm.ValidatorsCacheInterval + 1, 1},
		{checkpoint + 2*sm.ValidatorsCacheInterval, sm.ValidatorsCacheInterval},
		{checkpoint + checkpoint - 1, 999},
		{checkpoint, 0},
		{changed + 3_001, 3_001},
		{changed + 2_500, 500},
	} {
		lastChanged := int64(1)
		if tc.height > changed {
			lastChanged = changed
		}
		sm.SaveValidatorsInfo(stateDB, tc.height, lastChanged, set)

		want, err := sm.LoadValidators(stateDB, tc.height)
		require.NoError(t, err)
		got, rounds, err := cache.LoadValidatorsCounting(stateDB, tc.height)
		require.NoError(t, err)
		assert.Equal(t, want.Validators, got.Validators, "height %d", tc.height)
		assert.Equal(t, want.Proposer, got.Proposer, "height %d", tc.height)
		assert.Equal(t, tc.rounds, rounds, "height %d", tc.height)
	}
}

// TestValidatorsCacheTwoDBs checks that one ValidatorsCache loading the same
// height from two DBs returns each DB's own set.
func TestValidatorsCacheTwoDBs(t *testing.T) {
	t.Parallel()

	const checkpoint = sm.ValSetCheckpointInterval
	height := int64(checkpoint + 1_500)
	var cache sm.ValidatorsCache
	for range 2 {
		stateDB := memdb.NewMemDB()
		val, _ := types.RandValidator(false, 10)
		set := types.NewValidatorSet([]*types.Validator{val})
		sm.SaveValidatorsInfo(stateDB, checkpoint, 1, set)
		sm.SaveValidatorsInfo(stateDB, height, 1, set)

		want, err := sm.LoadValidators(stateDB, height)
		require.NoError(t, err)
		got, err := cache.LoadValidators(stateDB, height)
		require.NoError(t, err)
		assert.Equal(t, want.Validators, got.Validators)
	}
}

// TestValidatorsCacheConcurrentLoads checks that loads running at once through
// one ValidatorsCache return what LoadValidators returns. Run it with -race.
func TestValidatorsCacheConcurrentLoads(t *testing.T) {
	t.Parallel()

	const checkpoint = sm.ValSetCheckpointInterval
	stateDB := memdb.NewMemDB()
	vals := make([]*types.Validator, 6)
	for i := range vals {
		vals[i], _ = types.RandValidator(false, int64(10*(i+1)))
	}
	set := types.NewValidatorSet(vals)
	sm.SaveValidatorsInfo(stateDB, checkpoint, 1, set)

	heights := make([]int64, 8)
	wants := make([]*types.ValidatorSet, len(heights))
	for i := range heights {
		heights[i] = checkpoint + int64(i+1)*sm.ValidatorsCacheInterval + 500
		sm.SaveValidatorsInfo(stateDB, heights[i], 1, set)
		var err error
		wants[i], err = sm.LoadValidators(stateDB, heights[i])
		require.NoError(t, err)
	}

	var cache sm.ValidatorsCache
	var wg sync.WaitGroup
	for i, height := range heights {
		wg.Go(func() {
			got, err := cache.LoadValidators(stateDB, height)
			if assert.NoError(t, err) {
				assert.Equal(t, wants[i].Validators, got.Validators, "height %d", height)
				assert.Equal(t, wants[i].Proposer, got.Proposer, "height %d", height)
			}
		})
	}
	wg.Wait()
}

// TestValidatorsCacheMixedPowers checks that a kept set resumes without being
// rescaled again, which voting powers far apart would show.
func TestValidatorsCacheMixedPowers(t *testing.T) {
	t.Parallel()

	const checkpoint = sm.ValSetCheckpointInterval
	stateDB := memdb.NewMemDB()
	powers := []int64{100, 2, 100, 1_000_000_000, 1, 100, 1 << 40, 7, 1_000_000_000, 1}
	vals := make([]*types.Validator, len(powers))
	for i, power := range powers {
		pk := ed25519.GenPrivKeyFromSecret([]byte{byte(i)})
		vals[i] = types.NewValidator(pk.PubKey(), power)
	}
	set := types.NewValidatorSet(vals)
	// Priorities far outside the window: with these powers, rescaling a kept
	// set again on resume changes the priorities returned.
	for i, p := range []int64{-383e12, 269e12, -284e12, 66e12, 262e12, 187e12, 138e12, 14e12, -169e12, 280e12} {
		set.Validators[i].ProposerPriority = p
	}
	sm.SaveValidatorsInfo(stateDB, checkpoint, 1, set)

	var cache sm.ValidatorsCache
	for _, height := range []int64{
		checkpoint + checkpoint - 1,
		checkpoint + 3*sm.ValidatorsCacheInterval + 1,
		checkpoint + 10*sm.ValidatorsCacheInterval + 999,
	} {
		sm.SaveValidatorsInfo(stateDB, height, 1, set)
		want, err := sm.LoadValidators(stateDB, height)
		require.NoError(t, err)
		got, err := cache.LoadValidators(stateDB, height)
		require.NoError(t, err)
		assert.Equal(t, want.Validators, got.Validators, "height %d", height)
		assert.Equal(t, want.Proposer, got.Proposer, "height %d", height)
	}
}

// TestLoadValidatorsAtInitialHeight verifies that when a genesis has
// InitialHeight > 1, validators are saved and loadable at that height.
// This test demonstrates a bug: saveState only saves validators when
// nextHeight == 1, so InitialHeight > 1 causes LoadValidators to fail.
func TestLoadValidatorsAtInitialHeight(t *testing.T) {
	t.Parallel()

	val, _ := types.RandValidator(true, 10)
	vals := []*types.Validator{val}
	genDoc := &types.GenesisDoc{
		ChainID:       "test-chain",
		InitialHeight: 50,
		Validators: []types.GenesisValidator{
			{Address: val.Address, PubKey: val.PubKey, Power: 10},
		},
	}
	require.NoError(t, genDoc.ValidateAndComplete())

	state, err := sm.MakeGenesisState(genDoc)
	require.NoError(t, err)

	// Simulate what the Handshaker does after InitChain with InitialHeight > 1:
	// it sets LastBlockHeight = InitialHeight - 1.
	state.LastBlockHeight = 49 // InitialHeight - 1

	stateDB := memdb.NewMemDB()
	sm.SaveState(stateDB, state)

	// Should be able to load validators at InitialHeight (50).
	// BUG: This fails with NoValSetForHeightError because saveState only
	// saves validators at nextHeight when nextHeight == 1.
	loadedVals, err := sm.LoadValidators(stateDB, 50)
	require.NoError(t, err, "should load validators at InitialHeight")
	require.NotNil(t, loadedVals)
	assert.Equal(t, len(vals), loadedVals.Size())
}

// TestLoadConsensusParamsAtInitialHeight verifies that consensus params are
// saved and loadable at InitialHeight when InitialHeight > 1.
// This test demonstrates the same class of bug as TestLoadValidatorsAtInitialHeight
// but for consensus params.
func TestLoadConsensusParamsAtInitialHeight(t *testing.T) {
	t.Parallel()

	val, _ := types.RandValidator(true, 10)
	genDoc := &types.GenesisDoc{
		ChainID:       "test-chain",
		InitialHeight: 50,
		Validators: []types.GenesisValidator{
			{Address: val.Address, PubKey: val.PubKey, Power: 10},
		},
	}
	require.NoError(t, genDoc.ValidateAndComplete())

	state, err := sm.MakeGenesisState(genDoc)
	require.NoError(t, err)

	// Simulate Handshaker setting LastBlockHeight = InitialHeight - 1.
	state.LastBlockHeight = 49

	stateDB := memdb.NewMemDB()
	sm.SaveState(stateDB, state)

	// Should be able to load consensus params at InitialHeight (50).
	// BUG: saveConsensusParamsInfo stores at nextHeight=50 with
	// changeHeight=1 (LastHeightConsensusParamsChanged defaults to 1 from
	// MakeGenesisState). Since changeHeight(1) != nextHeight(50), only a
	// reference is saved (not the full params). When loading at height 50,
	// it finds empty params, looks up LastHeightChanged=1, which doesn't
	// exist => panic.
	require.NotPanics(t, func() {
		params, err := sm.LoadConsensusParams(stateDB, 50)
		require.NoError(t, err, "should load consensus params at InitialHeight")
		assert.NotEmpty(t, params.Block)
	}, "LoadConsensusParams should not panic at InitialHeight")
}

// TestSaveABCIResponsesIsDurable pins that the responses reach the disk when
// they are written, not when the state that supersedes them is written.
func TestSaveABCIResponsesIsDurable(t *testing.T) {
	t.Parallel()

	responses := &sm.ABCIResponses{
		DeliverTxs: []abci.ResponseDeliverTx{
			{ResponseBase: abci.ResponseBase{Data: []byte("delivered")}},
		},
	}
	require.NotEmpty(t, responses.Bytes(), "payload must be non-empty for the assertion below to mean anything")

	stateDB := newUnsyncedWriteDB()
	require.NoError(t, sm.SaveABCIResponses(stateDB, 1, responses))

	loaded, err := sm.LoadABCIResponses(stateDB.Durable(), 1)
	require.NoError(t, err)
	assert.Equal(t, responses.DeliverTxs, loaded.DeliverTxs)
}

// TestSaveABCIResponsesReturnsWriteError pins that a failed write surfaces to
// the caller instead of being dropped: the caller must get the chance to stop
// the block before the application commits a height with no recovery record.
func TestSaveABCIResponsesReturnsWriteError(t *testing.T) {
	t.Parallel()

	writeErr := errors.New("injected responses write failure")
	stateDB := &failingWriteDB{
		DB:      memdb.NewMemDB(),
		failKey: sm.CalcABCIResponsesKey(1),
		failErr: writeErr,
	}

	err := sm.SaveABCIResponses(stateDB, 1, &sm.ABCIResponses{
		DeliverTxs: []abci.ResponseDeliverTx{
			{ResponseBase: abci.ResponseBase{Data: []byte("delivered")}},
		},
	})
	require.ErrorIs(t, err, writeErr)
	require.ErrorContains(t, err, "height 1")
}

func BenchmarkLoadValidators(b *testing.B) {
	const valSetSize = 100

	config, genesisFile := cfg.ResetTestRoot("state_")
	defer os.RemoveAll(config.RootDir)
	dbType := dbm.BackendType(config.DBBackend)
	stateDB, err := dbm.NewDB("state", dbType, config.DBDir())
	require.NoError(b, err)

	state, err := sm.LoadStateFromDBOrGenesisFile(stateDB, genesisFile)
	require.NoError(b, err)
	state.Validators = genValSet(valSetSize)
	state.NextValidators = state.Validators.CopyIncrementProposerPriority(1)
	sm.SaveState(stateDB, state)

	for i := 10; i < 10000000000; i *= 10 { // 10, 100, 1000, ...
		i := i
		sm.SaveValidatorsInfo(stateDB, int64(i), state.LastHeightValidatorsChanged, state.NextValidators)

		b.Run(fmt.Sprintf("height=%d", i), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_, err := sm.LoadValidators(stateDB, int64(i))
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
