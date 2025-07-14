package position

import (
	"std"
	"strconv"

	"gno.land/r/demo/wugnot"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
)

// wrap wraps the specified amount of the native token `ugnot` into the wrapped token `wugnot`.
//
// Parameters:
//   - ugnotAmount (uint64): The amount of `ugnot` tokens to wrap into `wugnot`.
//   - to (std.Address): The recipient's address to receive the wrapped tokens.
//
// Returns:
//   - error: An error if the `ugnot` amount is zero, below the minimum wrapping threshold, or any other issue occurs.
//
// Example:
//
//	wrap(1000, userAddress)
//	- Wraps 1000 UGNOT into WUGNOT and transfers the WUGNOT to `userAddress`.
//
// Errors:
//   - Returns an error if `ugnotAmount` is zero or less than the minimum deposit threshold.
func wrap(ugnotAmount int64, to std.Address) error {
	if ugnotAmount == 0 || ugnotAmount < int64(consts.UGNOT_MIN_DEPOSIT_TO_WRAP) {
		return ufmt.Errorf("amount(%d) < minimum(%d)", ugnotAmount, consts.UGNOT_MIN_DEPOSIT_TO_WRAP)
	}

	wugnotAddr := std.DerivePkgAddr(consts.WRAPPED_WUGNOT)
	transferUGNOT(positionAddr, wugnotAddr, ugnotAmount)

	wugnot.Deposit(cross)                   // POSITION HAS WUGNOT
	wugnot.Transfer(cross, to, ugnotAmount) // SEND WUGNOT: POSITION -> USER

	return nil
}

// unwrap converts a specified amount of `WUGNOT` tokens into `UGNOT` tokens
// and transfers the resulting `UGNOT` back to the specified recipient address.
//
// Parameters:
// - `wugnotAmount`: The amount of `WUGNOT` tokens to unwrap (uint64).
// - `to`: The recipient's address (std.Address) to receive the unwrapped `UGNOT`.
//
// Example:
// unwrap(100, userAddress)
// - Converts 100 WUGNOT into UGNOT and sends the resulting UGNOT to `userAddress`.
func unwrap(wugnotAmount int64, to std.Address) error {
	if wugnotAmount <= 0 {
		return ufmt.Errorf("amount(%d) is zero or negative", wugnotAmount)
	}

	wugnot.TransferFrom(cross, to, positionAddr, wugnotAmount) // SEND WUGNOT: USER -> POSITION
	wugnot.Withdraw(cross, wugnotAmount)                       // POSITION HAS UGNOT
	transferUGNOT(positionAddr, to, wugnotAmount)              // SEND UGNOT: POSITION -> USER
	return nil
}

// transferUGNOT transfers a specified amount of `UGNOT` tokens from one address to another.
// The function ensures that no transaction occurs if the transfer amount is zero.
// It uses the `std.BankerTypeRealmSend` banker type to facilitate the transfer.
//
// Parameters:
// - `from`: The sender's address (std.Address).
// - `to`: The recipient's address (std.Address).
// - `amount`: The amount of UGNOT tokens to transfer (uint64).
//
// Example:
// transferUGNOT(sender, receiver, 100) // Transfers 100 UGNOT from `sender` to `receiver`.
func transferUGNOT(from, to std.Address, amount int64) {
	if amount < 0 {
		panic(ufmt.Sprintf("amount(%d) is negative", amount))
	}
	if amount == 0 {
		return
	}

	banker := std.NewBanker(std.BankerTypeRealmSend)
	fromBalance := banker.GetCoins(from).AmountOf(consts.UGNOT)
	if fromBalance < int64(amount) {
		panic(newErrorWithDetail(
			errInsufficientUGNOT,
			ufmt.Sprintf("from(%s) balance(%d) is less than amount(%d)", from, fromBalance, amount)))
	}
	banker.SendCoins(from, to, std.Coins{
		{Denom: consts.UGNOT, Amount: int64(amount)},
	})
}

// refundUGNOT refunds a specified amount of `UGNOT` tokens to the provided address.
// This function uses `transferUGNOT` to perform the transfer from the contract's position address
// (`POSITION_ADDR`) to the recipient.
//
// Parameters:
// - `to`: The recipient's address (std.Address) who will receive the refund.
// - `amount`: The amount of `UGNOT` tokens to refund (uint64).
func refundUGNOT(to std.Address, amount int64) {
	transferUGNOT(positionAddr, to, amount)
}

// isNative checks whether the given token is a native token.
func isNative(token string) bool {
	return token == consts.GNOT
}

// isWrappedToken checks whether the tokenPath is wrapped token
func isWrappedToken(tokenPath string) bool {
	return tokenPath == consts.WRAPPED_WUGNOT
}

// safeWrapNativeToken safely wraps the native token `ugnot` into the wrapped token `wugnot` for a user.
//
// Parameters:
//   - amountDesired: The desired amount of `ugnot` to be wrapped, provided as a string.
//   - userAddress: The address of the user initiating the wrapping process.
//
// Returns:
//   - uint64: The amount of `ugnot` that was successfully wrapped into `wugnot`.
//
// Panics:
//   - If the sent `ugnot` amount is zero.
//   - If `amountDesired` cannot be parsed into a valid uint64 value.
//   - If the sent `ugnot` amount is less than `amountDesired`.
//   - If the `wrap` function fails to wrap the tokens.
//   - If there is a mismatch between the expected wrapped token amount and the user's balance after wrapping.
func safeWrapNativeToken(amountDesired string, userAddress std.Address) int64 {
	amount, err := strconv.ParseInt(amountDesired, 10, 64)
	if err != nil {
		panic(newErrorWithDetail(errWrapUnwrap, err.Error()))
	}

	// if amountDesired is zero, return 0
	if amount == 0 {
		return 0
	}

	beforeWugnotBalance := wugnot.BalanceOf(userAddress)
	sentNative := std.OriginSend()
	sentUgnotAmount := sentNative.AmountOf(consts.UGNOT)

	if sentUgnotAmount <= 0 {
		panic(newErrorWithDetail(errZeroUGNOT, "amount of ugnot is zero"))
	}

	if sentUgnotAmount < amount {
		panic(newErrorWithDetail(errInsufficientUGNOT, "amount of ugnot is less than desired amount"))
	}

	if sentUgnotAmount > amount {
		exceed := sentUgnotAmount - amount
		refundUGNOT(userAddress, exceed)

		sentUgnotAmount = amount
	}

	if err = wrap(sentUgnotAmount, userAddress); err != nil {
		panic(newErrorWithDetail(errWugnotMinimum, err.Error()))
	}

	afterWugnotBalance := wugnot.BalanceOf(userAddress)
	diff := afterWugnotBalance - beforeWugnotBalance

	if diff != sentUgnotAmount {
		panic(newErrorWithDetail(
			errWrapUnwrap,
			ufmt.Sprintf("amount of ugnot (%d) is not equal to amount of wugnot. (diff: %d)", sentUgnotAmount, diff),
		))
	}
	return sentUgnotAmount
}

func handleUnwrap(pToken0, pToken1 string, unwrapResult bool, userOldWugnotBalance int64, to std.Address) {
	if (pToken0 == consts.WRAPPED_WUGNOT || pToken1 == consts.WRAPPED_WUGNOT) && unwrapResult {
		userNewWugnotBalance := wugnot.BalanceOf(to)
		leftOver := userNewWugnotBalance - userOldWugnotBalance
		if leftOver > 0 {
			if err := unwrap(leftOver, to); err != nil {
				panic(newErrorWithDetail(errWrapUnwrap, err.Error()))
			}
		}
	}
}
