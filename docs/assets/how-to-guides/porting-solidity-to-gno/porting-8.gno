func Withdraw() {
	// Query the return amount to non-highest bidders
	amount, _ := pendingReturns.Get(std.OriginCaller().String())

	if amount > 0 {
		// If there's an amount, reset the amount first,
		pendingReturns.Set(std.OriginCaller().String(), 0)

		// Return the exceeded amount
		banker := std.NewBanker(std.BankerTypeRealmSend)
		pkgAddr := std.OriginPkgAddress()

		banker.SendCoins(pkgAddr, std.OriginCaller(), std.Coins{{"ugnot", amount.(int64)}})
	}
}
