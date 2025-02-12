package pool

import (
	"std"
	"time"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ownable"
	"gno.land/r/demo/grc20reg"

	"gno.land/p/gnolend/consts"
	"gno.land/p/gnolend/errors"
	"gno.land/p/gnolend/math"
	"gno.land/p/gnolend/types"
	"gno.land/r/gnolend/gtoken"
	"gno.land/r/gnolend/oracle"
)

var (
	owner *ownable.Ownable

	// Pool registry
	pools map[std.Address]*types.LendingPool

	// User positions
	positions map[std.Address]*types.UserPosition

	// Interest rate parameters
	interestRateParams map[std.Address]types.InterestRateParams
)

func init() {
	owner = ownable.NewWithAddress(consts.ADMIN)
	pools = make(map[std.Address]*types.LendingPool)
	positions = make(map[std.Address]*types.UserPosition)
	interestRateParams = make(map[std.Address]types.InterestRateParams)
}

// CreatePool creates a new lending pool for an asset
func CreatePool(asset std.Address, ltv, liquidationThreshold, liquidationBonus uint64) error {
	owner.AssertCallerIsOwner()

	if _, exists := pools[asset]; exists {
		return errors.ErrPoolAlreadyExists
	}

	// Create gToken for the pool
	assetToken := grc20reg.MustGet(asset)
	gTokenName := "GnoLend " + assetToken.GetName()
	gTokenSymbol := "g" + assetToken.GetSymbol()
	gtoken.InitToken(asset, gTokenName, gTokenSymbol)

	// Create pool
	pool := &types.LendingPool{
		AssetAddress:        asset,
		GTokenAddress:       gtoken.GetAddress(),
		TotalSupply:        0,
		TotalBorrow:        0,
		LiquidityIndex:     uint64(consts.PRECISION),
		BorrowIndex:        uint64(consts.PRECISION),
		LastUpdateBlock:    std.GetHeight(),
		LTV:                ltv,
		LiquidationThreshold: liquidationThreshold,
		LiquidationBonus:   liquidationBonus,
	}

	pools[asset] = pool

	// Set default interest rate parameters
	interestRateParams[asset] = types.InterestRateParams{
		BaseRate:           200,  // 2%
		Slope1:            1000, // 10%
		Slope2:            5000, // 50%
		OptimalUtilization: 8000, // 80%
	}

	return nil
}

// Supply assets to a lending pool
func Supply(asset std.Address, amount uint64) error {
	pool, exists := pools[asset]
	if !exists {
		return errors.ErrPoolNotFound
	}

	caller := std.GetOrigCaller()

	// Transfer assets from user
	assetToken := grc20reg.MustGet(asset)
	err := assetToken.TransferFrom(caller, std.CurrentRealm(), amount)
	if err != nil {
		return err
	}

	// Update indices
	updatePoolIndices(pool)

	// Mint gTokens to user
	err = gtoken.Mint(caller, amount)
	if err != nil {
		return err
	}

	// Update pool state
	pool.TotalSupply += amount

	// Update user position
	position := getOrCreatePosition(caller)
	supply, exists := position.SuppliedAssets[asset]
	if !exists {
		supply = types.Supply{
			Amount:         amount,
			IsCollateral:   true, // Enable as collateral by default
			LiquidityIndex: pool.LiquidityIndex,
		}
	} else {
		supply.Amount += amount
	}
	position.SuppliedAssets[asset] = supply

	return nil
}

// Borrow assets from a lending pool
func Borrow(asset std.Address, amount uint64) error {
	pool, exists := pools[asset]
	if !exists {
		return errors.ErrPoolNotFound
	}

	caller := std.GetOrigCaller()
	position := getOrCreatePosition(caller)

	// Check borrowing capacity
	if !canBorrow(position, asset, amount) {
		return errors.ErrInsufficientCollateral
	}

	// Update indices
	updatePoolIndices(pool)

	// Transfer assets to user
	assetToken := grc20reg.MustGet(asset)
	err := assetToken.Transfer(caller, amount)
	if err != nil {
		return err
	}

	// Update pool state
	pool.TotalBorrow += amount

	// Update user position
	borrow := types.Borrow{
		Amount:      amount,
		BorrowIndex: pool.BorrowIndex,
	}
	position.BorrowedAssets[asset] = borrow

	// Update health factor
	updateHealthFactor(position)

	return nil
}

// Repay borrowed assets
func Repay(asset std.Address, amount uint64) error {
	pool, exists := pools[asset]
	if !exists {
		return errors.ErrPoolNotFound
	}

	caller := std.GetOrigCaller()
	position := getOrCreatePosition(caller)

	borrow, exists := position.BorrowedAssets[asset]
	if !exists {
		return errors.ErrInvalidAmount
	}

	// Update indices
	updatePoolIndices(pool)

	// Calculate total debt with interest
	totalDebt := calculateDebtWithInterest(borrow.Amount, borrow.BorrowIndex, pool.BorrowIndex)
	amountToRepay := amount
	if amount > totalDebt {
		amountToRepay = totalDebt
	}

	// Transfer assets from user
	assetToken := grc20reg.MustGet(asset)
	err := assetToken.TransferFrom(caller, std.CurrentRealm(), amountToRepay)
	if err != nil {
		return err
	}

	// Update pool state
	pool.TotalBorrow -= amountToRepay

	// Update user position
	if amountToRepay == totalDebt {
		delete(position.BorrowedAssets, asset)
	} else {
		borrow.Amount = totalDebt - amountToRepay
		borrow.BorrowIndex = pool.BorrowIndex
		position.BorrowedAssets[asset] = borrow
	}

	// Update health factor
	updateHealthFactor(position)

	return nil
}

// Withdraw supplied assets
func Withdraw(asset std.Address, gTokenAmount uint64) error {
	pool, exists := pools[asset]
	if !exists {
		return errors.ErrPoolNotFound
	}

	caller := std.GetOrigCaller()
	position := getOrCreatePosition(caller)

	// Update indices
	updatePoolIndices(pool)

	// Calculate underlying amount
	underlyingAmount := gtoken.CalculateUnderlyingAmount(gTokenAmount)

	// Check if withdrawal would make position unhealthy
	if !canWithdraw(position, asset, underlyingAmount) {
		return errors.ErrPositionUnhealthy
	}

	// Burn gTokens
	actualAmount, err := gtoken.Burn(caller, gTokenAmount)
	if err != nil {
		return err
	}

	// Transfer assets to user
	assetToken := grc20reg.MustGet(asset)
	err = assetToken.Transfer(caller, actualAmount)
	if err != nil {
		return err
	}

	// Update pool state
	pool.TotalSupply -= actualAmount

	// Update user position
	supply := position.SuppliedAssets[asset]
	if actualAmount >= supply.Amount {
		delete(position.SuppliedAssets, asset)
	} else {
		supply.Amount -= actualAmount
		position.SuppliedAssets[asset] = supply
	}

	// Update health factor
	updateHealthFactor(position)

	return nil
}

// Liquidate an unhealthy position
func Liquidate(user std.Address, collateralAsset, debtAsset std.Address, debtAmount uint64) error {
	userPosition := getOrCreatePosition(user)
	if userPosition.HealthFactor > uint64(consts.PRECISION) {
		return errors.ErrPositionNotLiquidatable
	}

	collateralPool, exists := pools[collateralAsset]
	if !exists {
		return errors.ErrPoolNotFound
	}

	debtPool, exists := pools[debtAsset]
	if !exists {
		return errors.ErrPoolNotFound
	}

	// Calculate liquidation values
	collateralPrice, err := oracle.GetPrice(collateralAsset)
	if err != nil {
		return err
	}

	debtPrice, err := oracle.GetPrice(debtAsset)
	if err != nil {
		return err
	}

	// Calculate collateral to seize
	collateralAmount := (debtAmount * debtPrice * (10000 + collateralPool.LiquidationBonus)) / (collateralPrice * 10000)

	caller := std.GetOrigCaller()

	// Transfer debt tokens from liquidator
	debtToken := grc20reg.MustGet(debtAsset)
	err = debtToken.TransferFrom(caller, std.CurrentRealm(), debtAmount)
	if err != nil {
		return err
	}

	// Transfer collateral to liquidator
	collateralToken := grc20reg.MustGet(collateralAsset)
	err = collateralToken.Transfer(caller, collateralAmount)
	if err != nil {
		return err
	}

	// Update user position
	supply := userPosition.SuppliedAssets[collateralAsset]
	supply.Amount -= collateralAmount
	if supply.Amount == 0 {
		delete(userPosition.SuppliedAssets, collateralAsset)
	} else {
		userPosition.SuppliedAssets[collateralAsset] = supply
	}

	borrow := userPosition.BorrowedAssets[debtAsset]
	borrow.Amount -= debtAmount
	if borrow.Amount == 0 {
		delete(userPosition.BorrowedAssets, debtAsset)
	} else {
		userPosition.BorrowedAssets[debtAsset] = borrow
	}

	// Update health factor
	updateHealthFactor(userPosition)

	return nil
}

// Helper functions

func updatePoolIndices(pool *types.LendingPool) {
	currentBlock := std.GetHeight()
	if currentBlock == pool.LastUpdateBlock {
		return
	}

	utilizationRate := math.CalculateUtilizationRate(pool.TotalBorrow, pool.TotalSupply)
	params := interestRateParams[pool.AssetAddress]
	
	borrowRate := math.CalculateBorrowRate(params, utilizationRate)
	supplyRate := math.CalculateSupplyRate(borrowRate, utilizationRate)

	timeDelta := currentBlock - pool.LastUpdateBlock

	pool.LiquidityIndex = (pool.LiquidityIndex * math.CalculateCompoundedInterest(supplyRate, timeDelta)) / uint64(consts.PRECISION)
	pool.BorrowIndex = (pool.BorrowIndex * math.CalculateCompoundedInterest(borrowRate, timeDelta)) / uint64(consts.PRECISION)
	pool.LastUpdateBlock = currentBlock
}

func getOrCreatePosition(user std.Address) *types.UserPosition {
	position, exists := positions[user]
	if !exists {
		position = &types.UserPosition{
			Owner:           user,
			SuppliedAssets: make(map[std.Address]types.Supply),
			BorrowedAssets: make(map[std.Address]types.Borrow),
			HealthFactor:   math.MaxUint64,
			LastUpdateBlock: std.GetHeight(),
		}
		positions[user] = position
	}
	return position
}

func canBorrow(position *types.UserPosition, asset std.Address, amount uint64) bool {
	totalCollateralValue := calculateTotalCollateralValue(position)
	totalBorrowValue := calculateTotalBorrowValue(position)

	assetPrice, err := oracle.GetPrice(asset)
	if err != nil {
		return false
	}

	newBorrowValue := totalBorrowValue + (amount * assetPrice)
	maxBorrowValue := (totalCollateralValue * pools[asset].LTV) / 10000

	return newBorrowValue <= maxBorrowValue
}

func canWithdraw(position *types.UserPosition, asset std.Address, amount uint64) bool {
	supply, exists := position.SuppliedAssets[asset]
	if !exists || amount > supply.Amount {
		return false
	}

	if !supply.IsCollateral {
		return true
	}

	assetPrice, err := oracle.GetPrice(asset)
	if err != nil {
		return false
	}

	withdrawValue := amount * assetPrice
	totalCollateralValue := calculateTotalCollateralValue(position)
	totalBorrowValue := calculateTotalBorrowValue(position)

	remainingCollateral := totalCollateralValue - withdrawValue
	if remainingCollateral == 0 {
		return totalBorrowValue == 0
	}

	newHealthFactor := (remainingCollateral * pools[asset].LiquidationThreshold) / (totalBorrowValue * 10000)
	return newHealthFactor >= uint64(consts.PRECISION)
}

func calculateTotalCollateralValue(position *types.UserPosition) uint64 {
	total := uint64(0)
	for asset, supply := range position.SuppliedAssets {
		if !supply.IsCollateral {
			continue
		}
		price, err := oracle.GetPrice(asset)
		if err != nil {
			continue
		}
		total += supply.Amount * price
	}
	return total
}

func calculateTotalBorrowValue(position *types.UserPosition) uint64 {
	total := uint64(0)
	for asset, borrow := range position.BorrowedAssets {
		price, err := oracle.GetPrice(asset)
		if err != nil {
			continue
		}
		total += calculateDebtWithInterest(borrow.Amount, borrow.BorrowIndex, pools[asset].BorrowIndex) * price
	}
	return total
}

func calculateDebtWithInterest(amount, borrowIndexStart, borrowIndexCurrent uint64) uint64 {
	return (amount * borrowIndexCurrent) / borrowIndexStart
}

func updateHealthFactor(position *types.UserPosition) {
	totalCollateralValue := calculateTotalCollateralValue(position)
	totalBorrowValue := calculateTotalBorrowValue(position)

	if totalBorrowValue == 0 {
		position.HealthFactor = math.MaxUint64
		return
	}

	// Use the lowest liquidation threshold among all collateral assets
	minLiquidationThreshold := uint64(10000) // 100%
	for asset := range position.SuppliedAssets {
		if pools[asset].LiquidationThreshold < minLiquidationThreshold {
			minLiquidationThreshold = pools[asset].LiquidationThreshold
		}
	}

	position.HealthFactor = (totalCollateralValue * minLiquidationThreshold) / (totalBorrowValue * 10000)
} 