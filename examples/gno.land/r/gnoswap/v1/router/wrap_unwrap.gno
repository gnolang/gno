package router

import (
	"std"

	"gno.land/r/demo/wugnot"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
)

var (
	errFailedToWrapZeroUgnot = "cannot wrap 0 ugnot"
	errFailedToWrapBelowMin  = "amount(%d) < minimum(%d)"
)

func wrap(ugnotAmount int64) {
	if ugnotAmount <= 0 {
		panic(addDetailToError(errWrapUnwrap, errFailedToWrapZeroUgnot))
	}

	if ugnotAmount < int64(consts.UGNOT_MIN_DEPOSIT_TO_WRAP) {
		panic(addDetailToError(
			errWugnotMinimum,
			ufmt.Sprintf(errFailedToWrapBelowMin, ugnotAmount, consts.UGNOT_MIN_DEPOSIT_TO_WRAP),
		))
	}

	// WRAP IT
	wugnotAddr := std.DerivePkgAddr(consts.WRAPPED_WUGNOT)
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(routerAddr, wugnotAddr, std.Coins{{"ugnot", int64(ugnotAmount)}})
	wugnot.Deposit(cross) // ROUTER HAS WUGNOT

	// SEND WUGNOT: ROUTER -> USER
	wugnot.Transfer(cross, std.PreviousRealm().Address(), ugnotAmount)
}

func unwrap(wugnotAmount int64) {
	if wugnotAmount == 0 {
		return
	}

	// SEND WUGNOT: USER -> ROUTER
	wugnot.TransferFrom(cross, std.PreviousRealm().Address(), routerAddr, wugnotAmount)

	// UNWRAP IT
	wugnot.Withdraw(cross, wugnotAmount)

	// SEND GNOT: ROUTER -> USER
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(routerAddr, std.PreviousRealm().Address(), std.Coins{{"ugnot", int64(wugnotAmount)}})
}
