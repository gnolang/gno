package config

import (
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// -----------------------------------------------------------------------------
// Application Config

// AppConfig defines the configuration options for the Application
type AppConfig struct {
	// Lowest gas prices accepted by a validator in the form of "100tokenA/3gas;10tokenB/5gas" separated by semicolons
	MinGasPrices string `json:"min_gas_prices" toml:"min_gas_prices" comment:"Lowest gas prices accepted by a validator"`
}

// DefaultAppConfig returns a default configuration for the application
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		MinGasPrices: "",
	}
}

// ValidateBasic performs basic validation, checking format and param bounds, etc., and
// returns an error if any check fails.
func (cfg *AppConfig) ValidateBasic() error {
	if cfg.MinGasPrices == "" {
		return nil
	}
	if _, err := std.ParseGasPrices(cfg.MinGasPrices); err != nil {
		return errors.Wrap(err, "invalid min gas prices")
	}

	return nil
}
