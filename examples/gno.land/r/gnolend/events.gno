package gnolend

import (
	"std"
	
	"gno.land/p/demo/ufmt"
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
	EventLiquidatorKey      = "liquidator"
	EventBorrowerKey        = "borrower"
	EventLiquidatedAmtKey   = "liquidated_amount"
	EventCollateralAmtKey   = "collateral_amount"
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

func emitDeposit(marketId string, user std.Address, amount int64) {
	std.Emit(
		DepositEvent,
		EventMarketIDKey, marketId,
		EventUserKey, user.String(),
		EventAmountKey, ufmt.Sprintf("%d", amount),
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

func emitBorrow(marketId string, user std.Address, amount int64) {
	std.Emit(
		BorrowEvent,
		EventMarketIDKey, marketId,
		EventUserKey, user.String(),
		EventAmountKey, ufmt.Sprintf("%d", amount),
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