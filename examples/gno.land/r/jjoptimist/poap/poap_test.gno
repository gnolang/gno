package poap

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
)

func resetState() {
	poaps = avl.NewTree()
	poapCount = 0
	nftTokens = &token{Tokens: avl.NewTree()}
}

func TestPOAPStructure(t *testing.T) {
	resetState()
	date := time.Now()
	poap := POAP{
		ID:          1,
		Name:        "Test Event",
		Description: "A test event",
		Date:        date,
		MaxMints:    100,
		MintCount:   0,
	}

	if poap.Name != "Test Event" {
		t.Errorf("Expected name 'Test Event', got '%s'", poap.Name)
	}

	if poap.MaxMints != 100 {
		t.Errorf("Expected maxMints 100, got %d", poap.MaxMints)
	}
}

func TestCreatePOAP(t *testing.T) {
	resetState()
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

	if poap.Name != "Gno Launch Event" {
		t.Errorf("Expected name 'Gno Launch Event', got '%s'", poap.Name)
	}

	if poap.MaxMints != 100 {
		t.Errorf("Expected maxMints 100, got %d", poap.MaxMints)
	}

	if poap.MintCount != 0 {
		t.Errorf("Expected mintCount 0, got %d", poap.MintCount)
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
	resetState()
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
	if poap.MintCount != 1 {
		t.Errorf("Expected mintCount 1, got %d", poap.MintCount)
	}

	// Verify NFT ownership
	tokenId := grc721.TokenID("1-1") // poapId is 1, first mint
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
	MintPOAP(poapId) // Should panic - limit reached
}

func TestStorageFunctions(t *testing.T) {
	resetState()
	// Setup test data
	date := time.Now()

	// Create two POAPs
	poap1Id := CreatePOAP("Event 1", "First event", date, 2)
	poap2Id := CreatePOAP("Event 2", "Second event", date, 2)

	// Setup test addresses
	addr1 := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	addr2 := std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")

	// Test ListPOAPs
	allPOAPs := ListPOAPs()
	if len(allPOAPs) != 2 {
		t.Errorf("Expected 2 POAPs, got %d", len(allPOAPs))
	}

	// Mint POAPs to different addresses
	std.TestSetOrigCaller(addr1)
	MintPOAP(poap1Id) // addr1 mints from Event 1

	std.TestSetOrigCaller(addr2)
	MintPOAP(poap1Id) // addr2 mints from Event 1
	MintPOAP(poap2Id) // addr2 mints from Event 2

	// Test GetPOAPsByOwner
	addr1POAPs := GetPOAPsByOwner(addr1)
	if len(addr1POAPs) != 1 {
		t.Errorf("Expected addr1 to have 1 POAP, got %d", len(addr1POAPs))
	}

	addr2POAPs := GetPOAPsByOwner(addr2)
	if len(addr2POAPs) != 2 {
		t.Errorf("Expected addr2 to have 2 POAPs, got %d", len(addr2POAPs))
	}

	// Test GetPOAPTokens
	poap1Tokens := GetPOAPTokens(poap1Id)
	if len(poap1Tokens) != 2 {
		t.Errorf("Expected Event 1 to have 2 tokens, got %d", len(poap1Tokens))
	}

	poap2Tokens := GetPOAPTokens(poap2Id)
	if len(poap2Tokens) != 1 {
		t.Errorf("Expected Event 2 to have 1 token, got %d", len(poap2Tokens))
	}

	// Verify token ownership
	var addr1TokenFound, addr2TokenFound bool
	for _, token := range poap1Tokens {
		if token.Owner == addr1 {
			addr1TokenFound = true
		}
		if token.Owner == addr2 {
			addr2TokenFound = true
		}
	}

	if !addr1TokenFound {
		t.Error("Token owned by addr1 not found in Event 1 tokens")
	}
	if !addr2TokenFound {
		t.Error("Token owned by addr2 not found in Event 1 tokens")
	}
}
