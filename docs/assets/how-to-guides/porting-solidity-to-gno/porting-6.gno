// Bid Function Test - Send Coin
func TestBidCoins(t *testing.T) {
	// Sending two types of coins
	t.SetOriginCaller(bidder01)
	t.SetOriginSend(std.Coins{{"ugnot", 0}, {"test", 1}})
	shouldPanic(t, Bid)

	// Sending lower amount than the current highest bid
	t.SetOriginCaller(bidder01)
	t.SetOriginSend(std.Coins{{"ugnot", 0}})
	shouldPanic(t, Bid)

	// Sending more amount than the current highest bid (exceeded)
	t.SetOriginCaller(bidder01)
	t.SetOriginSend(std.Coins{{"ugnot", 1}})
	shouldNoPanic(t, Bid)
}

// Bid Function Test - Bid by two or more people
func TestBidCoins(t *testing.T) {
	// bidder01 bidding with 1 coin
	t.SetOriginCaller(bidder01)
	t.SetOriginSend(std.Coins{{"ugnot", 1}})
	shouldNoPanic(t, Bid)
	shouldEqual(t, highestBid, 1)
	shouldEqual(t, highestBidder, bidder01)
	shouldEqual(t, pendingReturns.Size(), 0)

	// bidder02 bidding with 1 coin
	t.SetOriginCaller(bidder02)
	t.SetOriginSend(std.Coins{{"ugnot", 1}})
	shouldPanic(t, Bid)

	// bidder02 bidding with 2 coins
	t.SetOriginCaller(bidder02)
	t.SetOriginSend(std.Coins{{"ugnot", 2}})
	shouldNoPanic(t, Bid)
	shouldEqual(t, highestBid, 2)
	shouldEqual(t, highestBidder, bidder02)
	shouldEqual(t, pendingReturns.Size(), 1)
}
