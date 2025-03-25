package gnolend

import (
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"std"
	u256 "gno.land/p/demo/uint256"
	pl "gno.land/r/gnoswap/v1/pool"
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
	irmRegistry *avl.Tree // irmName -> IRM
)

func init() {
	// Initialize market state
	markets = avl.NewTree()
	marketParams = avl.NewTree()
	positions = avl.NewTree()

	// Initialize enabled IRMs and LLTVs
	enabledIRMs = avl.NewTree()
	enabledLLTVs = avl.NewTree()

	// Initialize IRM registry
	irmRegistry = avl.NewTree()

	// Set initial owner
	Ownable = ownable.New()
}

// RegisterIRM registers a new interest rate model
func RegisterIRM(irm IRM) {
	// Get IRM name
	name := irm.Name()
	
	// Check if IRM is already registered
	if _, exists := irmRegistry.Get(name); exists {
		panic("IRM already registered")
	}

	// Register the IRM
	irmRegistry.Set(name, irm)
	
	// Emit registration event
	emitRegisterIRM(std.PreviousRealm().PkgPath(), name)
}

// Owner functions

func EnableIRM(irm string) {
	Ownable.AssertCallerIsOwner()

	// Check if IRM exists in registry
	if _, exists := irmRegistry.Get(irm); !exists {
		panic("IRM not registered: " + irm)
	}

	if _, exists := enabledIRMs.Get(irm); exists {
		panic(ErrAlreadySet)
	}

	enabledIRMs.Set(irm, true)
}

func EnableLLTV(lltv *u256.Uint) {
	Ownable.AssertCallerIsOwner()

	// Convert LLTV to string for use as key in tree
	lltvStr := lltv.String()

	if _, exists := enabledLLTVs.Get(lltvStr); exists {
		panic(ErrAlreadySet)
	}

	// Check if LLTV is greater than 100%
	maxLLTV := u256.NewUint(100)
	if lltv.Gt(maxLLTV) {
		panic(ErrMaxFeeExceeded) // Reuse fee error since both are percentages
	}

	enabledLLTVs.Set(lltvStr, true)
}

func SetFeeRecipient(newFeeRecipient std.Address) {
	Ownable.AssertCallerIsOwner()

	if newFeeRecipient == feeRecipient {
		panic(ErrAlreadySet)
	}

	feeRecipient = newFeeRecipient
}

// CreateMarket initializes a new lending market with basic parameters
func CreateMarket(poolPath string, isToken0Loan bool, irm string, lltv int64) {
	if poolPath == "" {
		panic(ErrZeroAddress)
	}

	// Verify pool exists in Gnoswap
	if !pl.DoesPoolPathExist(poolPath) {
		panic(ErrTokenPairNotInGnoswap)
	}

	// Convert LLTV percentage to WAD-scaled value (e.g., 75% -> 0.75 * 1e18)
	lltvUint := u256.NewUint(uint64(lltv))
	lltvWad := WMulDown(lltvUint, WAD) // This will give us (lltv * 1e18) / 100

	// Create market params
	params := MarketParams{
		PoolPath: poolPath,
		IRM: irm,
		LLTV: lltvWad,
		IsToken0Loan: isToken0Loan,
	}

	// Get market ID (same as pool path)
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
		Fee: new(u256.Uint), // Initialize fee as zero
	}

	// Store market and its params
	markets.Set(marketId, market)
	marketParams.Set(marketId, params)

	// Initialize positions tree for this market
	positions.Set(marketId, avl.NewTree())

	// Get IRM from registry and verify it exists
	irmInstance := GetIRM(irm)
	if irmInstance == nil {
		panic("IRM not found: " + irm)
	}

	emitCreateMarket(marketId, params.GetLoanToken(), params.GetCollateralToken())
}

// Supply tokens to a market
// Either assets or shares must be non-zero (not both)
func Supply(marketId string, assets, shares uint64) {
	if (assets > 0 && shares > 0) || (assets == 0 && shares == 0) {
		panic(ErrInconsistentAmount)
	}

	// Convert inputs to uint256
	assetsU256 := u256.NewUint(assets)
	sharesU256 := u256.NewUint(shares)

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
	tokenTeller := GetTokenCallerTeller(params.GetLoanToken())
	if err := tokenTeller.TransferFrom(caller, std.CurrentRealm().Address(), assets); err != nil {
		panic(err)
	}

	emitDeposit(marketId, caller, assetsU256)
}

// Borrow assets from a market using collateral
// Either assets or shares must be non-zero (not both)
func Borrow(marketId string, assets, shares uint64) {
	if (assets > 0 && shares > 0) || (assets == 0 && shares == 0) {
		panic(ErrInconsistentAmount)
	}

	// Convert inputs to uint256
	assetsU256 := u256.NewUint(assets)
	sharesU256 := u256.NewUint(shares)

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

	// Update position
	position.BorrowShares = new(u256.Uint).Add(position.BorrowShares, sharesToMint)

	// Get market's positions tree and update position temporarily to check health
	marketPositionsInterface, exists := positions.Get(marketId)
	if !exists {
		panic(ErrMarketNotCreated)
	}
	marketPositions := marketPositionsInterface.(*avl.Tree)
	marketPositions.Set(caller.String(), position)

	// Check if position would be healthy after borrow
	if !isHealthy(marketId, caller.String()) {
		panic(ErrExceedsLTV)
	}

	// Update market state
	market.TotalBorrowShares = new(u256.Uint).Add(market.TotalBorrowShares, sharesToMint)
	market.TotalBorrowAssets = new(u256.Uint).Add(market.TotalBorrowAssets, assetsU256)
	markets.Set(marketId, market)

	// Transfer borrowed tokens to caller
	tokenTeller := GetTokenCallerTeller(params.GetLoanToken())
	if err := tokenTeller.Transfer(caller, assets); err != nil {
		panic(err)
	}

	emitBorrow(marketId, caller, assetsU256)
}

// getCollateralValue returns the value of collateral using Gnoswap price
func getCollateralValue(params MarketParams, collateralAmount *u256.Uint) *u256.Uint {
	// Get price from oracle (using pool path)
	price, err := GetPrice(params.PoolPath)
	if err != nil {
		panic(err)
	}

	// Convert collateral amount to WAD
	collateralWAD := new(u256.Uint).Mul(collateralAmount, WAD)

	// Calculate collateral value in loan token terms
	collateralValue := WMulDown(collateralWAD, price)

	return collateralValue
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
	borrowRate := irm.BorrowRate(market.TotalSupplyAssets, market.TotalBorrowAssets)

	// Calculate interest using Taylor series approximation
	// interest = totalBorrowAssets * (e^(rate * time) - 1)
	elapsedUint := u256.NewUint(uint64(elapsed))
	rateTimesTerm := WTaylorCompounded(borrowRate, elapsedUint)
	interest := WMulDown(market.TotalBorrowAssets, rateTimesTerm)

	// Update market state
	market.TotalBorrowAssets = new(u256.Uint).Add(market.TotalBorrowAssets, interest)
	market.TotalSupplyAssets = new(u256.Uint).Add(market.TotalSupplyAssets, interest)

	// Handle fees if any
	if !market.Fee.IsZero() && feeRecipient != "" {
		// Calculate fee amount
		feeAmount := WMulDown(interest, market.Fee)
		
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
		marketPositionsInterface, exists := positions.Get(marketId)
		if !exists {
			panic(ErrMarketNotCreated)
		}
		marketPositions := marketPositionsInterface.(*avl.Tree)
		marketPositions.Set(feeRecipient.String(), feeRecipientPos)

		market.TotalSupplyShares = new(u256.Uint).Add(market.TotalSupplyShares, feeShares)
	}

	market.LastUpdate = now
	markets.Set(marketId, market)

	emitAccrueInterest(marketId, borrowRate, interest)
}

// SupplyCollateral supplies collateral to a market
// The collateral can be used to borrow assets from the market
func SupplyCollateral(marketId string, amount uint64) {
	caller := std.PreviousRealm().Address()

	_, params := GetMarket(marketId)

	// Get user's current position
	position := GetPosition(marketId, caller.String())

	// Update position
	position.Collateral = new(u256.Uint).Add(position.Collateral, u256.NewUint(amount))

	// Get market's positions tree and update position
	marketPositionsInterface, exists := positions.Get(marketId)
	if !exists {
		panic(ErrMarketNotCreated)
	}
	marketPositions := marketPositionsInterface.(*avl.Tree)
	marketPositions.Set(caller.String(), position)

	// Handle token transfer using GRC20 interface
	tokenTeller := GetTokenCallerTeller(params.GetCollateralToken())
	if err := tokenTeller.TransferFrom(caller, std.CurrentRealm().Address(), amount); err != nil {
		panic(err)
	}

	emitSupplyCollateral(marketId, caller, amount)
}

// WithdrawCollateral withdraws collateral from a market
// The withdrawal will fail if it would make the user's position unhealthy
func WithdrawCollateral(marketId string, amount uint64) {
	caller := std.PreviousRealm().Address()

	_, params := GetMarket(marketId)

	// Accrue interest before any state changes
	accrueInterest(marketId)

	// Get user's current position
	position := GetPosition(marketId, caller.String())

	// Convert amount to uint256
	amountU256 := u256.NewUint(amount)

	// Check if user has enough collateral
	if amountU256.Gt(position.Collateral) {
		panic(ErrInsufficientCollateral)
	}

	// Update position
	position.Collateral = new(u256.Uint).Sub(position.Collateral, amountU256)

	// Check if position would still be healthy after withdrawal
	if !position.BorrowShares.IsZero() {
		// Get market's positions tree and update position temporarily to check health
		marketPositionsInterface, exists := positions.Get(marketId)
		if !exists {
			panic(ErrMarketNotCreated)
		}
		marketPositions := marketPositionsInterface.(*avl.Tree)
		marketPositions.Set(caller.String(), position)

		// Check if position would be healthy after withdrawal
		if !isHealthy(marketId, caller.String()) {
			panic(ErrExceedsLTV)
		}
	}

	// Handle token transfer using GRC20 interface
	tokenTeller := GetTokenCallerTeller(params.GetCollateralToken())
	if err := tokenTeller.Transfer(caller, amount); err != nil {
		panic(err)
	}

	emitWithdrawCollateral(marketId, caller, amount)
}

// isHealthy checks if a position's health factor is above 1
// Returns true if:
// 1. The user has no borrows, or
// 2. The user's collateral value * LLTV >= borrowed value
func isHealthy(marketId string, userAddr string) bool {
	position := GetPosition(marketId, userAddr)

	// If no borrows, position is healthy
	if position.BorrowShares.IsZero() {
		return true
	}

	// Get market and params
	market, params := GetMarket(marketId)

	// Calculate current borrowed value
	borrowed := ToAssetsUp(
		position.BorrowShares,
		market.TotalBorrowAssets,
		market.TotalBorrowShares,
	)

	// Get collateral value
	collateralValue := getCollateralValue(params, position.Collateral)

	// Print collateral value for debugging
	println("collateral value:", collateralValue.String())

	// Calculate max borrow allowed
	maxBorrow := WMulDown(collateralValue, params.LLTV)

	// Position is healthy if borrowed <= maxBorrow
	return borrowed.Cmp(maxBorrow) <= 0
}