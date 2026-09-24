package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRPCConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultRPCConfig()

	// Zero preserves the historical idle behavior: net/http falls back to
	// the read timeout as the keep-alive idle deadline.
	assert.Equal(t, time.Duration(0), cfg.IdleTimeout)
	require.NoError(t, cfg.ValidateBasic())
}

func TestRPCConfigValidateBasic(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name   string
		mutate func(*RPCConfig)
	}{
		{"negative grpc_max_open_connections", func(c *RPCConfig) { c.GRPCMaxOpenConnections = -1 }},
		{"negative max_open_connections", func(c *RPCConfig) { c.MaxOpenConnections = -1 }},
		{"negative timeout_broadcast_tx_commit", func(c *RPCConfig) { c.TimeoutBroadcastTxCommit = -time.Second }},
		{"negative idle_timeout", func(c *RPCConfig) { c.IdleTimeout = -time.Second }},
		{"negative max_body_bytes", func(c *RPCConfig) { c.MaxBodyBytes = -1 }},
		{"negative max_header_bytes", func(c *RPCConfig) { c.MaxHeaderBytes = -1 }},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := DefaultRPCConfig()
			tc.mutate(cfg)
			assert.Error(t, cfg.ValidateBasic())
		})
	}
}
