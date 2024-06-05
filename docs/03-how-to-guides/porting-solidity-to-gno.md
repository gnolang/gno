---
id: port-solidity-to-gno
---

# Port a Solidity Contract to a Gno Realm


## Overview

This guide shows you how to port a Solidity contract `Simple Auction` to a Gno Realm `auction.gno` with test cases (Test Driven Development (TDD) approach).

You can check the Solidity contract in this [link](https://docs.soliditylang.org/en/latest/solidity-by-example.html#simple-open-auction), and here's the code for porting.

```solidity
// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.4;
contract SimpleAuction {
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

    /// Bid on the auction with the value sent
    /// together with this transaction.
    /// The value will only be refunded if the
    /// auction is not won.
    function bid() external payable {
        // No arguments are necessary, all
        // information is already part of
        // the transaction. The keyword payable
        // is required for the function to
        // be able to receive Ether.

        // Revert the call if the bidding
        // period is over.
        if (block.timestamp > auctionEndTime)
            revert AuctionAlreadyEnded();

        // If the bid is not higher, send the
        // money back (the revert statement
        // will revert all changes in this
        // function execution including
        // it having received the money).
        if (msg.value <= highestBid)
            revert BidNotHighEnough(highestBid);

        if (highestBid != 0) {
            // Sending back the money by simply using
            // highestBidder.send(highestBid) is a security risk
            // because it could execute an untrusted contract.
            // It is always safer to let the recipients
            // withdraw their money themselves.
            pendingReturns[highestBidder] += highestBid;
        }
        highestBidder = msg.sender;
        highestBid = msg.value;
        emit HighestBidIncreased(msg.sender, msg.value);
    }

    /// Withdraw a bid that was overbid.
    function withdraw() external returns (bool) {
        uint amount = pendingReturns[msg.sender];
        if (amount > 0) {
            // It is important to set this to zero because the recipient
            // can call this function again as part of the receiving call
            // before `send` returns.
            pendingReturns[msg.sender] = 0;

            // msg.sender is not of type `address payable` and must be
            // explicitly converted using `payable(msg.sender)` in order
            // use the member function `send()`.
            if (!payable(msg.sender).send(amount)) {
                // No need to call throw here, just reset the amount owing
                pendingReturns[msg.sender] = amount;
                return false;
            }
        }
        return true;
    }

    /// End the auction and send the highest bid
    /// to the beneficiary.
    function auctionEnd() external {
        // It is a good guideline to structure functions that interact
        // with other contracts (i.e. they call functions or send Ether)
        // into three phases:
        // 1. checking conditions
        // 2. performing actions (potentially changing conditions)
        // 3. interacting with other contracts
        // If these phases are mixed up, the other contract could call
        // back into the current contract and modify the state or cause
        // effects (ether payout) to be performed multiple times.
        // If functions called internally include interaction with external
        // contracts, they also have to be considered interaction with
        // external contracts.

        // 1. Conditions
        if (block.timestamp < auctionEndTime)
            revert AuctionNotYetEnded();
        if (ended)
            revert AuctionEndAlreadyCalled();

        // 2. Effects
        ended = true;
        emit AuctionEnded(highestBidder, highestBid);

        // 3. Interaction
        beneficiary.transfer(highestBid);
    }
}
```

These are the basic concepts of the Simple Auction contract:

* Everyone can send their bids during a bidding period.
* The bids already include sending money / Ether in order to bind the bidders to their bids.
* If the highest bid is raised, the previous highest bidder gets their money back.
* After the end of the bidding period, the contract has to be called manually for the beneficiary to receive their money - contracts cannot activate themselves.

The contract consists of:

* A variable declaration
* Initialization by a constructor
* Three functions

Let's dive into the details of the role of each function, and learn how to port each function into Gno with test cases.

When writing a test case, the following conditions are often used to determine whether the function has been properly executed:

* Value matching
* Error status
* Panic status

Below is a test case helper that will help implement each condition.

### Gno - Testcase Helper

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-1.gno go)
```go
func shouldEqual(t *testing.T, got interface{}, expected interface{}) {
	t.Helper()

	if got != expected {
		t.Errorf("expected %v(%T), got %v(%T)", expected, expected, got, got)
	}
}

func shouldErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected an error, but got nil.")
	}
}

func shouldNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("expected no error, but got err: %s.", err.Error())
	}
}

func shouldPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should have panic")
		}
	}()
	f()
}

func shouldNoPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not have panic")
		}
	}()
	f()
}
```

## Variable init - Solidity

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-2.sol solidity)
```solidity
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
```

* `address payable public beneficiary;` : Address to receive the amount after the auction's ending.
* `uint public auctionEndTime;` : Auction ending time.
* `address public highestBidder;` : The highest bidder.
* `uint public highestBid;` : The highest bid.
* `mapping(address => uint) pendingReturns;` : Bidder's address and amount to be returned (in case of the highest bid changes).
* `bool ended;` : Whether the auction is closed.

### Variable init - Gno

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-3.gno go)
```go
var (
	receiver        = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	auctionEndBlock = std.GetHeight() + uint(300) // in blocks
	highestBidder   std.Address
	highestBid      = uint(0)
	pendingReturns  avl.Tree
	ended           = false
)
```

> **Note:** In Solidity, the Auction ending time is set by a time basis, but in the above case, it's set by a block basis.

###

## bid() - Solidity

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-4.sol solidity)
```solidity
function bid() external payable {
    // No arguments are necessary, all
    // information is already part of
    // the transaction. The keyword payable
    // is required for the function to
    // be able to receive Ether.

    // Revert the call if the bidding
    // period is over.
    if (block.timestamp > auctionEndTime)
        revert AuctionAlreadyEnded();

    // If the bid is not higher, send the
    // money back (the revert statement
    // will revert all changes in this
    // function execution including
    // it having received the money).
    if (msg.value <= highestBid)
        revert BidNotHighEnough(highestBid);

    if (highestBid != 0) {
        // Sending back the money by simply using
        // highestBidder.send(highestBid) is a security risk
        // because it could execute an untrusted contract.
        // It is always safer to let the recipients
        // withdraw their money themselves.
        pendingReturns[highestBidder] += highestBid;
    }
    highestBidder = msg.sender;
    highestBid = msg.value;
    emit HighestBidIncreased(msg.sender, msg.value);
}
```

`bid()` function is for participating in an auction and includes:

* Determining whether an auction is closed.
* Comparing a new bid with the current highest bid.
* Prepare data to return the bid amount to the existing highest bidder in case of the highest bid is increased.
* Update variables with the top bidder & top bid amount.

### bid() - Gno

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-5.gno go)
```go
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
```

### bid() - Gno Testcase

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-6.gno go)
```go
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
```

###

## withdraw() - Solidity

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-7.sol solidity)
```solidity
/// Withdraw a bid that was overbid.
function withdraw() external returns (bool) {
    uint amount = pendingReturns[msg.sender];
    if (amount > 0) {
        // It is important to set this to zero because the recipient
        // can call this function again as part of the receiving call
        // before `send` returns.
        pendingReturns[msg.sender] = 0;

        // msg.sender is not of type `address payable` and must be
        // explicitly converted using `payable(msg.sender)` in order
        // use the member function `send()`.
        if (!payable(msg.sender).send(amount)) {
            // No need to call throw here, just reset the amount owing
            pendingReturns[msg.sender] = amount;
            return false;
        }
    }
    return true;
}
```

`withdraw()` is to return the bid amount to the existing highest bidder in case of the highest bid changes and includes:

* When called, determine if there's a bid amount to be returned to the address.
* (If there's an amount to be returned) Before returning, set the previously recorded amount to `0` and return the actual amount.

### withdraw() - Gno

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-8.gno go)
```go
func Withdraw() {
	// Query the return amount to non-highest bidders
	amount, _ := pendingReturns.Get(std.GetOrigCaller().String())

	if amount > 0 {
		// If there's an amount, reset the amount first,
		pendingReturns.Set(std.GetOrigCaller().String(), 0)

		// Return the exceeded amount
		banker := std.GetBanker(std.BankerTypeRealmSend)
		pkgAddr := std.GetOrigPkgAddr()

		banker.SendCoins(pkgAddr, std.GetOrigCaller(), std.Coins{{"ugnot", amount.(int64)}})
	}
}
```

###

### withdraw() - Gno Testcase

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-9.gno go)
```go
// Withdraw Function Test
func TestWithdraw(t *testing.T) {
	// If there's no participants for return
	shouldEqual(t, pendingReturns.Size(), 0)

	// If there's participants for return (data generation
	returnAddr := bidder01.String()
	returnAmount := int64(3)
	pendingReturns.Set(returnAddr, returnAmount)
	shouldEqual(t, pendingReturns.Size(), 1)
	shouldEqual(t, pendingReturns.Has(returnAddr), true)

	banker := std.GetBanker(std.BankerTypeRealmSend)
	pkgAddr := std.GetOrigPkgAddr()
	banker.SendCoins(pkgAddr, std.Address(returnAddr), std.Coins{{"ugnot", returnAmount}})
	shouldEqual(t, banker.GetCoins(std.Address(returnAddr)).String(), "3ugnot")
}
```

## auctionEnd() - Solidity

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-10.sol solidity)
```solidity
/// End the auction and send the highest bid
/// to the beneficiary.
function auctionEnd() external {
    // It is a good guideline to structure functions that interact
    // with other contracts (i.e. they call functions or send Ether)
    // into three phases:
    // 1. checking conditions
    // 2. performing actions (potentially changing conditions)
    // 3. interacting with other contracts
    // If these phases are mixed up, the other contract could call
    // back into the current contract and modify the state or cause
    // effects (ether payout) to be performed multiple times.
    // If functions called internally include interaction with external
    // contracts, they also have to be considered interaction with
    // external contracts.

    // 1. Conditions
    if (block.timestamp < auctionEndTime)
        revert AuctionNotYetEnded();
    if (ended)
        revert AuctionEndAlreadyCalled();

    // 2. Effects
    ended = true;
    emit AuctionEnded(highestBidder, highestBid);

    // 3. Interaction
    beneficiary.transfer(highestBid);
}
```

`auctionEnd()` function is for ending the auction and includes:

* Determines if the auction should end by comparing the end time.
* Determines if the auction has already ended or not.
  * (If not ended) End the auction.
  * (If not ended) Send the highest bid amount to the recipient.

### auctionEnd() - Gno

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-11.gno go)
```go
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
```

### auctionEnd() - Gno Testcase

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-12.gno go)
```go
// AuctionEnd() Function Test
func TestAuctionEnd(t *testing.T) {
	// Auction is ongoing
	shouldPanic(t, AuctionEnd)

	// Auction ends
	highestBid = 3
	std.TestSkipHeights(500)
	shouldNoPanic(t, AuctionEnd)
	shouldEqual(t, ended, true)

	banker := std.GetBanker(std.BankerTypeRealmSend)
	shouldEqual(t, banker.GetCoins(receiver).String(), "3ugnot")

	// Auction has already ended
	shouldPanic(t, AuctionEnd)
	shouldEqual(t, ended, true)
}
```

## Precautions for Running Test Cases

* Each test function should be executed separately one by one, to return all passes without any errors.
* Same as Go, Gno doesn't support `setup()` & `teardown()` functions. So running two or more test functions simultaneously can result in tainted data.
* If you want to do the whole test at once, make it into a single function as below:

[embedmd]:# (../assets/how-to-guides/porting-solidity-to-gno/porting-13.gno go)
```go
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
```
