package oracle

import (
	"std"
	"time"
	"strings"

	"gno.land/p/demo/ownable"
	"gno.land/r/gnoswap/v1/pool"

	"gno.land/p/gnolend/consts"
	"gno.land/p/gnolend/errors"
	"gno.land/p/gnolend/types"
)

var (
	owner *ownable.Ownable

	// Price data storage
	prices map[std.Address]types.PriceData

	// Maximum age of price data before it's considered stale (in blocks)
	maxPriceAge int64 = 100 // ~100 seconds assuming 1s block time

	// Gnoswap pool configurations for price feeds
	// asset -> (token0Path, token1Path, fee)
	poolConfigs map[std.Address]PoolConfig
)

type PoolConfig struct {
	Token0Path string
	Token1Path string
	Fee        uint32
	IsToken0   bool // true if the asset is token0 in the pool
}

func init() {
	owner = ownable.NewWithAddress(consts.ADMIN)
	prices = make(map[std.Address]types.PriceData)
	poolConfigs = make(map[std.Address]PoolConfig)
}

// SetPoolConfig sets the Gnoswap pool configuration for an asset
func SetPoolConfig(asset std.Address, token0Path, token1Path string, fee uint32, isToken0 bool) {
	owner.AssertCallerIsOwner()
	poolConfigs[asset] = PoolConfig{
		Token0Path: token0Path,
		Token1Path: token1Path,
		Fee:       fee,
		IsToken0:  isToken0,
	}
}

// SetMaxPriceAge sets the maximum age for price data
func SetMaxPriceAge(blocks int64) {
	owner.AssertCallerIsOwner()
	maxPriceAge = blocks
}

// UpdatePrice updates the price data for an asset from its Gnoswap pool
func UpdatePrice(asset std.Address) error {
	config, exists := poolConfigs[asset]
	if !exists {
		return errors.ErrPriceNotAvailable
	}

	// Get the pool
	gnoPool := pool.GetPool(config.Token0Path, config.Token1Path, config.Fee)
	if gnoPool == nil {
		return errors.ErrPriceNotAvailable
	}

	// Get the current sqrtPriceX96 from the pool
	sqrtPriceX96Str := pool.PoolGetSlot0SqrtPriceX96(gnoPool.PoolPath())
	sqrtPriceX96 := parseUint64(sqrtPriceX96Str)

	// Convert sqrtPriceX96 to actual price
	// price = (sqrtPriceX96 * sqrtPriceX96) / (2^192)
	price := (sqrtPriceX96 * sqrtPriceX96) >> 192

	// Get token decimals
	token0Decimals := uint8(6) // TODO: Get from token
	token1Decimals := uint8(6) // TODO: Get from token

	// Adjust price based on which token we're pricing
	if !config.IsToken0 {
		// If we're pricing token1, we need to invert the price
		if price == 0 {
			return errors.ErrInvalidPrice
		}
		price = (1 << 192) / price
	}

	// Adjust price for decimal differences
	decimalAdjustment := 0
	if token0Decimals > token1Decimals {
		decimalAdjustment = int(token0Decimals - token1Decimals)
		price = price * uint64(10^decimalAdjustment)
	} else if token1Decimals > token0Decimals {
		decimalAdjustment = int(token1Decimals - token0Decimals)
		price = price / uint64(10^decimalAdjustment)
	}

	// Store the updated price data
	prices[asset] = types.PriceData{
		Price:           price,
		LastUpdateBlock: std.GetHeight(),
		Decimals:       config.IsToken0 ? token0Decimals : token1Decimals,
	}

	std.Emit("PriceUpdate", 
		"asset", asset.String(),
		"price", price,
		"timestamp", std.GetTimestamp(),
	)

	return nil
}

// GetPrice returns the current price for an asset
func GetPrice(asset std.Address) (uint64, error) {
	priceData, exists := prices[asset]
	if !exists {
		return 0, errors.ErrPriceNotAvailable
	}

	// Check if price is stale
	if std.GetHeight() - priceData.LastUpdateBlock > maxPriceAge {
		return 0, errors.ErrStalePrice
	}

	return priceData.Price, nil
}

// GetPriceWithDecimals returns the price and decimals for an asset
func GetPriceWithDecimals(asset std.Address) (uint64, uint8, error) {
	priceData, exists := prices[asset]
	if !exists {
		return 0, 0, errors.ErrPriceNotAvailable
	}

	// Check if price is stale
	if std.GetHeight() - priceData.LastUpdateBlock > maxPriceAge {
		return 0, 0, errors.ErrStalePrice
	}

	return priceData.Price, priceData.Decimals, nil
}

// GetLastUpdateBlock returns the last block when price was updated
func GetLastUpdateBlock(asset std.Address) (int64, error) {
	priceData, exists := prices[asset]
	if !exists {
		return 0, errors.ErrPriceNotAvailable
	}

	return priceData.LastUpdateBlock, nil
}

// IsPriceStale checks if the price data is stale
func IsPriceStale(asset std.Address) (bool, error) {
	lastUpdate, err := GetLastUpdateBlock(asset)
	if err != nil {
		return true, err
	}

	return std.GetHeight() - lastUpdate > maxPriceAge, nil
}

// Helper function to parse uint64 from string
func parseUint64(s string) uint64 {
	// Remove any whitespace
	s = strings.TrimSpace(s)
	
	// Simple base-10 string to uint64 conversion
	var result uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		result = result*10 + uint64(c-'0')
	}
	return result
} 