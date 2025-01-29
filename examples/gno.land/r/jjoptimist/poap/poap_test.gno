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

func TestCreatePOAP(t *testing.T) {
	date := time.Now()
	id := CreatePOAP(
		"Gno Launch Event",
		"Celebrating Gno launch",
		date,
		100,
	)

	if id != 1 {
		t.Errorf("Expected first POAP ID to be 1, got %d", id)
	}

	poap, exists := GetPOAP(id)
	if !exists {
		t.Error("POAP was not created")
	}

	if poap.name != "Gno Launch Event" {
		t.Errorf("Expected name 'Gno Launch Event', got '%s'", poap.name)
	}

	if poap.maxMints != 100 {
		t.Errorf("Expected maxMints 100, got %d", poap.maxMints)
	}

	if poap.mintCount != 0 {
		t.Errorf("Expected mintCount 0, got %d", poap.mintCount)
	}

	// Test invalid maxMints
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with invalid maxMints")
		}
	}()
	CreatePOAP("Invalid", "Test", date, 0)
}
