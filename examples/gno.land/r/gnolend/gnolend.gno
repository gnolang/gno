package gnolend

import (
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/r/gnoswap/v1/pool"
	"std"
	"gno.land/p/demo/u256"
)

// State variables
var (
	markets *avl.Tree // marketId -> Market
	// First level: marketId -> *avl.Tree
	// Second level: userAddr -> Position
	positions *avl.Tree
	marketParams *avl.Tree // marketId -> MarketParams
	enabledIRMs *avl.Tree // irm -> bool
	enabledLLTVs *avl.Tree // lltv -> bool
	Ownable *ownable.Ownable
	feeRecipient std.Address
)

func init() {
	// Initialize market state
	markets = avl.NewTree()
	marketParams = avl.NewTree()
	positions = avl.NewTree()

	// Initialize enabled IRMs and LLTVs
	enabledIRMs = avl.NewTree()
	enabledLLTVs = avl.NewTree()

	// Set initial owner
	Ownable = ownable.New()
}

// Owner functions

func EnableIRM(irm string) {
	Ownable.AssertCallerIsOwner()

	if _, exists := enabledIRMs.Get(irm); exists {
		panic(ErrAlreadySet)
	}

	enabledIRMs.Set(irm, true)
}

func EnableLLTV(lltv int64) {
	Ownable.AssertCallerIsOwner()

	if _, exists := enabledLLTVs.Get(lltv); exists {
		panic(ErrAlreadySet)
	}

	if lltv > 100 {
		panic(ErrMaxFeeExceeded) // Reuse fee error since both are percentages
	}

	enabledLLTVs.Set(lltv, true)
}

func SetFeeRecipient(newFeeRecipient std.Address) {
	Ownable.AssertCallerIsOwner()

	if newFeeRecipient == feeRecipient {
		panic(ErrAlreadySet)
	}

	feeRecipient = newFeeRecipient
}

// CreateMarket initializes a new lending market with basic parameters
func CreateMarket(loanToken, collateralToken, oracle, irm string, lltv int64) {
	// Basic validation
	if loanToken == "" || collateralToken == "" {
		panic(ErrZeroAddress)
	}
	if loanToken == collateralToken {
		panic(ErrSameToken)
	}

	// Create market params
	params := MarketParams{
		LoanToken: loanToken,
		CollateralToken: collateralToken,
		Oracle: oracle,
		IRM: irm,
		LLTV: lltv,
	}

	// Verify tokens exist together in a Gnoswap pool
	if !areTokensPairedInGnoswap(loanToken, collateralToken) {
		panic(ErrTokenPairNotInGnoswap)
	}

	// Get market ID
	marketId := params.ID()

	// Check market doesn't exist
	if _, exists := markets.Get(marketId); exists {
		panic(ErrMarketAlreadyCreated)
	}

	// Create market with initial values
	market := Market{
		TotalSupplyAssets: new(u256.Uint),
		TotalSupplyShares: new(u256.Uint),
		TotalBorrowAssets: new(u256.Uint),
		TotalBorrowShares: new(u256.Uint),
		LastUpdate: time.Now().Unix(),
		Fee: 0,
	}

	// Store market and its params
	markets.Set(marketId, market)
	marketParams.Set(marketId, params)

	// Initialize positions tree for this market
	positions.Set(marketId, avl.NewTree())

	// Initialize IRM
	irmInstance := GetIRM(irm)
	irmInstance.BorrowRate(market, params) // Call to initialize if stateful

	// Emit market creation event
	emitCreateMarket(marketId, loanToken, collateralToken)
}

// GetMarket returns market information for a given market ID
func GetMarket(marketId string) (Market, MarketParams) {
	market, exists := markets.Get(marketId)
	if !exists {
		panic(ErrMarketNotCreated)
	}

	params, exists := marketParams.Get(marketId)
	if !exists {
		panic(ErrMarketNotCreated)
	}

	return market.(Market), params.(MarketParams)
}

// GetPosition returns a user's position in a market
func GetPosition(marketId string, userAddr string) Position {
	marketPositions, exists := positions.Get(marketId)
	if !exists {
		return Position{} // Market doesn't exist, return empty position
	}

	position, exists := marketPositions.(*avl.Tree).Get(userAddr)
	if !exists {
		return Position{} // Position doesn't exist, return empty position
	}
	return position.(Position)
}

// Supply tokens to a market
// Either assets or shares must be non-zero (not both)
func Supply(marketId string, assets, shares uint64) {
	if (assets > 0 && shares > 0) || (assets == 0 && shares == 0) {
		panic(ErrInconsistentAmount)
	}

	// Convert inputs to uint256
	assetsU256 := u256.NewInt(assets)
	sharesU256 := u256.NewInt(shares)

	// Get caller
	caller := std.PreviousRealm().Address()

	// Get market and params
	market, params := GetMarket(marketId)

	// Accrue interest before any state changes
	accrueInterest(marketId)

	// Calculate shares to mint
	var sharesToMint *u256.Uint
	if assets > 0 {
		sharesToMint = ToSharesDown(
			assetsU256,
			market.TotalSupplyAssets,
			market.TotalSupplyShares,
		)
	} else {
		sharesToMint = sharesU256
		assetsU256 = ToAssetsUp(
			sharesU256,
			market.TotalSupplyAssets,
			market.TotalSupplyShares,
		)
	}

	// Get user's current position
	position := GetPosition(marketId, caller.String())

	// Update position
	position.SupplyShares = new(u256.Uint).Add(position.SupplyShares, sharesToMint)

	// Get market's positions tree and update position
	marketPositionsInterface, exists := positions.Get(marketId)
	if !exists {
		panic(ErrMarketNotCreated)
	}
	marketPositions := marketPositionsInterface.(*avl.Tree)
	marketPositions.Set(caller.String(), position)

	// Update market state
	market.TotalSupplyShares = new(u256.Uint).Add(market.TotalSupplyShares, sharesToMint)
	market.TotalSupplyAssets = new(u256.Uint).Add(market.TotalSupplyAssets, assetsU256)
	markets.Set(marketId, market)

	// Handle token transfer using GRC20 interface
	tokenTeller := GetTokenCallerTeller(params.LoanToken)
	if err := tokenTeller.TransferFrom(caller, std.CurrentRealm().Address(), assets); err != nil {
		panic(err)
	}

	// Emit supply event
	emitDeposit(marketId, caller, assetsU256)
}

// Borrow assets from a market using collateral
// Either assets or shares must be non-zero (not both)
func Borrow(marketId string, assets, shares uint64) {
	if (assets > 0 && shares > 0) || (assets == 0 && shares == 0) {
		panic(ErrInconsistentAmount)
	}

	// Convert inputs to uint256
	assetsU256 := u256.NewInt(assets)
	sharesU256 := u256.NewInt(shares)

	// Get caller
	caller := std.PreviousRealm().Address()

	// Get market and params
	market, params := GetMarket(marketId)

	// Accrue interest before any state changes
	accrueInterest(marketId)

	// Get user's current position
	position := GetPosition(marketId, caller.String())

	// Check if user has collateral
	if position.Collateral.IsZero() {
		panic(ErrNoCollateral)
	}

	// Calculate shares to mint
	var sharesToMint *u256.Uint
	if assets > 0 {
		sharesToMint = ToSharesUp(
			assetsU256,
			market.TotalBorrowAssets,
			market.TotalBorrowShares,
		)
	} else {
		sharesToMint = sharesU256
		assetsU256 = ToAssetsDown(
			sharesU256,
			market.TotalBorrowAssets,
			market.TotalBorrowShares,
		)
	}

	// Check if market has sufficient liquidity
	availableLiquidity := new(u256.Uint).Sub(market.TotalSupplyAssets, market.TotalBorrowAssets)
	if assetsU256.Cmp(availableLiquidity) > 0 {
		panic(ErrInsufficientLiquidity)
	}

	// Get collateral value and check against LLTV
	collateralValue := getCollateralValue(params, position.Collateral)
	lltvScaled := u256.NewInt(uint64(params.LLTV))
	maxBorrow := WMulDown(collateralValue, lltvScaled)
	
	// Calculate current borrowed value including this new borrow
	newTotalBorrowed := assetsU256
	if !position.BorrowShares.IsZero() {
		currentBorrowed := ToAssetsUp(
			position.BorrowShares,
			market.TotalBorrowAssets,
			market.TotalBorrowShares,
		)
		newTotalBorrowed = new(u256.Uint).Add(newTotalBorrowed, currentBorrowed)
	}

	// Check if borrow would exceed LLTV
	if newTotalBorrowed.Cmp(maxBorrow) > 0 {
		panic(ErrExceedsLTV)
	}

	// Update position
	position.BorrowShares = new(u256.Uint).Add(position.BorrowShares, sharesToMint)

	// Get market's positions tree and update position
	marketPositionsInterface, exists := positions.Get(marketId)
	if !exists {
		panic(ErrMarketNotCreated)
	}
	marketPositions := marketPositionsInterface.(*avl.Tree)
	marketPositions.Set(caller.String(), position)

	// Update market state
	market.TotalBorrowShares = new(u256.Uint).Add(market.TotalBorrowShares, sharesToMint)
	market.TotalBorrowAssets = new(u256.Uint).Add(market.TotalBorrowAssets, assetsU256)
	markets.Set(marketId, market)

	// Transfer borrowed tokens to caller
	tokenTeller := GetTokenCallerTeller(params.LoanToken)
	if err := tokenTeller.Transfer(caller, assets); err != nil {
		panic(err)
	}

	// Emit borrow event
	emitBorrow(marketId, caller, assetsU256)
}

// getCollateralValue returns the USD value of collateral using Gnoswap price
func getCollateralValue(params MarketParams, collateralAmount *u256.Uint) *u256.Uint {
	// TODO: Use oracle from params.Oracle to get price
	// For now return 1:1 ratio for testing
	return new(u256.Uint).Set(collateralAmount)
}

// accrueInterest accrues interest for a market using its IRM
func accrueInterest(marketId string) {
	market, params := GetMarket(marketId)
	
	// Calculate elapsed time
	now := time.Now().Unix()
	elapsed := now - market.LastUpdate
	if elapsed == 0 {
		return
	}

	// Skip if no IRM or no borrows
	if params.IRM == "" || market.TotalBorrowAssets.IsZero() {
		market.LastUpdate = now
		markets.Set(marketId, market)
		return
	}

	// Get borrow rate from IRM (already WAD-scaled and per second)
	irm := GetIRM(params.IRM)
	borrowRate := irm.BorrowRate(market, params)

	// Calculate interest using Taylor series approximation
	// interest = totalBorrowAssets * (e^(rate * time) - 1)
	elapsedUint := u256.NewInt(uint64(elapsed))
	rateTimesTerm := WTaylorCompounded(borrowRate, elapsedUint)
	interest := WMulDown(market.TotalBorrowAssets, rateTimesTerm)

	// Update market state
	market.TotalBorrowAssets = new(u256.Uint).Add(market.TotalBorrowAssets, interest)
	market.TotalSupplyAssets = new(u256.Uint).Add(market.TotalSupplyAssets, interest)

	// Handle fees if any
	if market.Fee > 0 && feeRecipient != "" {
		// Calculate fee amount
		feeUint := u256.NewInt(uint64(market.Fee))
		feeAmount := WMulDown(interest, feeUint)
		
		// Calculate fee shares
		// The fee amount is subtracted from total supply to compensate for already increased supply
		totalSupplyMinusFee := new(u256.Uint).Sub(market.TotalSupplyAssets, feeAmount)
		feeShares := ToSharesDown(
			feeAmount,
			totalSupplyMinusFee,
			market.TotalSupplyShares,
		)

		// Update fee recipient's position
		feeRecipientPos := GetPosition(marketId, feeRecipient.String())
		feeRecipientPos.SupplyShares = new(u256.Uint).Add(feeRecipientPos.SupplyShares, feeShares)
		marketPositions := positions.Get(marketId).(*avl.Tree)
		marketPositions.Set(feeRecipient.String(), feeRecipientPos)

		market.TotalSupplyShares = new(u256.Uint).Add(market.TotalSupplyShares, feeShares)
	}

	market.LastUpdate = now
	markets.Set(marketId, market)

	// Emit interest accrual event
	emitAccrueInterest(marketId, borrowRate, interest)
}

// GetIRM returns the IRM instance for a given path
func GetIRM(path string) irm.IRM {
	// TODO: Implement IRM registry and lookup
	return &irm.LinearIRM{} // For now just return linear IRM
}
