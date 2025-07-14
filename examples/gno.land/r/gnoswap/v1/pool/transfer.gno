package pool

import (
	"std"

	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/common"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

// safeTransfer performs a token transfer out of the pool while ensuring
// the pool has sufficient balance and updating internal accounting.
// This function is typically used during swaps and liquidity removals.
//
// Important requirements:
//   - The amount must be negative (representing an outflow from the pool)
//   - The pool must have sufficient balance for the transfer
//   - The transfer amount must fit within uint64 range
//
// Parameters:
//   - to: destination address for the transfer
//   - tokenPath: path identifier of the token to transfer
//   - amount: amount to transfer (must be negative)
//   - isToken0: true if transferring token0, false for token1
//
// The function will:
//  1. Validate the amount is negative
//  2. Check pool has sufficient balance
//  3. Execute the transfer
//  4. Update pool's internal balance
//
// Panics if any validation fails or if the transfer fails
func (p *Pool) safeTransfer(
	to std.Address,
	tokenPath string,
	amount *i256.Int,
	isToken0 bool,
) {
	if amount.Gt(i256.Zero()) {
		panic(ufmt.Sprintf(
			"%v. got: %s", errMustBeNegative, amount.ToString(),
		))
	}

	absAmount := amount.Abs()

	token0 := p.BalanceToken0()
	token1 := p.BalanceToken1()

	if err := validatePoolBalance(token0, token1, absAmount, isToken0); err != nil {
		panic(err)
	}
	amountInt64 := safeConvertToInt64(absAmount)

	checkTransferError(common.Transfer(cross, tokenPath, to, amountInt64))

	newBalance, err := updatePoolBalance(token0, token1, absAmount, isToken0)
	if err != nil {
		panic(err)
	}

	if isToken0 {
		p.balances.token0 = newBalance
	} else {
		p.balances.token1 = newBalance
	}
}

// safeTransferFrom securely transfers tokens into the pool while ensuring balance consistency.
//
// This function performs the following steps:
// 1. Validates and converts the transfer amount to `uint64` using `safeConvertToUint64`.
// 2. Executes the token transfer using `TransferFrom` via the token teller contract.
// 3. Verifies that the destination balance reflects the correct amount after transfer.
// 4. Updates the pool's internal balances (`token0` or `token1`) and validates the updated state.
//
// Parameters:
// - from (std.Address): Source address for the token transfer.
// - to (std.Address): Destination address, typically the pool address.
// - tokenPath (string): Path identifier for the token being transferred.
// - amount (*u256.Uint): The amount of tokens to transfer (must be a positive value).
// - isToken0 (bool): A flag indicating whether the token being transferred is token0 (`true`) or token1 (`false`).
//
// Panics:
// - If the `amount` exceeds the uint64 range during conversion.
// - If the token transfer (`TransferFrom`) fails.
// - If the destination balance after the transfer does not match the expected amount.
// - If the pool's internal balances (`token0` or `token1`) overflow or become inconsistent.
//
// Notes:
// - The function assumes that the sender (`from`) has approved the pool to spend the specified tokens.
// - The balance consistency check ensures that no tokens are lost or double-counted during the transfer.
// - Pool balance updates are performed atomically to ensure internal consistency.
//
// Example:
// p.safeTransferFrom(
//
//	sender, poolAddress, "path/to/token0", u256.MustFromDecimal("1000"), true
//
// )
func (p *Pool) safeTransferFrom(
	from, to std.Address,
	tokenPath string,
	amount *u256.Uint,
	isToken0 bool,
) {
	amountInt64 := safeConvertToInt64(amount)

	token := common.GetToken(tokenPath)
	beforeBalance := token.BalanceOf(to)

	checkTransferError(common.TransferFrom(cross, tokenPath, from, to, amountInt64))

	afterBalance := token.BalanceOf(to)
	if (beforeBalance + amountInt64) != afterBalance {
		panic(ufmt.Sprintf(
			"%v. beforeBalance(%d) + amount(%d) != afterBalance(%d)",
			errTransferFailed, beforeBalance, amountInt64, afterBalance,
		))
	}

	// update pool balances
	if isToken0 {
		beforeToken0 := p.balances.token0.Clone()
		p.balances.token0 = new(u256.Uint).Add(p.balances.token0, amount)
		if p.balances.token0.Lt(beforeToken0) {
			panic(ufmt.Sprintf(
				"%v. token0(%s) < beforeToken0(%s)",
				errBalanceUpdateFailed, p.balances.token0.ToString(), beforeToken0.ToString(),
			))
		}
	} else {
		beforeToken1 := p.balances.token1.Clone()
		p.balances.token1 = new(u256.Uint).Add(p.balances.token1, amount)
		if p.balances.token1.Lt(beforeToken1) {
			panic(ufmt.Sprintf(
				"%v. token1(%s) < beforeToken1(%s)",
				errBalanceUpdateFailed, p.balances.token1.ToString(), beforeToken1.ToString(),
			))
		}
	}
}

// validatePoolBalance checks if the pool has sufficient balance of either token0 and token1
// before proceeding with a transfer. This prevents the pool won't go into a negative balance.
func validatePoolBalance(token0, token1, amount *u256.Uint, isToken0 bool) error {
	if token0 == nil || token1 == nil || amount == nil {
		return ufmt.Errorf(
			"%v. token0(%s) or token1(%s) or amount(%s) is nil",
			errTransferFailed, token0.ToString(), token1.ToString(), amount.ToString(),
		)
	}

	if isToken0 {
		if token0.Lt(amount) {
			return ufmt.Errorf(
				"%v. token0(%s) >= amount(%s)",
				errTransferFailed, token0.ToString(), amount.ToString(),
			)
		}
		return nil
	}
	if token1.Lt(amount) {
		return ufmt.Errorf(
			"%v. token1(%s) >= amount(%s)",
			errTransferFailed, token1.ToString(), amount.ToString(),
		)
	}
	return nil
}

// updatePoolBalance calculates the new balance after a transfer and validate.
// It ensures the resulting balance won't be negative or overflow.
func updatePoolBalance(
	token0, token1, amount *u256.Uint,
	isToken0 bool,
) (*u256.Uint, error) {
	var overflow bool
	var newBalance *u256.Uint

	if isToken0 {
		newBalance, overflow = new(u256.Uint).SubOverflow(token0, amount)
		if isBalanceOverflowOrNegative(overflow, newBalance) {
			return nil, ufmt.Errorf(
				"%v. cannot decrease, token0(%s) - amount(%s)",
				errBalanceUpdateFailed, token0.ToString(), amount.ToString(),
			)
		}
		return newBalance, nil
	}

	newBalance, overflow = new(u256.Uint).SubOverflow(token1, amount)
	if isBalanceOverflowOrNegative(overflow, newBalance) {
		return nil, ufmt.Errorf(
			"%v. cannot decrease, token1(%s) - amount(%s)",
			errBalanceUpdateFailed, token1.ToString(), amount.ToString(),
		)
	}
	return newBalance, nil
}

// isBalanceOverflowOrNegative checks if the balance calculation resulted in an overflow or negative value.
func isBalanceOverflowOrNegative(overflow bool, newBalance *u256.Uint) bool {
	return overflow || newBalance.Lt(u256.Zero())
}
