func Withdraw() {
	// Query the return amount to non-highest bidders
	amount, _ := pendingReturns.Get(std.GetOrigCaller().String())

	if amount > 0 {
		// If there's an amount, reset the amount first,
		pendingReturns.Set(std.GetOrigCaller().String(), 0)

		// Return the exceeded amount
		banker := std.GetBanker(std.BankerTypeRealmSend)
		pkgAddr := std.GetOrigPkgAddr()

		banker.SendCoins(pkgAddr, std.GetOrigCaller(), std.Coins{{"ugnot", amount.(int64)}})
	}
}
