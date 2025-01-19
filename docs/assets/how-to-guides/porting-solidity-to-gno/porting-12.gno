// AuctionEnd() Function Test
func TestAuctionEnd(t *testing.T) {
	// Auction is ongoing
	shouldPanic(t, AuctionEnd)

	// Auction ends
	highestBid = 3
	std.TestSkipHeights(500)
	shouldNoPanic(t, AuctionEnd)
	shouldEqual(t, ended, true)

	banker := std.NewBanker(std.BankerTypeRealmSend)
	shouldEqual(t, banker.GetCoins(receiver).String(), "3ugnot")

	// Auction has already ended
	shouldPanic(t, AuctionEnd)
	shouldEqual(t, ended, true)
}
