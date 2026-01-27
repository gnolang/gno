package consensus

import (
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mock"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cnscfg "github.com/gnolang/gno/tm2/pkg/bft/consensus/config"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

func TestHandler_ValidatorsHandler(t *testing.T) {
	t.Parallel()

	t.Run("Invalid height param", func(t *testing.T) {
		t.Parallel()

		var (
			db            = memdb.NewMemDB()
			mockConsensus = &mock.Consensus{
				GetStateFn: func() sm.State {
					return sm.State{
						LastBlockHeight: 10,
					}
				},
			}
			params = []any{"not-an-int"}
		)

		h := NewHandler(mockConsensus, db, &mock.Peers{})

		res, err := h.ValidatorsHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Height below minimum", func(t *testing.T) {
		t.Parallel()

		var (
			db            = memdb.NewMemDB()
			mockConsensus = &mock.Consensus{
				GetStateFn: func() sm.State {
					return sm.State{
						LastBlockHeight: 10,
					}
				},
			}
			params = []any{int64(-1)}
		)

		h := NewHandler(mockConsensus, db, &mock.Peers{})

		res, err := h.ValidatorsHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Validators not found", func(t *testing.T) {
		t.Parallel()

		var (
			db            = memdb.NewMemDB()
			mockConsensus = &mock.Consensus{
				GetStateFn: func() sm.State {
					return sm.State{
						LastBlockHeight: 0,
					}
				},
			}
		)

		h := NewHandler(mockConsensus, db, &mock.Peers{})

		res, err := h.ValidatorsHandler(nil, nil)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Valid default height", func(t *testing.T) {
		t.Parallel()

		var (
			db = memdb.NewMemDB()

			valSet          = &types.ValidatorSet{}
			consensusParams = abci.ConsensusParams{}

			st = sm.State{
				LastBlockHeight:                  0,
				Validators:                       valSet,
				NextValidators:                   valSet,
				LastHeightValidatorsChanged:      1,
				ConsensusParams:                  consensusParams,
				LastHeightConsensusParamsChanged: 1,
			}

			mockConsensus = &mock.Consensus{
				GetStateFn: func() sm.State {
					return st
				},
			}
		)

		// Seed the state
		sm.SaveState(db, st)

		h := NewHandler(mockConsensus, db, &mock.Peers{})

		res, err := h.ValidatorsHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultValidators)
		require.True(t, ok)

		assert.Equal(t, int64(1), result.BlockHeight)
		assert.Equal(t, valSet.Validators, result.Validators)
	})
}

func TestHandler_DumpConsensusStateHandler(t *testing.T) {
	t.Parallel()

	t.Run("Unexpected params", func(t *testing.T) {
		t.Parallel()

		h := NewHandler(nil, nil, nil)

		res, err := h.DumpConsensusStateHandler(nil, []any{"extra"})
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Valid dump", func(t *testing.T) {
		t.Parallel()

		var (
			cfg = &cnscfg.ConsensusConfig{}
			rs  = &cstypes.RoundState{}

			mockConsensus = &mock.Consensus{
				GetConfigDeepCopyFn: func() *cnscfg.ConsensusConfig {
					return cfg
				},
				GetRoundStateDeepCopyFn: func() *cstypes.RoundState {
					return rs
				},
			}

			mockPeers = &mock.Peers{
				PeersFn: func() p2p.PeerSet {
					return &mock.PeerSet{}
				},
			}
		)

		h := NewHandler(mockConsensus, nil, mockPeers)

		res, err := h.DumpConsensusStateHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultDumpConsensusState)
		require.True(t, ok)

		assert.Same(t, cfg, result.Config)
		assert.Same(t, rs, result.RoundState)
		assert.Len(t, result.Peers, 0)
	})
}

func TestHandler_ConsensusStateHandler(t *testing.T) {
	t.Parallel()

	t.Run("Unexpected params", func(t *testing.T) {
		t.Parallel()

		h := NewHandler(
			&mock.Consensus{},
			memdb.NewMemDB(),
			&mock.Peers{},
		)

		res, err := h.ConsensusStateHandler(nil, []any{"extra"})
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Valid simple round state", func(t *testing.T) {
		t.Parallel()

		var (
			simple = cstypes.RoundStateSimple{
				HeightRoundStep: "10/0/0",
			}

			mockConsensus = &mock.Consensus{
				GetRoundStateSimpleFn: func() cstypes.RoundStateSimple {
					return simple
				},
			}
		)

		h := NewHandler(mockConsensus, memdb.NewMemDB(), &mock.Peers{})

		res, err := h.ConsensusStateHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultConsensusState)
		require.True(t, ok)

		assert.Equal(t, simple, result.RoundState)
	})
}

func TestHandler_ConsensusParamsHandler(t *testing.T) {
	t.Parallel()

	t.Run("Invalid height param", func(t *testing.T) {
		t.Parallel()

		var (
			mockConsensus = &mock.Consensus{
				GetStateFn: func() sm.State {
					return sm.State{
						LastBlockHeight: 10,
					}
				},
			}
			db     = memdb.NewMemDB()
			params = []any{"not-an-int"}
		)

		h := NewHandler(mockConsensus, db, &mock.Peers{})

		res, err := h.ConsensusParamsHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Height below minimum", func(t *testing.T) {
		t.Parallel()

		var (
			mockConsensus = &mock.Consensus{
				GetStateFn: func() sm.State {
					return sm.State{
						LastBlockHeight: 10,
					}
				},
			}

			db     = memdb.NewMemDB()
			params = []any{int64(-1)}
		)

		h := NewHandler(mockConsensus, db, &mock.Peers{})

		res, err := h.ConsensusParamsHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Consensus params not found", func(t *testing.T) {
		t.Parallel()

		var (
			mockConsensus = &mock.Consensus{
				GetStateFn: func() sm.State {
					return sm.State{
						LastBlockHeight: 0,
					}
				},
			}

			db = memdb.NewMemDB()
		)

		h := NewHandler(mockConsensus, db, &mock.Peers{})

		res, err := h.ConsensusParamsHandler(nil, nil)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
	})

	t.Run("Valid latest height", func(t *testing.T) {
		t.Parallel()

		var (
			db              = memdb.NewMemDB()
			consensusParams = abci.ConsensusParams{}

			st = sm.State{
				LastBlockHeight:                  0,
				Validators:                       &types.ValidatorSet{},
				NextValidators:                   &types.ValidatorSet{},
				LastHeightValidatorsChanged:      1,
				ConsensusParams:                  consensusParams,
				LastHeightConsensusParamsChanged: 1,
			}

			mockConsensus = &mock.Consensus{
				GetStateFn: func() sm.State {
					return st
				},
			}
		)

		sm.SaveState(db, st)

		h := NewHandler(mockConsensus, db, &mock.Peers{})

		res, err := h.ConsensusParamsHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultConsensusParams)
		require.True(t, ok)

		assert.Equal(t, int64(1), result.BlockHeight)
		assert.Equal(t, consensusParams, result.ConsensusParams)
	})
}
