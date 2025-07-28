package config

import "errors"

var (
	ErrInvalidMaxTxCount = errors.New("invalid maximum transaction count")
	ErrInvalidMaxBytes   = errors.New("invalid maximum mempool size (in bytes)")
)

// Config defines the configuration options for the Tendermint mempool
type Config struct {
	Broadcast  bool  `json:"broadcast" toml:"broadcast" comment:"Gossip transactions to other peers."`
	MaxTxCount int   `json:"max_tx_count" toml:"max_tx_count" comment:"Maximum number of transactions in the mempool (count)."`
	MaxBytes   int64 `json:"max_bytes" toml:"max_bytes" comment:"The maximum combined size of all txs in the mempool.\n This only accounts for raw transactions (e.g. given 1MB transactions and\n max_txs_bytes=5MB, mempool will only accept 5 transactions)."`
}

// DefaultConfig returns a default configuration for the Tendermint mempool
func DefaultConfig() *Config {
	return &Config{
		Broadcast:  true,
		MaxTxCount: 15000,
		MaxBytes:   1024 * 1024 * 1024, // 1GB
	}
}

// ValidateBasic performs basic validation on the mempool configuration
func (cfg *Config) ValidateBasic() error {
	if cfg.MaxBytes < 0 {
		return ErrInvalidMaxBytes
	}

	if cfg.MaxTxCount < 0 {
		return ErrInvalidMaxTxCount
	}

	return nil
}
