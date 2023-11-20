// Parameters of the auction. Times are either
// absolute unix timestamps (seconds since 1970-01-01)
// or time periods in seconds.
address payable public beneficiary;
uint public auctionEndTime;

// Current state of the auction.
address public highestBidder;
uint public highestBid;

// Allowed withdrawals of previous bids
mapping(address => uint) pendingReturns;

// Set to true at the end, disallows any change.
// By default initialized to `false`.
bool ended;

// Events that will be emitted on changes.
event HighestBidIncreased(address bidder, uint amount);
event AuctionEnded(address winner, uint amount);

// Errors that describe failures.

// The triple-slash comments are so-called natspec
// comments. They will be shown when the user
// is asked to confirm a transaction or
// when an error is displayed.

/// The auction has already ended.
error AuctionAlreadyEnded();
/// There is already a higher or equal bid.
error BidNotHighEnough(uint highestBid);
/// The auction has not ended yet.
error AuctionNotYetEnded();
/// The function auctionEnd has already been called.
error AuctionEndAlreadyCalled();

/// Create a simple auction with `biddingTime`
/// seconds bidding time on behalf of the
/// beneficiary address `beneficiaryAddress`.
constructor(
    uint biddingTime,
    address payable beneficiaryAddress
) {
    beneficiary = beneficiaryAddress;
    auctionEndTime = block.timestamp + biddingTime;
}