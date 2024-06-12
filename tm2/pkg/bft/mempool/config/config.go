package config

import "github.com/gnolang/gno/tm2/pkg/errors"

// -----------------------------------------------------------------------------
// MempoolConfig

// MempoolConfig defines the configuration options for the Tendermint mempool
type MempoolConfig struct {
	RootDir            string `toml:"home"`
	Recheck            bool   `toml:"recheck"`
	Broadcast          bool   `toml:"broadcast"`
	WalPath            string `toml:"wal_dir"`
	Size               int    `toml:"size" comment:"Maximum number of transactions in the mempool"`
	MaxPendingTxsBytes int64  `toml:"max_pending_txs_bytes" comment:"Limit the total size of all txs in the mempool.\n This only accounts for raw transactions (e.g. given 1MB transactions and\n max_txs_bytes=5MB, mempool will only accept 5 transactions)."`
	CacheSize          int    `toml:"cache_size" comment:"Size of the cache (used to filter transactions we saw earlier) in transactions"`
}

// DefaultMempoolConfig returns a default configuration for the Tendermint mempool
func DefaultMempoolConfig() *MempoolConfig {
	return &MempoolConfig{
		Recheck:   true,
		Broadcast: true,
		WalPath:   "",
		// Each signature verification takes .5ms, Size reduced until we implement
		// ABCI Recheck
		Size:               5000,
		MaxPendingTxsBytes: 1024 * 1024 * 1024, // 1GB
		CacheSize:          10000,
	}
}

// TestMempoolConfig returns a configuration for testing the Tendermint mempool
func TestMempoolConfig() *MempoolConfig {
	cfg := DefaultMempoolConfig()
	cfg.CacheSize = 1000
	return cfg
}

// WalDir returns the full path to the mempool's write-ahead log
func (cfg *MempoolConfig) WalDir() string {
	return join(cfg.RootDir, cfg.WalPath)
}

// WalEnabled returns true if the WAL is enabled.
func (cfg *MempoolConfig) WalEnabled() bool {
	return cfg.WalPath != ""
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *MempoolConfig) ValidateBasic() error {
	if cfg.Size < 0 {
		return errors.New("size can't be negative")
	}
	if cfg.MaxPendingTxsBytes < 0 {
		return errors.New("max_txs_bytes can't be negative")
	}
	if cfg.CacheSize < 0 {
		return errors.New("cache_size can't be negative")
	}
	return nil
}
