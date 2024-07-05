func AuctionEnd() {
	if std.GetHeight() < auctionEndBlock {
		panic("Auction hasn't ended")
	}

	if ended {
		panic("Auction has ended")

	}
	ended = true

	// Send the highest bid to the recipient
	banker := std.GetBanker(std.BankerTypeRealmSend)
	pkgAddr := std.GetOrigPkgAddr()

	banker.SendCoins(pkgAddr, receiver, std.Coins{{"ugnot", int64(highestBid)}})
}
