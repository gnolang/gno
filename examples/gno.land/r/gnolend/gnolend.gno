package gnolend

import (
	"time"

	"gno.land/p/demo/avl"
	"gno.land/r/gnoswap/v1/pool"
	"std"
	"gno.land/r/gnoswap/v1/common"
)

// State variables
var (
	markets *avl.Tree // marketId -> Market
	// First level: marketId -> *avl.Tree
	// Second level: userAddr -> Position
	positions *avl.Tree
)

func init() {
	markets = avl.NewTree()
	positions = avl.NewTree()
}

// CreateMarket initializes a new lending market
// Gnoswap pool pair must exist
func CreateMarket(loanToken, collateralToken string) {
	// Basic validation
	if loanToken == "" || collateralToken == "" {
		panic(ErrZeroAddress)
	}
	if loanToken == collateralToken {
		panic(ErrSameToken)
	}

	params := MarketParams{
		LoanToken:       loanToken,
		CollateralToken: collateralToken,
	}
	marketId := params.ID()
	if _, exists := markets.Get(marketId); exists {
		panic(ErrMarketAlreadyCreated)
	}

	// Verify tokens exist together in a Gnoswap pool
	if !areTokensPairedInGnoswap(loanToken, collateralToken) {
		panic(ErrTokenPairNotInGnoswap)
	}

	// Create market
	market := Market{
		LastUpdate:      time.Now(),
		LoanToken:       loanToken,
		CollateralToken: collateralToken,
	}
	markets.Set(marketId, market)

	// Initialize positions tree for this market
	positions.Set(marketId, avl.NewTree())

	// Emit market creation event using helper
	emitCreateMarket(marketId, loanToken, collateralToken)
}

// GetMarket returns market information for a given market ID
func GetMarket(marketId string) Market {
	market, exists := markets.Get(marketId)
	if !exists {
		panic(ErrMarketNotCreated)
	}
	return market.(Market)
}

// GetPosition returns a user's position in a market
func GetPosition(marketId string, user std.Address) Position {
	marketPositions, exists := positions.Get(marketId)
	if !exists {
		return Position{} // Market doesn't exist, return empty position
	}

	position, exists := marketPositions.(*avl.Tree).Get(user.String())
	if !exists {
		return Position{} // Position doesn't exist, return empty position
	}
	return position.(Position)
}

// Supply tokens to a market
// Either assets or shares must be non-zero (not both)
func Supply(marketId string, assets, shares int64) {
	if (assets > 0 && shares > 0) || (assets == 0 && shares == 0) {
		panic(ErrInconsistentAmount)
	}

	// Get caller
	caller := std.PreviousRealm().Address()

	// Get market
	market := GetMarket(marketId)

	// Calculate shares to mint
	sharesToMint := shares
	if assets > 0 {
		sharesToMint = ToSharesDown(
			assets,
			market.TotalSupplyAssets,
			market.TotalSupplyShares,
		)
	} else {
		assets = ToAssetsUp(
			shares,
			market.TotalSupplyAssets,
			market.TotalSupplyShares,
		)
	}

	// Get user's current position
	position := GetPosition(marketId, caller)

	// Update position
	position.SupplyShares += sharesToMint

	// Get market's positions tree and update position
	marketPositionsInterface, exists := positions.Get(marketId)
	if !exists {
		panic(ErrMarketNotCreated)
	}
	marketPositions := marketPositionsInterface.(*avl.Tree)
	marketPositions.Set(caller.String(), position)

	// Update market state
	market.TotalSupplyShares += sharesToMint
	market.TotalSupplyAssets += assets
	markets.Set(marketId, market)

	// Handle token transfer using GRC20 interface
	token := common.GetTokenTeller(market.LoanToken)
	if err := token.TransferFrom(caller, std.CurrentRealm().Address(), uint64(assets)); err != nil {
		panic(err)
	}

	// Emit supply event
	emitDeposit(marketId, caller, assets)
}

// Helper function to check if two tokens exist together in a Gnoswap pool
func areTokensPairedInGnoswap(token0, token1 string) bool {
	poolPaths := pool.PoolGetPoolList()
	for _, poolPath := range poolPaths {
		poolToken0 := pool.PoolGetToken0Path(poolPath)
		poolToken1 := pool.PoolGetToken1Path(poolPath)

		// Check if both tokens are in this pool (in either order)
		if (poolToken0 == token0 && poolToken1 == token1) ||
			(poolToken0 == token1 && poolToken1 == token0) {
			return true
		}
	}
	return false
}
