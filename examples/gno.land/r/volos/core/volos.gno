package volos

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"

	u256 "gno.land/p/gnoswap/uint256"
	pl "gno.land/r/gnoswap/v1/pool"
)

/* STATE VARIABLES */

var (
	markets *avl.Tree // marketId -> Market
	// First level: marketId -> *avl.Tree
	// Second level: userAddr -> Position
	positions    *avl.Tree
	marketParams *avl.Tree // marketId -> MarketParams
	enabledIRMs  *avl.Tree // irm -> bool
	enabledLLTVs *avl.Tree // lltv -> bool
	Ownable      *ownable.Ownable
	feeRecipient std.Address
	irmRegistry  *avl.Tree // irmName -> IRM
	// Authorization: authorizer -> (AVL tree: authorized -> bool)
	authorizers *avl.Tree
)

/* INITIALIZATION */

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

	// Initialize authorization mapping
	authorizers = avl.NewTree()

	// Set initial owner
	Ownable = ownable.NewWithAddress(std.Address("g1tzl3sgre0c2zgxfpws9xhq0c069wf7zqh6aqqy"))
}

// RegisterIRM registers a new interest rate model
func RegisterIRM(cur realm, irm IRM) {
	// Get IRM name
	name := irm.Name()

	// Check if IRM is already registered
	if _, exists := irmRegistry.Get(name); exists {
		panic(ErrIRMAlreadyRegistered)
	}

	// Register the IRM
	irmRegistry.Set(name, irm)

	// Emit registration event
	emitRegisterIRM(std.PreviousRealm().PkgPath(), name)
}

/* GOVERNANCE FUNCTIONS */

func EnableIRM(cur realm, irm string) {
	print("Current realm:", std.CurrentRealm().Address())
	print("Previous realm:", std.PreviousRealm().Address())
	print("Owner:", Ownable.Owner())
	Ownable.AssertOwnedByPrevious()

	// Check if IRM exists in registry
	if _, exists := irmRegistry.Get(irm); !exists {
		panic(ErrIRMNotRegistered)
	}

	if _, exists := enabledIRMs.Get(irm); exists {
		panic(ErrAlreadySet)
	}

	enabledIRMs.Set(irm, true)
}

func EnableLLTV(cur realm, lltv int64) {
	Ownable.AssertOwnedByPrevious()

	// Check if LLTV is greater than 100%
	if lltv > 100 {
		panic(ErrMaxFeeExceeded) // Reuse fee error since both are percentages
	}

	// Convert LLTV percentage to WAD-scaled value (e.g., 75% -> 0.75 * 1e18)
	lltvUint := u256.NewUint(uint64(lltv))
	lltvWad := MulDivDown(lltvUint, WAD, u256.NewUint(100)) // This will give us (lltv * 1e18) / 100

	// Convert LLTV to string for use as key in tree
	lltvStr := lltvWad.ToString()

	if _, exists := enabledLLTVs.Get(lltvStr); exists {
		panic(ErrAlreadySet)
	}

	enabledLLTVs.Set(lltvStr, true)
}

func SetFeeRecipient(cur realm, newFeeRecipient std.Address) {
	Ownable.AssertOwnedByPrevious()

	if newFeeRecipient == feeRecipient {
		panic(ErrAlreadySet)
	}

	feeRecipient = newFeeRecipient
}

// setFee sets the fee for a specific market
func setFee(cur realm, marketId string, newFee *u256.Uint) {
	Ownable.AssertOwnedByPrevious()

	// Get market (will panic if not found)
	market, _ := GetMarket(marketId)

	// Check if fee is already set to the same value
	if market.Fee.Cmp(newFee) == 0 {
		panic(ErrAlreadySet)
	}

	// Check fee doesn't exceed maximum
	if newFee.Cmp(MAX_FEE) > 0 {
		panic(ErrMaxFeeExceeded)
	}

	// Accrue interest using the previous fee before changing it
	accrueInterest(marketId)

	// Update market fee
	market.Fee = newFee
	markets.Set(marketId, market)

	emitSetFee(marketId, newFee.ToString())
}

/* MARKET CREATION */

// CreateMarket initializes a new lending market with basic parameters
func CreateMarket(cur realm, poolPath string, isToken0Loan bool, irm string, lltv int64) {
	if poolPath == "" {
		panic(ErrZeroAddress)
	}

	// Verify pool exists in Gnoswap
	if !pl.DoesPoolPathExist(poolPath) {
		panic(ErrTokenPairNotInGnoswap)
	}

	// Convert LLTV percentage to WAD-scaled value (e.g., 75% -> 0.75 * 1e18)
	lltvUint := u256.NewUint(uint64(lltv))
	lltvWad := MulDivDown(lltvUint, WAD, u256.NewUint(100)) // This will give us (lltv * 1e18) / 100

	// Create market params
	params := MarketParams{
		PoolPath:     poolPath,
		IRM:          irm,
		LLTV:         lltvWad,
		IsToken0Loan: isToken0Loan,
	}

	// Get market ID (same as pool path)
	marketId := params.ID()

	// Check market doesn't exist
	if _, exists := markets.Get(marketId); exists {
		panic(ErrMarketAlreadyCreated)
	}

	// Check if IRM is whitelisted
	IsIRMEnabled(irm)

	// Check if LLTV is whitelisted 
	IsLLTVEnabled(lltvWad.ToString())

	// Create market with initial values
	market := Market{
		TotalSupplyAssets: new(u256.Uint),
		TotalSupplyShares: new(u256.Uint),
		TotalBorrowAssets: new(u256.Uint),
		TotalBorrowShares: new(u256.Uint),
		LastUpdate:        time.Now().Unix(),
		Fee:               new(u256.Uint), // Initialize fee as zero
	}

	// Store market and its params
	markets.Set(marketId, market)
	marketParams.Set(marketId, params)

	// Initialize positions tree for this market
	positions.Set(marketId, avl.NewTree())

	// Get IRM from registry and verify it exists
	irmInstance := GetIRM(irm)
	if irmInstance == nil {
		panic(ErrIRMNotRegistered)
	}

	emitCreateMarket(marketId, params.GetLoanToken(), params.GetCollateralToken())
}

/* SUPPLY MANAGEMENT */

// Supply tokens to a market
// Either assets or shares must be non-zero (XOR)
func Supply(cur realm, marketId string, assets, shares uint64) {
	caller := std.PreviousRealm().Address()
	SupplyOnBehalf(cur, marketId, assets, shares, caller)
}

// Supply tokens to a market
// Either assets or shares must be non-zero (XOR)
func SupplyOnBehalf(cur realm, marketId string, assets, shares uint64, onBehalf std.Address) {
	if (assets > 0 && shares > 0) || (assets == 0 && shares == 0) {
		panic(ErrInconsistentAmount)
	}

	// Validate onBehalf is not zero address
	if onBehalf == std.Address("") {
		panic(ErrZeroAddress)
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

	// Get onBehalf's current position
	position := GetPosition(marketId, onBehalf.String())

	// Update position
	position.SupplyShares = new(u256.Uint).Add(position.SupplyShares, sharesToMint)

	// Get market's positions tree and update position
	marketPositionsInterface, _ := positions.Get(marketId)
	marketPositions := marketPositionsInterface.(*avl.Tree)
	marketPositions.Set(onBehalf.String(), position)

	// Update market state
	market.TotalSupplyShares = new(u256.Uint).Add(market.TotalSupplyShares, sharesToMint)
	market.TotalSupplyAssets = new(u256.Uint).Add(market.TotalSupplyAssets, assetsU256)
	markets.Set(marketId, market)

	// Handle token transfer using GRC20 interface
	safeTransferFrom(params.GetLoanToken(), caller, assetsU256.Int64())

	emitSupply(marketId, caller, onBehalf, assetsU256)
}

// Withdraw tokens from a market
// Either assets or shares must be non-zero (XOR)
func Withdraw(cur realm, marketId string, assets, shares uint64) {
	caller := std.PreviousRealm().Address()
	WithdrawOnBehalf(cur, marketId, assets, shares, caller, caller)
}

// Withdraw tokens from a market
// Either assets or shares must be non-zero (XOR)
func WithdrawOnBehalf(cur realm, marketId string, assets, shares uint64, onBehalf std.Address, receiver std.Address) {
	if (assets > 0 && shares > 0) || (assets == 0 && shares == 0) {
		panic(ErrInconsistentAmount)
	}

	// Validate receiver is not zero address
	if receiver == std.Address("") {
		panic(ErrZeroAddress)
	}

	// Check authorization
	if !isSenderAuthorized(onBehalf) {
		panic(ErrUnauthorized)
	}

	// Convert inputs to uint256
	assetsU256 := u256.NewUint(assets)
	sharesU256 := u256.NewUint(shares)

	caller := std.PreviousRealm().Address()

	// Get market and params
	market, params := GetMarket(marketId)

	// Accrue interest before any state changes
	accrueInterest(marketId)

	// Calculate shares to burn
	var sharesToBurn *u256.Uint
	if assets > 0 {
		sharesToBurn = ToSharesUp(
			assetsU256,
			market.TotalSupplyAssets,
			market.TotalSupplyShares,
		)
	} else {
		sharesToBurn = sharesU256
		assetsU256 = ToAssetsDown(
			sharesU256,
			market.TotalSupplyAssets,
			market.TotalSupplyShares,
		)
	}

	// Get onBehalf's current position
	position := GetPosition(marketId, onBehalf.String())

	// Check if onBehalf has enough shares
	if sharesToBurn.Gt(position.SupplyShares) {
		panic(ErrInsufficientShares)
	}

	// Update position
	position.SupplyShares = new(u256.Uint).Sub(position.SupplyShares, sharesToBurn)

	// Get market's positions tree and update position
	marketPositionsInterface, _ := positions.Get(marketId)
	marketPositions := marketPositionsInterface.(*avl.Tree)
	marketPositions.Set(onBehalf.String(), position)

	// Update market state
	market.TotalSupplyShares = new(u256.Uint).Sub(market.TotalSupplyShares, sharesToBurn)
	market.TotalSupplyAssets = new(u256.Uint).Sub(market.TotalSupplyAssets, assetsU256)
	markets.Set(marketId, market)

	// Check if there's enough liquidity after withdrawal
	if market.TotalBorrowAssets.Gt(market.TotalSupplyAssets) {
		panic(ErrInsufficientLiquidity)
	}

	// Handle token transfer using GRC20 interface (to receiver, not caller)
	safeTransferTo(params.GetLoanToken(), receiver, assetsU256.Int64())

	emitWithdraw(marketId, caller, onBehalf, receiver, assetsU256)
}

/* BORROW MANAGEMENT */

// Borrow assets from a market using collateral
// Either assets or shares must be non-zero (not both)
func Borrow(cur realm, marketId string, assets, shares uint64) {
	caller := std.PreviousRealm().Address()
	BorrowOnBehalf(cur, marketId, assets, shares, caller, caller)
}

// Borrow assets from a market using collateral
// Either assets or shares must be non-zero (not both)
func BorrowOnBehalf(cur realm, marketId string, assets, shares uint64, onBehalf std.Address, receiver std.Address) {
	if (assets > 0 && shares > 0) || (assets == 0 && shares == 0) {
		panic(ErrInconsistentAmount)
	}

	// Validate receiver is not zero address
	if receiver == std.Address("") {
		panic(ErrZeroAddress)
	}

	// Check authorization
	if !isSenderAuthorized(onBehalf) {
		panic(ErrUnauthorized)
	}

	// Convert inputs to uint256
	assetsU256 := u256.NewUint(assets)
	sharesU256 := u256.NewUint(shares)

	caller := std.PreviousRealm().Address()

	// Get market and params
	market, params := GetMarket(marketId)

	// Accrue interest before any state changes
	accrueInterest(marketId)

	// Get onBehalf's current position
	position := GetPosition(marketId, onBehalf.String())

	// Check if onBehalf has collateral
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
	marketPositionsInterface, _ := positions.Get(marketId)
	marketPositions := marketPositionsInterface.(*avl.Tree)
	marketPositions.Set(onBehalf.String(), position)

	// Check if position would be healthy after borrow
	if !isHealthy(marketId, onBehalf.String()) {
		panic(ErrExceedsLTV)
	}

	// Update market state
	market.TotalBorrowShares = new(u256.Uint).Add(market.TotalBorrowShares, sharesToMint)
	market.TotalBorrowAssets = new(u256.Uint).Add(market.TotalBorrowAssets, assetsU256)
	markets.Set(marketId, market)

	// Transfer borrowed tokens to receiver
	safeTransferTo(params.GetLoanToken(), receiver, assetsU256.Int64())

	emitBorrow(marketId, caller, onBehalf, receiver, assetsU256)
}

// Repay borrowed tokens to a market
// Either assets or shares must be non-zero (XOR)
func Repay(cur realm, marketId string, assets, shares uint64) {
	caller := std.PreviousRealm().Address()
	RepayOnBehalf(cur, marketId, assets, shares, caller)
}

// Repay borrowed tokens to a market
// Either assets or shares must be non-zero (XOR)
func RepayOnBehalf(cur realm, marketId string, assets, shares uint64, onBehalf std.Address) {
	if (assets > 0 && shares > 0) || (assets == 0 && shares == 0) {
		panic(ErrInconsistentAmount)
	}

	// Validate onBehalf is not zero address
	if onBehalf == std.Address("") {
		panic(ErrZeroAddress)
	}

	// Convert inputs to uint256
	assetsU256 := u256.NewUint(assets)
	sharesU256 := u256.NewUint(shares)

	caller := std.PreviousRealm().Address()

	// Get market and params
	market, params := GetMarket(marketId)

	// Accrue interest before any state changes
	accrueInterest(marketId)

	// Calculate shares to burn
	var sharesToBurn *u256.Uint
	if assets > 0 {
		sharesToBurn = ToSharesDown(
			assetsU256,
			market.TotalBorrowAssets,
			market.TotalBorrowShares,
		)
	} else {
		sharesToBurn = sharesU256
		assetsU256 = ToAssetsUp(
			sharesU256,
			market.TotalBorrowAssets,
			market.TotalBorrowShares,
		)
	}

	// Get onBehalf's current position
	position := GetPosition(marketId, onBehalf.String())

	// Check if onBehalf has enough shares
	if sharesToBurn.Gt(position.BorrowShares) {
		panic(ErrInsufficientShares)
	}

	// Update position
	position.BorrowShares = new(u256.Uint).Sub(position.BorrowShares, sharesToBurn)

	// Get market's positions tree and update position
	marketPositionsInterface, _ := positions.Get(marketId)
	marketPositions := marketPositionsInterface.(*avl.Tree)
	marketPositions.Set(onBehalf.String(), position)

	// Update market state
	market.TotalBorrowShares = new(u256.Uint).Sub(market.TotalBorrowShares, sharesToBurn)
	market.TotalBorrowAssets = new(u256.Uint).Sub(market.TotalBorrowAssets, assetsU256)
	markets.Set(marketId, market)

	// Handle token transfer using GRC20 interface 
	safeTransferFrom(params.GetLoanToken(), caller, assetsU256.Int64())

	emitRepay(marketId, caller, onBehalf, assetsU256, sharesToBurn)
}

/* COLLATERAL MANAGEMENT */

// Supply collateral to a market
func SupplyCollateral(cur realm, marketId string, amount uint64) {
	caller := std.PreviousRealm().Address()
	SupplyCollateralOnBehalf(cur, marketId, amount, caller)
}

// SupplyCollateral supplies collateral to a market
// The collateral can be used to borrow assets from the market
func SupplyCollateralOnBehalf(cur realm, marketId string, amount uint64, onBehalf std.Address) {
	// Validate onBehalf is not zero address
	if onBehalf == std.Address("") {
		panic(ErrZeroAddress)
	}

	caller := std.PreviousRealm().Address()

	_, params := GetMarket(marketId)

	// Get onBehalf's current position
	position := GetPosition(marketId, onBehalf.String())

	// Update position
	position.Collateral = new(u256.Uint).Add(position.Collateral, u256.NewUint(amount))

	// Get market's positions tree and update position
	marketPositionsInterface, _ := positions.Get(marketId)
	marketPositions := marketPositionsInterface.(*avl.Tree)
	marketPositions.Set(onBehalf.String(), position)

	// Handle token transfer using GRC20 interface
	safeTransferFrom(params.GetCollateralToken(), caller, int64(amount))

	emitSupplyCollateral(marketId, caller, onBehalf, amount)
}

// Withdraw collateral from a market
func WithdrawCollateral(cur realm, marketId string, amount uint64) {
	caller := std.PreviousRealm().Address()
	WithdrawCollateralOnBehalf(cur, marketId, amount, caller, caller)
}

// WithdrawCollateral withdraws collateral from a market
// The withdrawal will fail if it would make the user's position unhealthy
func WithdrawCollateralOnBehalf(cur realm, marketId string, amount uint64, onBehalf std.Address, receiver std.Address) {
	// Validate receiver is not zero address
	if receiver == std.Address("") {
		panic(ErrZeroAddress)
	}

	// Check authorization
	if !isSenderAuthorized(onBehalf) {
		panic(ErrUnauthorized)
	}

	caller := std.PreviousRealm().Address()

	_, params := GetMarket(marketId)

	// Accrue interest before any state changes
	accrueInterest(marketId)

	// Get onBehalf's current position
	position := GetPosition(marketId, onBehalf.String())

	// Convert amount to uint256
	amountU256 := u256.NewUint(amount)

	// Check if onBehalf has enough collateral
	if amountU256.Gt(position.Collateral) {
		panic(ErrInsufficientCollateral)
	}

	// Update position
	position.Collateral = new(u256.Uint).Sub(position.Collateral, amountU256)

	// Get market's positions tree and update position
	marketPositionsInterface, _ := positions.Get(marketId)
	marketPositions := marketPositionsInterface.(*avl.Tree)
	marketPositions.Set(onBehalf.String(), position)

	// Check if position would still be healthy after withdrawal
	if !position.BorrowShares.IsZero() {
		if !isHealthy(marketId, onBehalf.String()) {
			panic(ErrExceedsLTV)
		}
	}

	// Handle token transfer using GRC20 interface
	safeTransferTo(params.GetCollateralToken(), receiver, int64(amount))

	emitWithdrawCollateral(marketId, caller, onBehalf, receiver, amount)
}

/* LIQUIDATION */

// Liquidate liquidates a position that is below the liquidation threshold.
// It takes either seizedAssets (collateral to seize) or repaidShares (debt to repay), but not both.
func Liquidate(cur realm, marketId string, borrower std.Address, seizedAssets, repaidShares uint64) (uint64, uint64) {
	// Check that exactly one of seizedAssets or repaidShares is non-zero
	if (seizedAssets == 0) == (repaidShares == 0) {
		panic(ErrInconsistentAmount)
	}

	// Get market data
	market, params := GetMarket(marketId)

	// Convert inputs to uint256
	seizedAssetsU256 := u256.NewUint(seizedAssets)
	repaidSharesU256 := u256.NewUint(repaidShares)

	// Accrue interest before making state changes
	accrueInterest(marketId)

	// Get borrower's position
	borrowerPos := GetPosition(marketId, borrower.String())

	// Get collateral price from oracle
	collateralPrice := GetPrice(marketId)

	// Check if position is unhealthy
	if isHealthy(marketId, borrower.String()) {
		panic(ErrHealthyPosition)
	}

	// Calculate the liquidation incentive factor: min(maxLiquidationIncentiveFactor, 1/(1 - cursor*(1 - lltv)))
	incentiveFactor := WDivDown(
		WAD,
		new(u256.Uint).Sub(WAD, WMulDown(LIQUIDATION_CURSOR, new(u256.Uint).Sub(WAD, params.LLTV))),
	)
	incentiveFactor = Min(incentiveFactor, MAX_LIQUIDATION_INCENTIVE_FACTOR)

	// Calculate seized assets or repaid shares based on input
	if seizedAssets > 0 {
		// Calculate repaid shares from seized assets
		seizedAssetsQuoted := MulDivUp(seizedAssetsU256, collateralPrice, ORACLE_PRICE_SCALE)
		repaidSharesU256 = ToSharesUp(
			WDivUp(seizedAssetsQuoted, incentiveFactor),
			market.TotalBorrowAssets,
			market.TotalBorrowShares,
		)
	} else {
		// Calculate seized assets from repaid shares
		seizedAssetsU256 = MulDivDown(
			WMulDown(
				ToAssetsDown(
					repaidSharesU256,
					market.TotalBorrowAssets,
					market.TotalBorrowShares,
				),
				incentiveFactor,
			),
			ORACLE_PRICE_SCALE,
			collateralPrice,
		)
	}

	// Calculate repaid assets
	repaidAssets := ToAssetsUp(
		repaidSharesU256,
		market.TotalBorrowAssets,
		market.TotalBorrowShares,
	)

	// Update borrower's position
	borrowerPos.BorrowShares = new(u256.Uint).Sub(borrowerPos.BorrowShares, repaidSharesU256)
	borrowerPos.Collateral = new(u256.Uint).Sub(borrowerPos.Collateral, seizedAssetsU256)

	// Update market state
	market.TotalBorrowShares = new(u256.Uint).Sub(market.TotalBorrowShares, repaidSharesU256)
	market.TotalBorrowAssets = new(u256.Uint).Sub(market.TotalBorrowAssets, repaidAssets)
	markets.Set(marketId, market)

	// Handle bad debt if all collateral is seized
	if borrowerPos.Collateral.IsZero() {
		badDebtShares := borrowerPos.BorrowShares
		badDebtAssets := ToAssetsUp(
			badDebtShares,
			market.TotalBorrowAssets,
			market.TotalBorrowShares,
		)
		if badDebtAssets.Cmp(market.TotalBorrowAssets) > 0 {
			badDebtAssets = market.TotalBorrowAssets
		}

		market.TotalBorrowAssets = new(u256.Uint).Sub(market.TotalBorrowAssets, badDebtAssets)
		market.TotalSupplyAssets = new(u256.Uint).Sub(market.TotalSupplyAssets, badDebtAssets)
		market.TotalBorrowShares = new(u256.Uint).Sub(market.TotalBorrowShares, badDebtShares)
		borrowerPos.BorrowShares = u256.Zero()
	}

	// Transfer seized collateral to liquidator
	caller := std.PreviousRealm().Address()
	safeTransferTo(params.GetCollateralToken(), caller, seizedAssetsU256.Int64())

	// Transfer repaid assets from liquidator to contract
	safeTransferFrom(params.GetLoanToken(), caller, repaidAssets.Int64())

	// Emit liquidate event
	emitLiquidate(marketId, caller, borrower, repaidAssets, repaidSharesU256, seizedAssetsU256)

	return seizedAssetsU256.Uint64(), repaidAssets.Uint64()
}

/* AUTHORIZATION */

// getAuthorizeds gets or creates the authorized AVL tree for an authorizer
func getAuthorizeds(authorizer std.Address) *avl.Tree {
	if authorizedsTree, exists := authorizers.Get(authorizer.String()); exists {
		return authorizedsTree.(*avl.Tree)
	}
	authorizeds := avl.NewTree()
	authorizers.Set(authorizer.String(), authorizeds)
	return authorizeds
}

// SetAuthorization allows authorized to act on behalf of msg.sender
func SetAuthorization(cur realm, authorized std.Address, isAuth bool) {
	authorizer := std.PreviousRealm().Address()
	authorizeds := getAuthorizeds(authorizer)

	// Check if authorization is already set to the same value
	if currentAuth, exists := authorizeds.Get(authorized.String()); exists {
		if currentAuth.(bool) == isAuth {
			panic(ErrAlreadySet)
		}
	}

	authorizeds.Set(authorized.String(), isAuth)
	emitAuthorizationSet(authorizer, authorized, isAuth)
}

// IsAuthorized checks if authorized can act on behalf of authorizer
func IsAuthorized(authorizer std.Address, authorized std.Address) bool {
	if authorizedsTree, exists := authorizers.Get(authorizer.String()); exists {
		authorizeds := authorizedsTree.(*avl.Tree)
		if auth, exists := authorizeds.Get(authorized.String()); exists {
			return auth.(bool)
		}
	}
	return false
}

// isSenderAuthorized checks if msg.sender can act on behalf of onBehalf
func isSenderAuthorized(onBehalf std.Address) bool {
	sender := std.PreviousRealm().Address()
	if sender == onBehalf {
		return true
	}
	return IsAuthorized(onBehalf, sender)
}


/* INTEREST ACCRUAL */

func AccrueInterest(cur realm, marketId string) {
	accrueInterest(marketId)
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

	// Calculate accrued interest using Taylor series approximation of e^(rate * time) - 1
	// wTaylorCompounded returns the sum of first 3 terms: x*n + (x*n)^2/2 + (x*n)^3/6
	// This approximates continuous compound interest more accurately than simple interest
	elapsedUint := u256.NewUint(uint64(elapsed))
	wtc := WTaylorCompounded(borrowRate, elapsedUint)

	interest := WMulDown(market.TotalBorrowAssets, wtc)

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
		marketPositionsInterface, _ := positions.Get(marketId)
		marketPositions := marketPositionsInterface.(*avl.Tree)
		marketPositions.Set(feeRecipient.String(), feeRecipientPos)

		market.TotalSupplyShares = new(u256.Uint).Add(market.TotalSupplyShares, feeShares)
	}

	market.LastUpdate = now
	markets.Set(marketId, market)

	emitAccrueInterest(marketId, borrowRate, interest)
}

/* HEALTH CALCULATIONS */

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

	market, params := GetMarket(marketId)

	// Calculate current borrowed value
	borrowed := ToAssetsUp(
		position.BorrowShares,
		market.TotalBorrowAssets,
		market.TotalBorrowShares,
	)

	collateralPrice := GetPrice(params.PoolPath)

	// Calculate max borrow allowed
	maxBorrow := WMulDown(MulDivDown(position.Collateral, collateralPrice, ORACLE_PRICE_SCALE), params.LLTV)

	// Position is healthy if borrowed <= maxBorrow
	return borrowed.Cmp(maxBorrow) <= 0
}

/* FLASH LOANS */

// FlashLoan allows users to borrow assets without collateral, provided they are repaid within the same transaction
func FlashLoan(cur realm, token string, assets int64, data []byte, callback FlashLoanCallback) {
	if assets == 0 {
		panic(ErrZeroAssets)
	}

	caller := std.PreviousRealm().Address()
	
	// Emit flash loan event
	emitFlashLoan(caller, token, assets)

	// Transfer tokens to borrower
	safeTransferTo(token, caller, assets)

	// Execute the callback function
	callback.OnVolosFlashLoan(cross, assets, data)

	// Transfer tokens back from borrower
	safeTransferFrom(token, caller, assets)
}

