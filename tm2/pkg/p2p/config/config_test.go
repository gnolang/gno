package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestP2PConfig_ValidateBasic(t *testing.T) {
	t.Parallel()

	t.Run("invalid flush throttle timeout", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.FlushThrottleTimeout = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), errInvalidFlushThrottleTimeout)
	})

	t.Run("invalid max packet payload size", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.MaxPacketMsgPayloadSize = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), errInvalidMaxPayloadSize)
	})

	t.Run("invalid send rate", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.SendRate = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), errInvalidSendRate)
	})

	t.Run("invalid receive rate", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.RecvRate = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), errInvalidReceiveRate)
	})

	t.Run("valid configuration", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		assert.NoError(t, cfg.ValidateBasic())
	})
}
