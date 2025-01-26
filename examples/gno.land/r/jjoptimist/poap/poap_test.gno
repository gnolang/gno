package poap

import (
    "testing"
    "time"
)

func TestPOAPStructure(t *testing.T) { 
    date := time.Now()
    poap := POAP{
        id:          1,
        name:        "Test Event",
        description: "A test event",
        date:        date,
        maxMints:    100,
        mintCount:   0,
    }
    
    if poap.name != "Test Event" {
        t.Errorf("Expected name 'Test Event', got '%s'", poap.name)
    }
    
    if poap.maxMints != 100 {
        t.Errorf("Expected maxMints 100, got %d", poap.maxMints)
    }
} 