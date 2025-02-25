// Bid Function Test - Send Coin
func TestBidCoins(t *testing.T) {
	// Sending two types of coins
	testing.SetOriginCaller(bidder01)
	testing.SetOriginSend(std.Coins{{"ugnot", 0}, {"test", 1}})
	shouldPanic(t, Bid)

	// Sending lower amount than the current highest bid
	testing.SetOriginCaller(bidder01)
	testing.SetOriginSend(std.Coins{{"ugnot", 0}})
	shouldPanic(t, Bid)

	// Sending more amount than the current highest bid (exceeded)
	testing.SetOriginCaller(bidder01)
	testing.SetOriginSend(std.Coins{{"ugnot", 1}})
	shouldNoPanic(t, Bid)
}

// Bid Function Test - Bid by two or more people
func TestBidCoins(t *testing.T) {
	// bidder01 bidding with 1 coin
	testing.SetOriginCaller(bidder01)
	testing.SetOriginSend(std.Coins{{"ugnot", 1}})
	shouldNoPanic(t, Bid)
	shouldEqual(t, highestBid, 1)
	shouldEqual(t, highestBidder, bidder01)
	shouldEqual(t, pendingReturns.Size(), 0)

	// bidder02 bidding with 1 coin
	testing.SetOriginCaller(bidder02)
	testing.SetOriginSend(std.Coins{{"ugnot", 1}})
	shouldPanic(t, Bid)

	// bidder02 bidding with 2 coins
	testing.SetOriginCaller(bidder02)
	testing.SetOriginSend(std.Coins{{"ugnot", 2}})
	shouldNoPanic(t, Bid)
	shouldEqual(t, highestBid, 2)
	shouldEqual(t, highestBidder, bidder02)
	shouldEqual(t, pendingReturns.Size(), 1)
}
