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

// setupStatusGlobals sets the global consensus/transport/pubkey state
// needed by the Status handler's success path.
func setupStatusGlobals() {
	SetGetFastSync(func() bool { return true })
	SetConsensusState(&mockConsensus{
		getValidatorsFn: func() (int64, []*types.Validator) {
			return 0, nil
		},
	})
	SetP2PTransport(&mockTransport{
		nodeInfoFn: func() p2pTypes.NodeInfo {
			return p2pTypes.NodeInfo{}
		},
	})
	SetPubKey(ed25519.GenPrivKey().PubKey())
}

func TestStatusHandler(t *testing.T) {
	t.Run("nil block meta", func(t *testing.T) {
		var height int64 = 10

		SetBlockStore(&mockBlockStore{
			heightFn:        func() int64 { return height },
			loadBlockMetaFn: func(int64) *types.BlockMeta { return nil },
		})
		SetGetFastSync(func() bool { return true })

		result, err := Status(&rpctypes.Context{}, nil)
		require.Nil(t, result)
		assert.ErrorContains(t, err, "block meta not found for height 10")
	})

	t.Run("success with block meta", func(t *testing.T) {
		var height int64 = 10

		SetBlockStore(&mockBlockStore{
			heightFn: func() int64 { return height },
			loadBlockMetaFn: func(int64) *types.BlockMeta {
				return &types.BlockMeta{Header: types.Header{Height: height}}
			},
		})
		setupStatusGlobals()

		result, err := Status(&rpctypes.Context{}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, height, result.SyncInfo.LatestBlockHeight)
	})

	t.Run("height zero skips block meta load", func(t *testing.T) {
		SetBlockStore(&mockBlockStore{
			heightFn: func() int64 { return 0 },
			loadBlockMetaFn: func(int64) *types.BlockMeta {
				t.Fatal("LoadBlockMeta should not be called when height is 0")
				return nil
			},
		})
		setupStatusGlobals()

		result, err := Status(&rpctypes.Context{}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(0), result.SyncInfo.LatestBlockHeight)
	})
}
