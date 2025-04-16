package discovery

import (
	"net"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateNetAddrs generates random net addresses
func generateNetAddrs(t *testing.T, count int) []*types.NetAddress {
	t.Helper()

	addrs := make([]*types.NetAddress, count)

	for i := range count {
		var (
			key     = types.GenerateNodeKey()
			address = "127.0.0.1:8080"
		)

		tcpAddr, err := net.ResolveTCPAddr("tcp", address)
		require.NoError(t, err)

		addr, err := types.NewNetAddress(key.ID(), tcpAddr)
		require.NoError(t, err)

		addrs[i] = addr
	}

	return addrs
}

func TestRequest_ValidateBasic(t *testing.T) {
	t.Parallel()

	r := &Request{}

	assert.NoError(t, r.ValidateBasic())
}

func TestResponse_ValidateBasic(t *testing.T) {
	t.Parallel()

	t.Run("empty peer set", func(t *testing.T) {
		t.Parallel()

		r := &Response{
			Peers: make([]*types.NetAddress, 0),
		}

		assert.ErrorIs(t, r.ValidateBasic(), errNoPeers)
	})

	t.Run("invalid peer dial address", func(t *testing.T) {
		t.Parallel()

		r := &Response{
			Peers: []*types.NetAddress{
				{
					ID: "", // invalid ID
				},
			},
		}

		assert.Error(t, r.ValidateBasic())
	})

	t.Run("valid peer set", func(t *testing.T) {
		t.Parallel()

		r := &Response{
			Peers: generateNetAddrs(t, 10),
		}

		assert.NoError(t, r.ValidateBasic())
	})
}
