package wallet

import "std"

func Receive(cur realm) {
	std.AssertOriginCall()
}

func GetBalance(cur realm) std.Coins {
	banker := std.NewBanker(std.BankerTypeReadonly)
	return banker.GetCoins(std.CurrentRealm().Address())
}

func Send(cur realm, to std.Address) {
	std.AssertOriginCall()

	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(
		std.CurrentRealm().Address(),
		to,
		std.OriginSend(),
	)
}

func RenderDemo() string {
	content := "# DEMO\n"
	content += "Balance: **" + GetBalance(cross).String() + "**\n\n"
	content += "[Send](sendcoins$help&func=Send)\n\n"
	return content
}
