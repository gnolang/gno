package test

import (
	"std"
	"gno.land/p/demo/ufmt"
)

func Render(path string) string {
	banker := std.NewBanker(std.BankerTypeReadonly)
	coins := banker.GetCoins(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"))
	banker.TotalCoin("ugnot")

	return ufmt.Sprintf("my coins : %v", coins)
}
