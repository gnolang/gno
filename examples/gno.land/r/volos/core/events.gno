package volos

import (
	"std"
	"strconv"
	"time"

	u256 "gno.land/p/gnoswap/uint256"
)

// Event names
const (
	// Market events
	CreateMarketEvent       = "CreateMarket"
	SupplyEvent             = "Supply"
	WithdrawEvent           = "Withdraw"
	BorrowEvent             = "Borrow"
	RepayEvent              = "Repay"
	LiquidateEvent          = "Liquidate"
	RegisterIRMEvent        = "RegisterIRM"
	AccrueInterestEvent     = "AccrueInterest"
	SupplyCollateralEvent   = "SupplyCollateral"
	WithdrawCollateralEvent = "WithdrawCollateral"
	FlashLoanEvent          = "FlashLoan"
	SetFeeEvent             = "SetFee"

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
    AuthorizationSetEvent = "authorization_set"
)

// Event keys
const (
	// Common keys
	EventMarketIDKey = "market_id"
	EventAmountKey   = "amount"
	EventUserKey     = "user"
	EventOnBehalfKey = "on_behalf"
	EventReceiverKey = "receiver"

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
    EventAuthorizerKey    = "authorizer"
    EventAuthorizedKey    = "authorized"
    EventIsAuthorizedKey  = "is_authorized"
    EventTokenKey = "token"
    EventTimestampKey     = "currentTimestamp"
)

// Event emission helper functions
func emitCreateMarket(marketId string, loanToken string, collateralToken string) {
	std.Emit(
		CreateMarketEvent,
		EventMarketIDKey, marketId,
		EventLoanTokenKey, loanToken,
		EventCollateralTokenKey, collateralToken,
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitSupply(marketId string, caller std.Address, onBehalf std.Address, assets, shares *u256.Uint) {
	std.Emit(
		SupplyEvent,
		EventMarketIDKey, marketId,
		EventUserKey, caller.String(),
		EventOnBehalfKey, onBehalf.String(),
		EventAmountKey, assets.ToString(),
		EventSharesKey, shares.ToString(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitWithdraw(marketId string, caller std.Address, onBehalf std.Address, receiver std.Address, assets, shares *u256.Uint) {
	std.Emit(
		WithdrawEvent,
		EventMarketIDKey, marketId,
		EventUserKey, caller.String(),
		EventOnBehalfKey, onBehalf.String(),
		EventReceiverKey, receiver.String(),
		EventAmountKey, assets.ToString(),
		EventSharesKey, shares.ToString(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitBorrow(marketId string, caller std.Address, onBehalf std.Address, receiver std.Address, assets, shares *u256.Uint) {
	std.Emit(
		BorrowEvent,
		EventMarketIDKey, marketId,
		EventUserKey, caller.String(),
		EventOnBehalfKey, onBehalf.String(),
		EventReceiverKey, receiver.String(),
		EventAmountKey, assets.ToString(),
		EventSharesKey, shares.ToString(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitRepay(marketId string, caller std.Address, onBehalf std.Address, assets, shares *u256.Uint) {
	std.Emit(
		RepayEvent,
		EventMarketIDKey, marketId,
		EventUserKey, caller.String(),
		EventOnBehalfKey, onBehalf.String(),
		EventAmountKey, assets.ToString(),
		EventSharesKey, shares.ToString(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
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
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitAccrueInterest(marketId string, borrowRate, interest *u256.Uint) {
	std.Emit(
		AccrueInterestEvent,
		EventMarketIDKey, marketId,
		EventBorrowRateKey, borrowRate.ToString(),
		EventInterestKey, interest.ToString(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitRegisterIRM(pkgPath, name string) {
	std.Emit(
		RegisterIRMEvent,
		EventPkgPathKey, pkgPath,
		EventNameKey, name,
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitSupplyCollateral(marketId string, caller std.Address, onBehalf std.Address, amount uint64) {
	std.Emit(
		SupplyCollateralEvent,
		EventMarketIDKey, marketId,
		EventUserKey, caller.String(),
		EventOnBehalfKey, onBehalf.String(),
		EventAmountKey, u256.NewUint(amount).ToString(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitWithdrawCollateral(marketId string, caller std.Address, onBehalf std.Address, receiver std.Address, amount uint64) {
	std.Emit(
		WithdrawCollateralEvent,
		EventMarketIDKey, marketId,
		EventUserKey, caller.String(),
		EventOnBehalfKey, onBehalf.String(),
		EventReceiverKey, receiver.String(),
		EventAmountKey, u256.NewUint(amount).ToString(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// emitAuthorizationSet emits an event when authorization is set or revoked
func emitAuthorizationSet(authorizer std.Address, authorized std.Address, isAuthorized bool) {
	std.Emit(
		AuthorizationSetEvent,
		EventAuthorizerKey, authorizer.String(),
		EventAuthorizedKey, authorized.String(),
		EventIsAuthorizedKey, strconv.FormatBool(isAuthorized),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// emitFlashLoan emits an event when a flash loan occurs
func emitFlashLoan(caller std.Address, token string, assets int64) {
	std.Emit(
		FlashLoanEvent,
		EventUserKey, caller.String(),
		EventTokenKey, token,
		EventAmountKey, u256.NewUint(uint64(assets)).ToString(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// emitSetFee emits an event when a market fee is set (following Morpho pattern)
func emitSetFee(marketId string, fee string) {
	std.Emit(
		SetFeeEvent,
		EventMarketIDKey, marketId,
		EventFeeKey, fee,
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}
