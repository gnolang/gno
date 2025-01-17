package eventix

import (
    "std"
    "testing"
    "time"
    
    "gno.land/p/demo/grc/grc721"
)

func TestCreateEvent(t *testing.T) {
    // Test successful event creation
    eventId := CreateEvent(
        "Test Event",
        "A test event",
        "2024-12-31T23:59:59Z",
        100,
        1000000,
    )
    
    if eventId != 1 {
        t.Errorf("Expected first event ID to be 1, got %d", eventId)
    }
    
    event, exists := getEvent(eventId)
    if !exists {
        t.Error("Event was not created")
    }
    
    if event.name != "Test Event" {
        t.Errorf("Expected event name 'Test Event', got '%s'", event.name)
    }
    
    // Test invalid date format
    defer func() {
        if r := recover(); r == nil {
            t.Error("Expected panic with invalid date format")
        }
    }()
    CreateEvent("Test", "Test", "invalid-date", 100, 1000000)
}

func TestBuyTicket(t *testing.T) {
    // Setup test event
    eventId := CreateEvent(
        "Test Event",
        "A test event",
        "2024-12-31T23:59:59Z",
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
    std.TestSetOrigSend(std.Coins{{"ugnot", 1000000}}, nil)
    BuyTicket(eventId)
    
    event, _ := getEvent(eventId)
    if event.ticketsSold != 1 {
        t.Errorf("Expected 1 ticket sold, got %d", event.ticketsSold)
    }
    
    // Verify NFT ownership
    tokenId := grc721.TokenID("1-1")
    owner, err := tickets.OwnerOf(tokenId)  // Handle both return values
    if err != nil {
        t.Errorf("Error getting token owner: %v", err)
    }
    if owner != buyer {
        t.Errorf("Expected ticket owner to be %s, got %s", buyer, owner)
    }
    
    // Test buying sold out event
    std.TestSetOrigSend(std.Coins{{"ugnot", 1000000}}, nil)
    BuyTicket(eventId)  // Buy second ticket
    defer func() {
        if r := recover(); r == nil {
            t.Error("Expected panic when event is sold out")
        }
    }()
    BuyTicket(eventId)  // Should panic - sold out
} 