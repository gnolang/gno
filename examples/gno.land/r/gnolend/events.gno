package gnolend

import (
	"std"

	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/demo/uint256"
)

// Event names
const (
	// Market events
	CreateMarketEvent = "CreateMarket"
	DepositEvent      = "Deposit"
	WithdrawEvent     = "Withdraw"
	BorrowEvent       = "Borrow"
	RepayEvent        = "Repay"
	LiquidateEvent    = "Liquidate"
	RegisterIRMEvent  = "RegisterIRM"
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
		EventAmountKey, amount.String(),
	)
}

func emitWithdraw(marketId string, user std.Address, amount int64) {
	std.Emit(
		WithdrawEvent,
		EventMarketIDKey, marketId,
		EventUserKey, user.String(),
		EventAmountKey, ufmt.Sprintf("%d", amount),
	)
}

func emitBorrow(marketId string, user std.Address, amount *u256.Uint) {
	std.Emit(
		BorrowEvent,
		EventMarketIDKey, marketId,
		EventUserKey, user.String(),
		EventAmountKey, amount.String(),
	)
}

func emitRepay(marketId string, user std.Address, amount int64) {
	std.Emit(
		RepayEvent,
		EventMarketIDKey, marketId,
		EventUserKey, user.String(),
		EventAmountKey, ufmt.Sprintf("%d", amount),
	)
}

func emitLiquidate(marketId string, liquidator std.Address, borrower std.Address, liquidatedAmt int64, collateralAmt int64) {
	std.Emit(
		LiquidateEvent,
		EventMarketIDKey, marketId,
		EventLiquidatorKey, liquidator.String(),
		EventBorrowerKey, borrower.String(),
		EventLiquidatedAmtKey, ufmt.Sprintf("%d", liquidatedAmt),
		EventCollateralAmtKey, ufmt.Sprintf("%d", collateralAmt),
	)
}

func emitAccrueInterest(marketId string, borrowRate, interest *u256.Uint) {
	std.Emit(
		"AccrueInterest",
		"market_id", marketId,
		"borrow_rate", borrowRate.String(),
		"interest", interest.String(),
	)
}

func emitRegisterIRM(pkgPath, name string) {
	std.Emit(
		RegisterIRMEvent,
		"pkgpath", pkgPath,
		"name", name,
	)
}

func emitSupplyCollateral(marketId string, caller std.Address, amount uint64) {
	std.Emit(
		"supply_collateral",
		"market_id", marketId,
		"caller", caller.String(),
		"amount", ufmt.Sprintf("%d", amount),
	)
}

func emitWithdrawCollateral(marketId string, caller std.Address, amount uint64) {
	std.Emit(
		"withdraw_collateral",
		"market_id", marketId,
		"caller", caller.String(),
		"amount", ufmt.Sprintf("%d", amount),
	)
}
