func Bid() {
	if std.GetHeight() > auctionEndBlock {
		panic("Exceeded auction end block")
	}

	sentCoins := std.GetOrigSend()
	if len(sentCoins) != 1 {
		panic("Send only one type of coin")
	}

	sentAmount := uint(sentCoins[0].Amount)
	if sentAmount <= highestBid {
		panic("Too few coins sent")
	}

	// A new bid is higher than the current highest bid
	if sentAmount > highestBid {
		// If the highest bid is greater than 0,
		if highestBid > 0 {
			// Need to return the bid amount to the existing highest bidder
			// Create an AVL tree and save
			pendingReturns.Set(highestBidder.String(), highestBid)
		}

		// Update the top bidder address
		highestBidder = std.GetOrigCaller()
		// Update the top bid amount
		highestBid = sentAmount
	}
}
