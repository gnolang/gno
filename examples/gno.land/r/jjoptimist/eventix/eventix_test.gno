package eventix

import (
	"chain"
	"chain/runtime"
	"testing"

	"gno.land/p/demo/tokens/grc20"
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
	"gno.land/p/nt/urequire"
)

func TestCreateEvent(t *testing.T) {
	// Test successful event creation
	eventId := CreateEvent(
		cross,
		"Test Event",
		"A test event",
		"2024-12-31T23:59:59Z",
		"ugnot",
		100,
		1000000,
	)

	// Convert eventId to int for comparison since CreateEvent returns uint64
	uassert.Equal(t, uint64(1), eventId)

	event, exists := getEvent(eventId)
	urequire.True(t, exists, "Event was not created")
	uassert.Equal(t, "Test Event", event.name)

	// Test invalid date format
	urequire.AbortsWithMessage(t, "Invalid date format. Use: YYYY-MM-DDThh:mm:ssZ", func() {
		CreateEvent(cross, "Test", "Test", "invalid-date", "ugnot", 100, 1000000)
	})
}

func TestBuyTicket(t *testing.T) {
	// Setup test event
	eventId := CreateEvent(
		cross,
		"Test Event",
		"A test event",
		"2024-12-31T23:59:59Z",
		"ugnot",
		2,
		1000000,
	)

	// Setup test buyer
	buyer := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	testing.SetOriginCaller(buyer)

	uassert.AbortsWithMessage(t, "Invalid payment amount: needs to be 1000000ugnot", func() {
		BuyTicket(cross, eventId)
	})

	// Test successful purchase
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", 1000000)))
	BuyTicket(cross, eventId)

	event, _ := getEvent(eventId)
	if event.ticketsSold != 1 {
		t.Errorf("Expected 1 ticket sold, got %d", event.ticketsSold)
	}

	// Verify NFT ownership
	tokenId := grc721.TokenID(ufmt.Sprintf("event_%d_ticket_%d", eventId, 1))
	owner, err := tickets.OwnerOf(tokenId)
	urequire.NoError(t, err)
	urequire.Equal(t, buyer, owner)

	// Test buying sold out event
	testing.SetOriginSend(chain.NewCoins(chain.NewCoin("ugnot", 1000000)))
	BuyTicket(cross, eventId) // Buy second ticket

	uassert.AbortsWithMessage(t, "Event is sold out", func() {
		BuyTicket(cross, eventId)
	})
}

func TestBuyTicketWithGRC20(t *testing.T) {
	pkgAddr := chain.PackageAddress("gno.land/r/jjoptimist/eventix")
	alice := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(alice))

	// Create a test GRC20 token and set up addresses
	token, ledger := grc20.NewToken("Test Token", "TEST", 6)
	buyer := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	adminAddr := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf6")

	// Set up admin and mint tokens
	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(testing.NewUserRealm(adminAddr))
	err := ledger.Mint(buyer, 2000000) // Use ledger to mint
	urequire.NoError(t, err)

	// Create event with GRC20 token payment
	testing.SetOriginCaller(buyer)
	testing.SetRealm(testing.NewUserRealm(buyer))

	var eventId uint64
	crossThrough(testing.NewCodeRealm("gno.land/r/jjoptimist/eventix"), func() {
		eventId = CreateEvent(
			cross,
			"GRC20 Event",
			"An event with GRC20 payment",
			"2024-12-31T23:59:59Z",
			token,
			2,
			1000000,
		)
	})

	// Set up for approval
	testing.SetOriginCaller(buyer)
	testing.SetRealm(testing.NewUserRealm(buyer))

	// Approve using the ledger (following grc20factory_test pattern)
	err = ledger.Approve(buyer, pkgAddr, 1000000)
	urequire.NoError(t, err)

	// Verify approval
	allowance := token.Allowance(buyer, pkgAddr)
	urequire.Equal(t, int64(1000000), allowance, "Approval should be set correctly")

	testing.SetOriginCaller(buyer)
	testing.SetRealm(testing.NewUserRealm(buyer))

	// Buy ticket
	BuyTicket(cross, eventId)

	// Verify purchase
	event, exists := getEvent(eventId)
	urequire.True(t, exists)
	urequire.Equal(t, 1, event.ticketsSold)

	// Verify NFT ownership
	tokenId := grc721.TokenID(ufmt.Sprintf("event_%d_ticket_%d", eventId, 1))
	owner, err := tickets.OwnerOf(tokenId)
	urequire.NoError(t, err)
	urequire.Equal(t, buyer, owner)

	// Verify GRC20 balance changes
	buyerBalance := token.BalanceOf(buyer)
	urequire.Equal(t, int64(1000000), buyerBalance) // Should have 1M left after spending 1M
}

func crossThrough(rlm runtime.Realm, cr func()) {
	testing.SetRealm(rlm)
	cr()
}
