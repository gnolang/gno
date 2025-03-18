package staker

import (
	"std"

	"gno.land/r/demo/wugnot"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
)

// wrap converts `ugnot` tokens into `wugnot` tokens.
//
// Parameters:
// - ugnotAmount (uint64): The amount of `ugnot` to wrap.
//
// Panics:
// - If `ugnotAmount` is less than or equal to 0.
// - If `ugnotAmount` is less than the minimum deposit required (`consts.UGNOT_MIN_DEPOSIT_TO_WRAP`).
func wrap(ugnotAmount uint64) {
	if ugnotAmount <= 0 {
		panic(addDetailToError(
			errWrapUnwrap,
			"cannot wrap 0 ugnot",
		))
	}

	if ugnotAmount < consts.UGNOT_MIN_DEPOSIT_TO_WRAP {
		panic(addDetailToError(
			errWugnotMinimum,
			ufmt.Sprintf("amount(%d) < minimum(%d)", ugnotAmount, consts.UGNOT_MIN_DEPOSIT_TO_WRAP),
		))
	}

	// WRAP IT
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(stakerAddr, consts.WUGNOT_ADDR, std.Coins{{Denom: "ugnot", Amount: int64(ugnotAmount)}})
	wugnot.Deposit() // STAKER HAS WUGNOT

	// SEND WUGNOT: STAKER -> USER
	wugnot.Transfer(std.PreviousRealm().Address(), ugnotAmount)
}

// unwrap converts `wugnot` tokens back into `ugnot` tokens.
//
// Parameters:
// - wugnotAmount (uint64): The amount of `wugnot` to unwrap.
//
// Note:
// - If `wugnotAmount` is 0, the function simply returns without executing further steps.
func unwrap(wugnotAmount uint64) {
	if wugnotAmount == 0 {
		return
	}

	// SEND WUGNOT: USER -> STAKER
	wugnot.TransferFrom(std.PreviousRealm().Address(), stakerAddr, wugnotAmount)

	// UNWRAP IT
	wugnot.Withdraw(wugnotAmount)

	// SEND GNOT: STAKER -> USER
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(stakerAddr, std.PreviousRealm().Address(), std.Coins{{Denom: "ugnot", Amount: int64(wugnotAmount)}})
}
