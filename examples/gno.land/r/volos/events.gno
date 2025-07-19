package volos

import (
	"std"

	u256 "gno.land/p/gnoswap/uint256"
)

// Event names
const (
	// Market events
	CreateMarketEvent       = "CreateMarket"
	DepositEvent            = "Deposit"
	WithdrawEvent           = "Withdraw"
	BorrowEvent             = "Borrow"
	RepayEvent              = "Repay"
	LiquidateEvent          = "Liquidate"
	RegisterIRMEvent        = "RegisterIRM"
	AccrueInterestEvent     = "AccrueInterest"
	SupplyCollateralEvent   = "SupplyCollateral"
	WithdrawCollateralEvent = "WithdrawCollateral"

	// Event names
	EventAccrueInterest     = "accrue_interest"
	EventCreateMarket       = "create_market"
	EventSupply             = "supply"
	EventWithdraw           = "withdraw"
	EventBorrow             = "borrow"
	EventRepay              = "repay"
	EventLiquidate          = "liquidate"
	EventSupplyCollateral   = "supply_collateral"
	EventWithdrawCollateral = "withdraw_collateral"
)

// Event keys
const (
	// Common keys
	EventMarketIDKey = "market_id"
	EventAmountKey   = "amount"
	EventUserKey     = "user"

	// Market keys
	EventLoanTokenKey       = "loan_token"
	EventCollateralTokenKey = "collateral_token"

	// Liquidation keys
	EventLiquidatorKey    = "liquidator"
	EventBorrowerKey      = "borrower"
	EventLiquidatedAmtKey = "liquidated_amount"
	EventCollateralAmtKey = "collateral_amount"

	// Interest keys
	EventBorrowRateKey = "borrow_rate"
	EventInterestKey   = "interest"

	// Event keys
	EventMarketIdKey = "market_id"
	EventSharesKey   = "shares"
	EventSeizedKey   = "seized"
	EventFeeKey      = "fee"
	EventPkgPathKey  = "pkgpath"
	EventNameKey     = "name"
)

// Event emission helper functions
func emitCreateMarket(marketId string, loanToken string, collateralToken string) {
	std.Emit(
		CreateMarketEvent,
		EventMarketIDKey, marketId,
		EventLoanTokenKey, loanToken,
		EventCollateralTokenKey, collateralToken,
	)
}

func emitDeposit(marketId string, user std.Address, amount *u256.Uint) {
	std.Emit(
		DepositEvent,
		EventMarketIDKey, marketId,
		EventUserKey, user.String(),
		EventAmountKey, amount.ToString(),
	)
}

func emitWithdraw(marketId string, caller std.Address, amount *u256.Uint) {
	std.Emit(
		WithdrawEvent,
		EventMarketIDKey, marketId,
		EventUserKey, caller.String(),
		EventAmountKey, amount.ToString(),
	)
}

func emitBorrow(marketId string, user std.Address, amount *u256.Uint) {
	std.Emit(
		BorrowEvent,
		EventMarketIDKey, marketId,
		EventUserKey, user.String(),
		EventAmountKey, amount.ToString(),
	)
}

func emitRepay(marketId string, caller std.Address, assets, shares *u256.Uint) {
	std.Emit(
		RepayEvent,
		EventMarketIDKey, marketId,
		EventUserKey, caller.String(),
		EventAmountKey, assets.ToString(),
		EventSharesKey, shares.ToString(),
	)
}

func emitLiquidate(marketId string, caller std.Address, borrower std.Address, repaidAssets, repaidShares, seizedAssets *u256.Uint) {
	std.Emit(
		EventLiquidate,
		EventMarketIdKey, marketId,
		EventUserKey, caller.String(),
		EventBorrowerKey, borrower.String(),
		EventAmountKey, repaidAssets.ToString(),
		EventSharesKey, repaidShares.ToString(),
		EventSeizedKey, seizedAssets.ToString(),
	)
}

func emitAccrueInterest(marketId string, borrowRate, interest *u256.Uint) {
	std.Emit(
		AccrueInterestEvent,
		EventMarketIDKey, marketId,
		EventBorrowRateKey, borrowRate.ToString(),
		EventInterestKey, interest.ToString(),
	)
}

func emitRegisterIRM(pkgPath, name string) {
	std.Emit(
		RegisterIRMEvent,
		EventPkgPathKey, pkgPath,
		EventNameKey, name,
	)
}

func emitSupplyCollateral(marketId string, caller std.Address, amount uint64) {
	std.Emit(
		SupplyCollateralEvent,
		EventMarketIDKey, marketId,
		EventUserKey, caller.String(),
		EventAmountKey, u256.NewUint(amount).ToString(),
	)
}

func emitWithdrawCollateral(marketId string, caller std.Address, amount uint64) {
	std.Emit(
		WithdrawCollateralEvent,
		EventMarketIDKey, marketId,
		EventUserKey, caller.String(),
		EventAmountKey, u256.NewUint(amount).ToString(),
	)
}
