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

func TestMintPOAP(t *testing.T) {
	// Setup test POAP
	date := time.Now()
	poapId := CreatePOAP(
		"Test Event",
		"A test event",
		date,
		2, // max mints
	)
	
	// Setup test minter
	minter := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	std.TestSetOrigCaller(minter)
	
	// Test successful mint
	MintPOAP(poapId)
	
	poap, _ := GetPOAP(poapId)
	if poap.mintCount != 1 {
		t.Errorf("Expected mintCount 1, got %d", poap.mintCount)
	}
	
	// Verify NFT ownership
	tokenId := grc721.TokenID("1-1")  
	owner, err := nftTokens.OwnerOf(tokenId)
	if err != nil {
		t.Errorf("Error getting token owner: %v", err)
	}
	if owner != minter {
		t.Errorf("Expected token owner to be %s, got %s", minter, owner)
	}
	
	// Test minting limit
	MintPOAP(poapId)  
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when exceeding mint limit")
		}
	}()
	MintPOAP(poapId)  // Should panic - limit reached
}
