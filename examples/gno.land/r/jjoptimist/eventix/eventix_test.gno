package eventix

import (
	"fmt"
	"std"
	"testing"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestCreateEvent(t *testing.T) {
	// Test successful event creation
	eventId := CreateEvent(
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
	urequire.PanicsWithMessage(t, "Invalid date format. Use: YYYY-MM-DDThh:mm:ssZ", func() {
		CreateEvent("Test", "Test", "invalid-date", "ugnot", 100, 1000000)
	})
}

func TestBuyTicket(t *testing.T) {
	// Setup test event
	eventId := CreateEvent(
		"Test Event",
		"A test event",
		"2024-12-31T23:59:59Z",
		"ugnot",
		2,
		1000000,
	)

	// Setup test buyer
	buyer := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	std.TestSetOrigCaller(buyer)

	// Test buying without payment
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when buying without payment")
		}
	}()
	BuyTicket(eventId)

	// Test successful purchase
	std.TestSetOrigSend(std.NewCoins(std.NewCoin("ugnot", 1000000)), std.Coins{})
	BuyTicket(eventId)

	event, _ := getEvent(eventId)
	if event.ticketsSold != 1 {
		t.Errorf("Expected 1 ticket sold, got %d", event.ticketsSold)
	}

	// Verify NFT ownership
	tokenId := grc721.TokenID(fmt.Sprintf("event_%d_ticket_%d", eventId, 1))
	owner, err := tickets.OwnerOf(tokenId)
	urequire.NoError(t, err)
	urequire.Equal(t, buyer, owner)

	// Test buying sold out event
	std.TestSetOrigSend(std.NewCoins(std.NewCoin("ugnot", 1000000)), std.Coins{})
	BuyTicket(eventId) // Buy second ticket
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when event is sold out")
		}
	}()
	BuyTicket(eventId) // Should panic - sold out
}

func TestBuyTicketWithGRC20(t *testing.T) {
	// Set up package address first
	pkgAddr := std.GetOrigPkgAddr()
	std.TestSetOrigPkgAddr(pkgAddr)

	// Create a test GRC20 token and set up addresses
	token, ledger := grc20.NewToken("Test Token", "TEST", 6)
	buyer := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	adminAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf6")

	// Set up admin and mint tokens
	std.TestSetOrigCaller(adminAddr)
	std.TestSetRealm(std.NewUserRealm(adminAddr))
	err := ledger.Mint(buyer, 2000000) // Use ledger to mint
	urequire.NoError(t, err)

	// Create event with GRC20 token payment
	std.TestSetOrigCaller(buyer)
	std.TestSetRealm(std.NewUserRealm(buyer))

	eventId := CreateEvent(
		"GRC20 Event",
		"An event with GRC20 payment",
		"2024-12-31T23:59:59Z",
		token.RealmTeller(),
		2,
		1000000,
	)

	// Set up for approval
	std.TestSetOrigCaller(buyer)
	std.TestSetRealm(std.NewUserRealm(buyer))

	// Approve using the ledger (following grc20factory_test pattern)
	err = ledger.Approve(buyer, pkgAddr, 1000000)
	urequire.NoError(t, err)

	// Verify approval
	allowance := token.Allowance(buyer, pkgAddr)
	urequire.Equal(t, uint64(1000000), allowance, "Approval should be set correctly")

	// Buy ticket
	std.TestSetOrigCaller(buyer)
	std.TestSetRealm(std.NewUserRealm(buyer))
	BuyTicket(eventId)

	// Verify purchase
	event, exists := getEvent(eventId)
	urequire.True(t, exists)
	urequire.Equal(t, 1, event.ticketsSold)

	// Verify NFT ownership
	tokenId := grc721.TokenID(fmt.Sprintf("event_%d_ticket_%d", eventId, 1))
	owner, err := tickets.OwnerOf(tokenId)
	urequire.NoError(t, err)
	urequire.Equal(t, buyer, owner)

	// Verify GRC20 balance changes
	buyerBalance := token.BalanceOf(buyer)
	urequire.Equal(t, uint64(1000000), buyerBalance) // Should have 1M left after spending 1M
}
