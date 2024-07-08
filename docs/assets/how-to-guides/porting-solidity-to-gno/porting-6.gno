// Bid Function Test - Send Coin
func TestBidCoins(t *testing.T) {
	// Sending two types of coins
	std.TestSetOrigCaller(bidder01)
	std.TestSetOrigSend(std.Coins{{"ugnot", 0}, {"test", 1}}, nil)
	shouldPanic(t, Bid)

	// Sending lower amount than the current highest bid
	std.TestSetOrigCaller(bidder01)
	std.TestSetOrigSend(std.Coins{{"ugnot", 0}}, nil)
	shouldPanic(t, Bid)

	// Sending more amount than the current highest bid (exceeded)
	std.TestSetOrigCaller(bidder01)
	std.TestSetOrigSend(std.Coins{{"ugnot", 1}}, nil)
	shouldNoPanic(t, Bid)
}

// Bid Function Test - Bid by two or more people
func TestBidCoins(t *testing.T) {
	// bidder01 bidding with 1 coin
	std.TestSetOrigCaller(bidder01)
	std.TestSetOrigSend(std.Coins{{"ugnot", 1}}, nil)
	shouldNoPanic(t, Bid)
	shouldEqual(t, highestBid, 1)
	shouldEqual(t, highestBidder, bidder01)
	shouldEqual(t, pendingReturns.Size(), 0)

	// bidder02 bidding with 1 coin
	std.TestSetOrigCaller(bidder02)
	std.TestSetOrigSend(std.Coins{{"ugnot", 1}}, nil)
	shouldPanic(t, Bid)

	// bidder02 bidding with 2 coins
	std.TestSetOrigCaller(bidder02)
	std.TestSetOrigSend(std.Coins{{"ugnot", 2}}, nil)
	shouldNoPanic(t, Bid)
	shouldEqual(t, highestBid, 2)
	shouldEqual(t, highestBidder, bidder02)
	shouldEqual(t, pendingReturns.Size(), 1)
}
