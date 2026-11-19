package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// -----------------------------------------------------------------------------
// Application Config

var (
	ErrInvalidMinGasPrices  = errors.New("invalid min gas prices")
	ErrInvalidPruneStrategy = errors.New("invalid prune strategy")
	ErrInvalidQueryLimit    = errors.New("query limits must not be negative")
)

// AppConfig defines the configuration options for the Application
type AppConfig struct {
	// Lowest gas prices accepted by a validator in the form of "100tokenA/3gas;10tokenB/5gas" separated by semicolons
	MinGasPrices string `json:"min_gas_prices" toml:"min_gas_prices" comment:"Lowest gas prices accepted by a validator"`

	// The enforced state pruning stategy for the app
	PruneStrategy types.PruneStrategy `json:"prune_strategy" toml:"prune_strategy" comment:"State pruning strategy [everything, nothing, syncable]"`

	// VM budget of one query, set per node. 0 keeps the app default.
	QueryMaxGas   int64 `json:"query_max_gas" toml:"query_max_gas" comment:"Gas limit of a single VM query (qeval, qrender, ...). 0 uses the app default"`
	QueryMaxAlloc int64 `json:"query_max_alloc" toml:"query_max_alloc" comment:"Allocation limit in bytes of a single VM query. 0 uses the app default"`
	// Wall-clock limit of one query, set per node. 0 derives it from the gas
	// budget.
	QueryTimeout time.Duration `json:"query_timeout" toml:"query_timeout" comment:"Wall-clock limit of a single VM query. 0 derives it from query_max_gas (2ns per gas). Raise it on hardware slower than the gas calibration"`

	// AllowZeroFeeTxs opts this validator into accepting 0-fee transactions when
	// realms sponsor gas via PayGas. It is a per-validator mempool admission
	// policy (like MinGasPrices) and takes effect only when the chain's
	// Block.MaxGasCreditPerTx consensus param is > 0. Off by default.
	AllowZeroFeeTxs bool `json:"allow_zero_fee_txs" toml:"allow_zero_fee_txs" comment:"Accept 0-fee transactions when realms sponsor gas via PayGas (requires the MaxGasCreditPerTx consensus param > 0)"`
}

// DefaultAppConfig returns a default configuration for the application
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		MinGasPrices:    "",
		PruneStrategy:   types.PruneSyncableStrategy,
		AllowZeroFeeTxs: false,
	}
}

// ValidateBasic performs basic validation, checking format and param bounds, etc., and
// returns an error if any check fails.
func (cfg *AppConfig) ValidateBasic() error {
	// Make sure the minimum gas prices are valid, if set
	if cfg.MinGasPrices != "" {
		if _, err := std.ParseGasPrices(cfg.MinGasPrices); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidMinGasPrices, err)
		}
	}

	// Make sure the prune strategy is recognized
	if cfg.PruneStrategy != types.PruneEverythingStrategy &&
		cfg.PruneStrategy != types.PruneNothingStrategy &&
		cfg.PruneStrategy != types.PruneSyncableStrategy {
		return fmt.Errorf("%w: %q", ErrInvalidPruneStrategy, cfg.PruneStrategy)
	}

	if cfg.QueryMaxGas < 0 || cfg.QueryMaxAlloc < 0 || cfg.QueryTimeout < 0 {
		return ErrInvalidQueryLimit
	}

	return nil
}
