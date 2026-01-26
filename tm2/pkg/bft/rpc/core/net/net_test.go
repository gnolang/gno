package net

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mock"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_NetInfo(t *testing.T) {
	t.Parallel()

	t.Run("Unexpected params", func(t *testing.T) {
		t.Parallel()

		h := &Handler{}

		res, err := h.NetInfoHandler(nil, []any{"extra"})
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Valid, empty peer set", func(t *testing.T) {
		t.Parallel()

		var (
			mockPeerSet = &mock.PeerSet{}

			mockPeers = &mock.Peers{
				PeersFn: func() p2p.PeerSet {
					return mockPeerSet
				},
			}

			expectedListeners = []string{"tcp://0.0.0.0:26656"}

			mockTransport = &mock.Transport{
				ListenersFn: func() []string {
					return expectedListeners
				},
				IsListeningFn: func() bool {
					return true
				},
			}
		)

		h := &Handler{
			peers:     mockPeers,
			transport: mockTransport,
		}

		res, err := h.NetInfoHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultNetInfo)
		require.True(t, ok)

		assert.True(t, result.Listening)
		assert.Equal(t, expectedListeners, result.Listeners)
		assert.Equal(t, 0, result.NPeers)
		assert.Len(t, result.Peers, 0)
	})
}

func TestHandler_GenesisHandler(t *testing.T) {
	t.Parallel()

	t.Run("Unexpected params", func(t *testing.T) {
		t.Parallel()

		h := &Handler{}

		res, err := h.GenesisHandler(nil, []any{"extra"})
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Returns genesis doc", func(t *testing.T) {
		t.Parallel()

		genDoc := &types.GenesisDoc{
			ChainID: "test-chain",
		}

		h := &Handler{
			genesisDoc: genDoc,
		}

		res, err := h.GenesisHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultGenesis)
		require.True(t, ok)

		assert.Equal(t, genDoc, result.Genesis)
	})
}
