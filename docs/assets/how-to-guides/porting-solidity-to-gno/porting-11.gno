func AuctionEnd() {
	if std.ChainHeight() < auctionEndBlock {
		panic("Auction hasn't ended")
	}

	if ended {
		panic("Auction has ended")

	}
	ended = true

	// Send the highest bid to the recipient
	banker := std.NewBanker(std.BankerTypeRealmSend)
	pkgAddr := std.OriginPkgAddress()

	banker.SendCoins(pkgAddr, receiver, std.Coins{{"ugnot", int64(highestBid)}})
}
