var (
	receiver        = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	auctionEndBlock = std.GetHeight() + uint(300) // in blocks
	highestBidder   std.Address
	highestBid      = uint(0)
	pendingReturns  avl.Tree
	ended           = false
)
