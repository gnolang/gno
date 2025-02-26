package gnolend

import (
	"errors"
)

var (
	ErrMarketNotCreated      = errors.New("market not created")
	ErrMarketAlreadyCreated  = errors.New("market already created")
	ErrTokenPairNotInGnoswap = errors.New("token pair not in gnoswap pools")
	ErrZeroAddress           = errors.New("zero address")
	ErrSameToken             = errors.New("loan and collateral tokens cannot be the same")

	// Supply/Withdraw errors
	ErrInconsistentAmount  = errors.New("must specify either assets or shares, not both")
	ErrInsufficientBalance = errors.New("insufficient token balance")
	ErrInsufficientShares  = errors.New("insufficient shares")
	ErrDivisionByZero      = errors.New("division by zero")
)
