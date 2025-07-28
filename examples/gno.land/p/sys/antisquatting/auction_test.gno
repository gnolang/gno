package antisquatting

import (
	"std"
	"testing"
	"time"
)

func TestDefaultAuctionManager_StartAuction(t *testing.T) {
	am := NewDefaultAuctionManager()
	classifier := NewDefaultNameClassifier()
	
	// Test starting a new auction
	err := am.StartAuction("test", classifier)
	if err != nil {
		t.Fatalf("StartAuction failed: %v", err)
	}
	
	// Verify auction was created
	status, err := am.GetAuctionStatus("test")
	if err != nil {
		t.Fatalf("GetAuctionStatus failed: %v", err)
	}
	
	if status.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", status.Name)
	}
	
	if status.Phase != PhaseCommit {
		t.Errorf("Expected phase %v, got %v", PhaseCommit, status.Phase)
	}
	
	// Test starting duplicate auction
	err = am.StartAuction("test", classifier)
	if err == nil {
		t.Error("Expected error when starting duplicate auction")
	}
}

func TestDefaultAuctionManager_SubmitBid(t *testing.T) {
	am := NewDefaultAuctionManager()
	classifier := NewDefaultNameClassifier()
	
	// Start auction
	err := am.StartAuction("test", classifier)
	if err != nil {
		t.Fatalf("StartAuction failed: %v", err)
	}
	
	bidder := std.Address("g1234567890123456789012345678901234567890")
	bidHash := "a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890"
	deposit := classifier.GetMinimumDeposit("test")
	
	// Test valid bid submission
	err = am.SubmitBid("test", bidder, bidHash, deposit)
	if err != nil {
		t.Fatalf("SubmitBid failed: %v", err)
	}
	
	// Verify bid was recorded
	status, err := am.GetAuctionStatus("test")
	if err != nil {
		t.Fatalf("GetAuctionStatus failed: %v", err)
	}
	
	if status.BidCount != 1 {
		t.Errorf("Expected 1 bid, got %d", status.BidCount)
	}
	
	// Test duplicate bid from same bidder
	err = am.SubmitBid("test", bidder, bidHash, deposit)
	if err != ErrAlreadyCommitted {
		t.Errorf("Expected ErrAlreadyCommitted, got %v", err)
	}
	
	// Test insufficient deposit
	lowDeposit := deposit - 1
	bidder2 := std.Address("g1234567890123456789012345678901234567891")
	err = am.SubmitBid("test", bidder2, bidHash, lowDeposit)
	if err != ErrDepositTooLow {
		t.Errorf("Expected ErrDepositTooLow, got %v", err)
	}
	
	// Test invalid hash format
	invalidHash := "invalid"
	err = am.SubmitBid("test", bidder2, invalidHash, deposit)
	if err != ErrInvalidCommitment {
		t.Errorf("Expected ErrInvalidCommitment, got %v", err)
	}
}

func TestDefaultAuctionManager_RevealBid(t *testing.T) {
	am := NewDefaultAuctionManager()
	classifier := NewDefaultNameClassifier()
	
	// Start auction
	err := am.StartAuction("test", classifier)
	if err != nil {
		t.Fatalf("StartAuction failed: %v", err)
	}
	
	bidder := std.Address("g1234567890123456789012345678901234567890")
	amount := int64(5_000_000)
	salt := "mysalt123"
	
	// Compute correct hash
	bidHash := ComputeBidHash(amount, salt, bidder)
	deposit := classifier.GetMinimumDeposit("test")
	
	// Submit bid
	err = am.SubmitBid("test", bidder, bidHash, deposit)
	if err != nil {
		t.Fatalf("SubmitBid failed: %v", err)
	}
	
	// Move auction to reveal phase by manipulating time
	auction, exists := am.GetAuction("test")
	if !exists {
		t.Fatal("Auction not found")
	}
	auction.Phase = PhaseReveal
	am.auctions.Set("test", auction)
	
	// Test valid reveal
	err = am.RevealBid("test", bidder, amount, salt)
	if err != nil {
		t.Fatalf("RevealBid failed: %v", err)
	}
	
	// Test reveal with wrong amount
	err = am.RevealBid("test", bidder, amount+1, salt)
	if err != ErrRevealMismatch {
		t.Errorf("Expected ErrRevealMismatch, got %v", err)
	}
	
	// Test reveal from non-bidder
	nonBidder := std.Address("g1234567890123456789012345678901234567891")
	err = am.RevealBid("test", nonBidder, amount, salt)
	if err != ErrNotCommitted {
		t.Errorf("Expected ErrNotCommitted, got %v", err)
	}
}

func TestDefaultAuctionManager_FinalizeAuction(t *testing.T) {
	am := NewDefaultAuctionManager()
	classifier := NewDefaultNameClassifier()
	
	// Start auction
	err := am.StartAuction("test", classifier)
	if err != nil {
		t.Fatalf("StartAuction failed: %v", err)
	}
	
	// Add multiple bidders
	bidders := []struct {
		addr   std.Address
		amount int64
		salt   string
	}{
		{std.Address("g1234567890123456789012345678901234567890"), 5_000_000, "salt1"},
		{std.Address("g1234567890123456789012345678901234567891"), 3_000_000, "salt2"},
		{std.Address("g1234567890123456789012345678901234567892"), 7_000_000, "salt3"}, // Winner
	}
	
	deposit := classifier.GetMinimumDeposit("test")
	
	// Submit all bids
	for _, bidder := range bidders {
		bidHash := ComputeBidHash(bidder.amount, bidder.salt, bidder.addr)
		err = am.SubmitBid("test", bidder.addr, bidHash, deposit)
		if err != nil {
			t.Fatalf("SubmitBid failed for %s: %v", bidder.addr, err)
		}
	}
	
	// Move to reveal phase and reveal all bids
	auction, _ := am.GetAuction("test")
	auction.Phase = PhaseReveal
	am.auctions.Set("test", auction)
	
	for _, bidder := range bidders {
		err = am.RevealBid("test", bidder.addr, bidder.amount, bidder.salt)
		if err != nil {
			t.Fatalf("RevealBid failed for %s: %v", bidder.addr, err)
		}
	}
	
	// Finalize auction
	winner, winningBid, err := am.FinalizeAuction("test")
	if err != nil {
		t.Fatalf("FinalizeAuction failed: %v", err)
	}
	
	// Verify winner
	expectedWinner := bidders[2].addr // Highest bidder
	if winner != expectedWinner {
		t.Errorf("Expected winner %s, got %s", expectedWinner, winner)
	}
	
	if winningBid != bidders[2].amount {
		t.Errorf("Expected winning bid %d, got %d", bidders[2].amount, winningBid)
	}
	
	// Verify auction status
	status, err := am.GetAuctionStatus("test")
	if err != nil {
		t.Fatalf("GetAuctionStatus failed: %v", err)
	}
	
	if status.Phase != PhaseFinalized {
		t.Errorf("Expected phase %v, got %v", PhaseFinalized, status.Phase)
	}
	
	if status.Winner != expectedWinner {
		t.Errorf("Expected winner %s, got %s", expectedWinner, status.Winner)
	}
}

func TestDefaultAuctionManager_RefundLosers(t *testing.T) {
	am := NewDefaultAuctionManager()
	classifier := NewDefaultNameClassifier()
	
	// Start and complete an auction (simplified)
	err := am.StartAuction("test", classifier)
	if err != nil {
		t.Fatalf("StartAuction failed: %v", err)
	}
	
	// Add bidders and finalize
	bidder1 := std.Address("g1234567890123456789012345678901234567890")
	bidder2 := std.Address("g1234567890123456789012345678901234567891")
	
	deposit := classifier.GetMinimumDeposit("test")
	
	// Submit bids
	hash1 := ComputeBidHash(3_000_000, "salt1", bidder1)
	hash2 := ComputeBidHash(5_000_000, "salt2", bidder2)
	
	am.SubmitBid("test", bidder1, hash1, deposit)
	am.SubmitBid("test", bidder2, hash2, deposit)
	
	// Move to finalized state
	auction, _ := am.GetAuction("test")
	auction.Phase = PhaseFinalized
	auction.Winner = bidder2
	auction.WinningBid = 5_000_000
	am.auctions.Set("test", auction)
	
	// Test refund
	err = am.RefundLosers("test")
	if err != nil {
		t.Fatalf("RefundLosers failed: %v", err)
	}
	
	// In a real implementation, we would verify that deposits were actually refunded
	// For now, we just verify the function doesn't error
}

func TestDefaultAuctionManager_EdgeCases(t *testing.T) {
	am := NewDefaultAuctionManager()
	
	// Test operations on non-existent auction
	_, err := am.GetAuctionStatus("nonexistent")
	if err != ErrAuctionNotFound {
		t.Errorf("Expected ErrAuctionNotFound, got %v", err)
	}
	
	err = am.SubmitBid("nonexistent", std.Address("g1234567890123456789012345678901234567890"), "hash", 1000)
	if err != ErrAuctionNotFound {
		t.Errorf("Expected ErrAuctionNotFound, got %v", err)
	}
	
	err = am.RevealBid("nonexistent", std.Address("g1234567890123456789012345678901234567890"), 1000, "salt")
	if err != ErrAuctionNotFound {
		t.Errorf("Expected ErrAuctionNotFound, got %v", err)
	}
	
	_, _, err = am.FinalizeAuction("nonexistent")
	if err != ErrAuctionNotFound {
		t.Errorf("Expected ErrAuctionNotFound, got %v", err)
	}
}

func TestComputeBidHash(t *testing.T) {
	bidder := std.Address("g1234567890123456789012345678901234567890")
	amount := int64(5_000_000)
	salt := "mysalt123"
	
	// Test hash computation
	hash1 := ComputeBidHash(amount, salt, bidder)
	hash2 := ComputeBidHash(amount, salt, bidder)
	
	// Same inputs should produce same hash
	if hash1 != hash2 {
		t.Error("Same inputs should produce same hash")
	}
	
	// Different inputs should produce different hashes
	hash3 := ComputeBidHash(amount+1, salt, bidder)
	if hash1 == hash3 {
		t.Error("Different amounts should produce different hashes")
	}
	
	hash4 := ComputeBidHash(amount, salt+"x", bidder)
	if hash1 == hash4 {
		t.Error("Different salts should produce different hashes")
	}
	
	bidder2 := std.Address("g1234567890123456789012345678901234567891")
	hash5 := ComputeBidHash(amount, salt, bidder2)
	if hash1 == hash5 {
		t.Error("Different bidders should produce different hashes")
	}
	
	// Test hash format
	if len(hash1) != 64 {
		t.Errorf("Hash should be 64 characters, got %d", len(hash1))
	}
	
	if !ValidateBidHash(hash1) {
		t.Error("Generated hash should be valid")
	}
}

func TestValidateBidHash(t *testing.T) {
	tests := []struct {
		hash  string
		valid bool
	}{
		{"a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890", true},
		{"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", true},
		{"invalid", false},
		{"", false},
		{"a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890x", false}, // Too long
		{"a1b2c3d4e5f6789012345678901234567890123456789012345678901234567", false},   // Too short
		{"g1b2c3d4e5f6789012345678901234567890123456789012345678901234567890", false}, // Invalid hex
	}
	
	for _, test := range tests {
		result := ValidateBidHash(test.hash)
		if result != test.valid {
			t.Errorf("ValidateBidHash(%q) = %v, expected %v", test.hash, result, test.valid)
		}
	}
}
