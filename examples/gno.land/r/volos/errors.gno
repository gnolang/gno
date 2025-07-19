package volos

import (
	"errors"
)

var (
	// Market errors
	ErrMarketNotCreated      = errors.New("market not created")
	ErrMarketAlreadyCreated  = errors.New("market already created")
	ErrTokenPairNotInGnoswap = errors.New("token pair not in gnoswap pools")
	ErrZeroAddress           = errors.New("zero address")
	ErrSameToken             = errors.New("loan and collateral tokens cannot be the same")
	ErrIRMNotEnabled         = errors.New("IRM not enabled")
	ErrLLTVNotEnabled        = errors.New("LLTV not enabled")
	ErrMaxFeeExceeded        = errors.New("max fee exceeded")
	ErrNotOwner              = errors.New("not owner")
	ErrAlreadySet            = errors.New("already set")

	// Supply/Withdraw errors
	ErrInconsistentAmount    = errors.New("must specify either assets or shares, not both")
	ErrInsufficientBalance   = errors.New("insufficient token balance")
	ErrInsufficientShares    = errors.New("insufficient shares")
	ErrInsufficientLiquidity = errors.New("insufficient liquidity in market")
	ErrDivisionByZero        = errors.New("division by zero")

	// Borrow errors
	ErrInsufficientCollateral = errors.New("insufficient collateral for borrow")
	ErrExceedsLTV             = errors.New("borrow would exceed maximum LTV")
	ErrZeroBorrow             = errors.New("borrow amount must be greater than zero")
	ErrNoCollateral           = errors.New("must deposit collateral before borrowing")

	// Liquidation errors
	ErrHealthyPosition = errors.New("healthy position")

	// Oracle errors
	ErrPriceNotAvailable = errors.New("price not available from pool")

	// Transfer errors
	ErrInsufficientAllowance = errors.New("insufficient allowance")
	ErrTransferFailed        = errors.New("transfer failed")
	ErrBalanceVerificationFailed = errors.New("balance verification failed")
)
