package tokenaccess

import (
	"errors"
	"std"

	"gno.land/p/demo/grc/grc20"
)

var (
	ErrNoCaller              = errors.New("caller isn't identified")
	ErrInvalidCaller         = errors.New("caller isn't a valid address")
	ErrInsufficientBalance   = errors.New("access denied: insufficient token balance")
	ErrInsufficientAllowance = errors.New("access denied: insufficient allowance")
)

// RewardAndApprove mints rewardAmt tokens to owner and increases target's allowance by allowanceIncrement.
func RewardAndApprove(
	owner std.Address,
	rewardAmt int64,
	token *grc20.Token,
	ledger *grc20.PrivateLedger,
	target std.Address,
	allowanceIncrement int64,
) error {
	if err := checkCaller(target); err != nil {
		return err
	}
	currentAllowance := token.Allowance(owner, target)
	newAllowance := currentAllowance + allowanceIncrement

	if err := ledger.Mint(owner, rewardAmt); err != nil {
		return errors.New("Token mint failed: " + err.Error())
	}

	if err := ledger.Approve(owner, target, newAllowance); err != nil {
		return errors.New("Token approval failed: " + err.Error())
	}
	return nil
}

// Spend transfers amount tokens from 'from' to 'to', respecting allowance.
func Spend(
	token *grc20.Token,
	ledger *grc20.PrivateLedger,
	from std.Address,
	to std.Address,
	amount int64,
	caller std.Address,
) error {
	if err := checkCaller(caller); err != nil {
		return err
	}
	err := ledger.TransferFrom(from, caller, to, amount)
	if err != nil {
		return errors.New("token transfer failed: " + err.Error())
	}
	return nil
}

// HasAccess returns true if the caller meets both balance and allowance requirements.
func HasAccess(
	token *grc20.Token,
	owner std.Address,
	caller std.Address,
	minBalance int64,
	minAllowance int64,
) bool {
	assertCaller(caller)

	return token.BalanceOf(caller) >= minBalance &&
		token.Allowance(owner, caller) >= minAllowance
}


// requireMinimumBalance ensures a caller has at least amount of tokens.
func requireMinimumBalance(token *grc20.Token, amount int64, caller std.Address) error {
	if err := checkCaller(caller); err != nil {
		return err
	}
	balance := token.BalanceOf(caller)
	if balance < amount {
		return ErrInsufficientBalance
	}
	return nil
}

// requireMinimumAllowance ensures a caller has an approved allowance from the owner.
func requireMinimumAllowance(
	token *grc20.Token,
	owner std.Address,
	amount int64,
	caller std.Address,
) error {
	if err := checkCaller(caller); err != nil {
		return err
	}
	allowance := token.Allowance(owner, caller)
	if allowance < amount {
		return ErrInsufficientAllowance
	}
	return nil
}

// checkCaller checks if the caller is a valid address.
func checkCaller(caller std.Address) error {
	if caller == "" {
		return ErrNoCaller
	}
	if !caller.IsValid() {
		return ErrInvalidCaller
	}
	return nil
}

// assertCaller checks if the caller is a valid address and panics if not.
func assertCaller(caller std.Address) {
	if err := checkCaller(caller); err != nil {
		panic(err)
	}
}
