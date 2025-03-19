package staker

import (
	"std"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/gnft"

	pn "gno.land/r/gnoswap/v1/position"
)

// MintAndStake mints LP tokens and stakes them in a single transaction.
// Returns positionId, liquidity, amount0, amount1, poolPath
// ref: https://docs.gnoswap.io/contracts/staker/mint_stake.gno#mintandstake
func MintAndStake(
	token0 string,
	token1 string,
	fee uint32,
	tickLower int32,
	tickUpper int32,
	amount0Desired string, // *u256.Uint
	amount1Desired string, // *u256.Uint
	amount0Min string, // *u256.Uint
	amount1Min string, // *u256.Uint
	deadline int64,
) (uint64, string, string, string, string) {
	// if one click native
	if token0 == consts.GNOT || token1 == consts.GNOT {
		// check sent ugnot
		sent := std.OriginSend()
		ugnotSent := uint64(sent.AmountOf("ugnot"))

		// not enough ugnot sent
		if ugnotSent < consts.UGNOT_MIN_DEPOSIT_TO_WRAP {
			panic(addDetailToError(
				errWugnotMinimum,
				ufmt.Sprintf("mint_stake.gno__MintAndStake() || too less ugnot sent(%d), minimum:%d", ugnotSent, consts.UGNOT_MIN_DEPOSIT_TO_WRAP),
			))
		}

		// send it over to position to wrap
		banker := std.NewBanker(std.BankerTypeRealmSend)
		banker.SendCoins(stakerAddr, positionAddr, std.Coins{{Denom: "ugnot", Amount: int64(ugnotSent)}})
	}

	positionId, liquidity, amount0, amount1 := pn.Mint(
		token0,
		token1,
		fee,
		tickLower,
		tickUpper,
		amount0Desired,
		amount1Desired,
		amount0Min,
		amount1Min,
		deadline,
		stakerAddr,
		std.PreviousRealm().Address(),
		// TODO: need to add referrer address here?
		"",
	)

	// at this point, staker has minted token
	toTid := grc721.TokenID(ufmt.Sprintf("%d", positionId))
	gnft.SetTokenURIByImageURI(toTid)

	poolPath, _, _ := StakeToken(positionId, "") // poolPath, stakedAmount0, stakedAmount1

	return positionId, liquidity, amount0, amount1, poolPath
}
