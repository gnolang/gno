package router

import (
	"std"

	"gno.land/r/demo/wugnot"

	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
)

func wrap(ugnotAmount uint64) {
	if ugnotAmount <= 0 {
		panic(addDetailToError(
			errWrapUnwrap,
			ufmt.Sprintf("wrap.gno__wrap() || cannot wrap 0 ugnot"),
		))
	}

	if ugnotAmount < consts.UGNOT_MIN_DEPOSIT_TO_WRAP {
		panic(addDetailToError(
			errWugnotMinimum,
			ufmt.Sprintf("wrap.gno__wrap() || amount(%d) < minimum(%d)", ugnotAmount, consts.UGNOT_MIN_DEPOSIT_TO_WRAP),
		))
	}

	// WRAP IT
	wugnotAddr := std.DerivePkgAddr(consts.WRAPPED_WUGNOT)
	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(consts.ROUTER_ADDR, wugnotAddr, std.Coins{{"ugnot", int64(ugnotAmount)}})
	wugnot.Deposit() // ROUTER HAS WUGNOT

	// SEND WUGNOT: ROUTER -> USER
	wugnot.Transfer(std.PrevRealm().Addr(), ugnotAmount)
}

func unwrap(wugnotAmount uint64) {
	if wugnotAmount == 0 {
		return
	}

	// SEND WUGNOT: USER -> ROUTER
	wugnot.TransferFrom(std.PrevRealm().Addr(), consts.ROUTER_ADDR, wugnotAmount)

	// UNWRAP IT
	wugnot.Withdraw(wugnotAmount)

	// SEND GNOT: ROUTER -> USER
	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(consts.ROUTER_ADDR, std.PrevRealm().Addr(), std.Coins{{"ugnot", int64(wugnotAmount)}})
}
