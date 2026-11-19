package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestP2PConfig_ValidateBasic(t *testing.T) {
	t.Parallel()

	t.Run("invalid flush throttle timeout", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.FlushThrottleTimeout = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidFlushThrottleTimeout)
	})

	t.Run("invalid max packet payload size", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.MaxPacketMsgPayloadSize = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidMaxPayloadSize)
	})

	t.Run("invalid send rate", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.SendRate = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidSendRate)
	})

	t.Run("invalid receive rate", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.RecvRate = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidReceiveRate)
	})

	t.Run("invalid recv assembly timeout", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.RecvAssemblyTimeout = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidRecvAssemblyTimeout)
	})

	t.Run("invalid max recv buffer bytes", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.MaxRecvBufferBytes = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidMaxRecvBufferBytes)
	})

	t.Run("max recv buffer bytes below one block response", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.MaxRecvBufferBytes = MinRecvBufferBytes - 1
		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidMaxRecvBufferBytes)

		cfg.MaxRecvBufferBytes = MinRecvBufferBytes
		assert.NoError(t, cfg.ValidateBasic())

		cfg.MaxRecvBufferBytes = 0
		assert.NoError(t, cfg.ValidateBasic())
	})

	t.Run("invalid max send buffer bytes", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.MaxSendBufferBytes = -1

		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidMaxSendBufferBytes)
	})

	t.Run("max send buffer bytes below one block response", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		cfg.MaxSendBufferBytes = MinSendBufferBytes - 1
		assert.ErrorIs(t, cfg.ValidateBasic(), ErrInvalidMaxSendBufferBytes)

		cfg.MaxSendBufferBytes = MinSendBufferBytes
		assert.NoError(t, cfg.ValidateBasic())

		cfg.MaxSendBufferBytes = 0
		assert.NoError(t, cfg.ValidateBasic())
	})

	t.Run("valid configuration", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()

		assert.NoError(t, cfg.ValidateBasic())
	})
}

func TestP2PConfig_AddrBookFile(t *testing.T) {
	t.Parallel()

	t.Run("default relative path", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()
		cfg.RootDir = "/root"

		assert.Equal(t, filepath.Join("/root", defaultAddrBookPath), cfg.AddrBookFile())
	})

	t.Run("empty uses default", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()
		cfg.RootDir = "/root"
		cfg.AddrBook = ""

		assert.Equal(t, filepath.Join("/root", defaultAddrBookPath), cfg.AddrBookFile())
	})

	t.Run("absolute path preserved", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()
		cfg.RootDir = "/root"
		cfg.AddrBook = "/custom/addrbook.json"

		assert.Equal(t, "/custom/addrbook.json", cfg.AddrBookFile())
	})

	t.Run("custom relative path", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultP2PConfig()
		cfg.RootDir = "/root"
		cfg.AddrBook = "peers/book.json"

		assert.Equal(t, filepath.Join("/root", "peers/book.json"), cfg.AddrBookFile())
	})
}
