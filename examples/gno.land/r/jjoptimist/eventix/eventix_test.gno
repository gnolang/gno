package eventix

import (
	"std"
	"testing"

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

	// Using uassert to perform a soft assertion.
	uassert.Equal(t, 1, eventId)

	event, exists := getEvent(eventId)
	// urequire.True stops the test if the event was not created.
	urequire.True(t, exists, "Event was not created")

	// Soft-assert that the event name is as expected.
	uassert.Equal(t, "Test Event", event.name)

	// Test that providing an invalid date format causes a panic.
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
	tokenId := grc721.TokenID("1-1")
	owner, err := tickets.OwnerOf(tokenId) // Handle both return values
	if err != nil {
		t.Errorf("Error getting token owner: %v", err)
	}
	if owner != buyer {
		t.Errorf("Expected ticket owner to be %s, got %s", buyer, owner)
	}

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
