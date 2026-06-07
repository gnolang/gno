package core

import (
	"testing"

	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// applyStatusEnvDefaults populates the consensus/transport/pubkey fields
// needed by the Status handler's success path on an Environment that
// already has a BlockStore set.
func applyStatusEnvDefaults(env *Environment) {
	env.GetFastSync = func() bool { return true }
	env.Consensus = &mockConsensus{
		getValidatorsFn: func() (int64, []*types.Validator) {
			return 0, nil
		},
	}
	env.P2PTransport = &mockTransport{
		nodeInfoFn: func() p2pTypes.NodeInfo {
			return p2pTypes.NodeInfo{}
		},
	}
	env.PubKey = ed25519.GenPrivKey().PubKey()
}

func TestStatusHandler(t *testing.T) {
	t.Run("nil block meta", func(t *testing.T) {
		var height int64 = 10

		env := &Environment{
			BlockStore: &mockBlockStore{
				heightFn:        func() int64 { return height },
				loadBlockMetaFn: func(int64) *types.BlockMeta { return nil },
			},
			GetFastSync: func() bool { return true },
		}

		result, err := env.Status(&rpctypes.Context{}, nil)
		require.Nil(t, result)
		assert.ErrorContains(t, err, "block meta not found for height 10")
	})

	t.Run("success with block meta", func(t *testing.T) {
		var height int64 = 10

		env := &Environment{
			BlockStore: &mockBlockStore{
				heightFn: func() int64 { return height },
				loadBlockMetaFn: func(int64) *types.BlockMeta {
					return &types.BlockMeta{Header: types.Header{Height: height}}
				},
			},
		}
		applyStatusEnvDefaults(env)

		result, err := env.Status(&rpctypes.Context{}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, height, result.SyncInfo.LatestBlockHeight)
	})

	t.Run("height zero skips block meta load", func(t *testing.T) {
		env := &Environment{
			BlockStore: &mockBlockStore{
				heightFn: func() int64 { return 0 },
				loadBlockMetaFn: func(int64) *types.BlockMeta {
					t.Fatal("LoadBlockMeta should not be called when height is 0")
					return nil
				},
			},
		}
		applyStatusEnvDefaults(env)

		result, err := env.Status(&rpctypes.Context{}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(0), result.SyncInfo.LatestBlockHeight)
	})
}
