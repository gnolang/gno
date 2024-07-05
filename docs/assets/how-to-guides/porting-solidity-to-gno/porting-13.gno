// The whole test
func TestFull(t *testing.T) {
	bidder01 := testutils.TestAddress("bidder01") // g1vf5kger9wgcrzh6lta047h6lta047h6lufftkw
	bidder02 := testutils.TestAddress("bidder02") // g1vf5kger9wgcryh6lta047h6lta047h6lnhe2x2

	// Variables test
	{
		shouldEqual(t, highestBidder, "")
		shouldEqual(t, receiver, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
		shouldEqual(t, auctionEndBlock, 423)
		shouldEqual(t, highestBid, 0)
		shouldEqual(t, pendingReturns.Size(), 0)
		shouldEqual(t, ended, false)
	}

	// Send two or more types of coins
	{
		std.TestSetOrigCaller(bidder01)
		std.TestSetOrigSend(std.Coins{{"ugnot", 0}, {"test", 1}}, nil)
		shouldPanic(t, Bid)
	}

	// Send less than the highest bid
	{
		std.TestSetOrigCaller(bidder01)
		std.TestSetOrigSend(std.Coins{{"ugnot", 0}}, nil)
		shouldPanic(t, Bid)
	}

	// Send more than the highest bid
	{
		std.TestSetOrigCaller(bidder01)
		std.TestSetOrigSend(std.Coins{{"ugnot", 1}}, nil)
		shouldNoPanic(t, Bid)

		shouldEqual(t, pendingReturns.Size(), 0)
		shouldEqual(t, highestBid, 1)
		shouldEqual(t, highestBidder, "g1vf5kger9wgcrzh6lta047h6lta047h6lufftkw")
	}

	// Other participants in the auction
	{

		// Send less amount than the current highest bid (current: 1)
		std.TestSetOrigCaller(bidder02)
		std.TestSetOrigSend(std.Coins{{"ugnot", 1}}, nil)
		shouldPanic(t, Bid)

		// Send more amount than the current highest bid (exceeded)
		std.TestSetOrigCaller(bidder02)
		std.TestSetOrigSend(std.Coins{{"ugnot", 2}}, nil)
		shouldNoPanic(t, Bid)

		shouldEqual(t, highestBid, 2)
		shouldEqual(t, highestBidder, "g1vf5kger9wgcryh6lta047h6lta047h6lnhe2x2")

		shouldEqual(t, pendingReturns.Size(), 1) // Return to the existing bidder
		shouldEqual(t, pendingReturns.Has("g1vf5kger9wgcrzh6lta047h6lta047h6lufftkw"), true)
	}

	// Auction ends
	{
		std.TestSkipHeights(150)
		shouldPanic(t, AuctionEnd)
		shouldEqual(t, ended, false)

		std.TestSkipHeights(301)
		shouldNoPanic(t, AuctionEnd)
		shouldEqual(t, ended, true)

		banker := std.GetBanker(std.BankerTypeRealmSend)
		shouldEqual(t, banker.GetCoins(receiver).String(), "2ugnot")
	}
}
