package eventix

import (
    "std"
    "testing"
    "time"
)

func TestCreateEvent(t *testing.T) {
    date := time.Now().Add(24 * time.Hour)
    eventId := CreateEvent("Test Event", "A test event", date, 100, 1000000)
    
    if eventId != 1 {
        t.Errorf("Expected first event ID to be 1, got %d", eventId)
    }
    
    event, exists := events[eventId]
    if !exists {
        t.Error("Event was not created")
    }
    
    if event.name != "Test Event" {
        t.Errorf("Expected event name 'Test Event', got '%s'", event.name)
    }
}

func TestBuyTicket(t *testing.T) {
    // Setup test event
    date := time.Now().Add(24 * time.Hour)
    eventId := CreateEvent("Test Event", "A test event", date, 2, 1000000)
    
    // Setup test buyer
    buyer := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
    std.TestSetOrigCaller(buyer)
    
    // Buy ticket
    BuyTicket(eventId)
    
    event := events[eventId]
    if event.ticketsSold != 1 {
        t.Errorf("Expected 1 ticket sold, got %d", event.ticketsSold)
    }
    
    // Verify NFT ownership
    tokenId := "1-1"  // eventId-ticketNumber
    owner := tickets.OwnerOf(tokenId)
    if owner != buyer {
        t.Errorf("Expected ticket owner to be %s, got %s", buyer, owner)
    }
} 