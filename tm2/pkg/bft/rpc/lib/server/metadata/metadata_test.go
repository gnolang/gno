package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadata_NewMetadata(t *testing.T) {
	t.Parallel()

	t.Run("HTTP metadata", func(t *testing.T) {
		t.Parallel()

		address := "remote address"
		m := NewMetadata(address)

		require.NotNil(t, m)

		assert.Equal(t, address, m.RemoteAddr)
		assert.False(t, m.IsWS())
	})

	t.Run("WS metadata", func(t *testing.T) {
		t.Parallel()

		address := "remote address"
		wsID := "ws ID"
		m := NewMetadata(address, WithWebSocketID(wsID))

		require.NotNil(t, m)

		assert.Equal(t, address, m.RemoteAddr)
		assert.True(t, m.IsWS())
		assert.Equal(t, wsID, *m.WebSocketID)
	})
}
