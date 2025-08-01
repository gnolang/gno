package antisquatting

import (
	"std"
	"testing"
	"time"
)

func TestDefaultDisputeResolver_CreateDispute(t *testing.T) {
	dr := NewDefaultDisputeResolver()
	
	name := "admin"
	owner := std.Address("g1234567890123456789012345678901234567890")
	disputer := std.Address("g1234567890123456789012345678901234567891")
	reason := "This name violates trademark"
	
	// Test creating a new dispute
	disputeID, err := dr.CreateDispute(name, owner, disputer, reason)
	if err != nil {
		t.Fatalf("CreateDispute failed: %v", err)
	}
	
	if disputeID == "" {
		t.Error("Expected non-empty dispute ID")
	}
	
	// Verify dispute was created
	dispute, exists := dr.GetDispute(disputeID)
	if !exists {
		t.Fatal("Dispute not found after creation")
	}
	
	if dispute.Name != name {
		t.Errorf("Expected name '%s', got '%s'", name, dispute.Name)
	}
	
	if dispute.Owner != owner {
		t.Errorf("Expected owner %s, got %s", owner, dispute.Owner)
	}
	
	if dispute.Disputer != disputer {
		t.Errorf("Expected disputer %s, got %s", disputer, dispute.Disputer)
	}
	
	if dispute.Reason != reason {
		t.Errorf("Expected reason '%s', got '%s'", reason, dispute.Reason)
	}
	
	if dispute.Status != DisputePending {
		t.Errorf("Expected status %v, got %v", DisputePending, dispute.Status)
	}
	
	// Test duplicate dispute
	_, err = dr.CreateDispute(name, owner, disputer, reason)
	if err != ErrDisputeExists {
		t.Errorf("Expected ErrDisputeExists, got %v", err)
	}
}

func TestDefaultDisputeResolver_CanDispute(t *testing.T) {
	dr := NewDefaultDisputeResolver()
	
	name := "admin"
	owner := std.Address("g1234567890123456789012345678901234567890")
	disputer := std.Address("g1234567890123456789012345678901234567891")
	
	// Test disputing allowed initially
	if !dr.CanDispute(name, disputer) {
		t.Error("Should be able to dispute initially")
	}
	
	// Create a dispute
	disputeID, err := dr.CreateDispute(name, owner, disputer, "test reason")
	if err != nil {
		t.Fatalf("CreateDispute failed: %v", err)
	}
	
	// Test disputing not allowed after creating dispute
	if dr.CanDispute(name, disputer) {
		t.Error("Should not be able to dispute after creating dispute")
	}
	
	// Test different disputer can still dispute
	disputer2 := std.Address("g1234567890123456789012345678901234567892")
	if !dr.CanDispute(name, disputer2) {
		t.Error("Different disputer should be able to dispute")
	}
	
	// Resolve the dispute
	err = dr.ResolveDispute(disputeID, false)
	if err != nil {
		t.Fatalf("ResolveDispute failed: %v", err)
	}
	
	// Test disputing allowed again after resolution
	if !dr.CanDispute(name, disputer) {
		t.Error("Should be able to dispute again after resolution")
	}
}

func TestDefaultDisputeResolver_ResolveDispute(t *testing.T) {
	dr := NewDefaultDisputeResolver()
	
	name := "admin"
	owner := std.Address("g1234567890123456789012345678901234567890")
	disputer := std.Address("g1234567890123456789012345678901234567891")
	reason := "Test dispute"
	
	// Create dispute
	disputeID, err := dr.CreateDispute(name, owner, disputer, reason)
	if err != nil {
		t.Fatalf("CreateDispute failed: %v", err)
	}
	
	// Test approving dispute
	err = dr.ResolveDispute(disputeID, true)
	if err != nil {
		t.Fatalf("ResolveDispute failed: %v", err)
	}
	
	// Verify dispute status
	status, err := dr.GetDisputeStatus(disputeID)
	if err != nil {
		t.Fatalf("GetDisputeStatus failed: %v", err)
	}
	
	if status != DisputeApproved {
		t.Errorf("Expected status %v, got %v", DisputeApproved, status)
	}
	
	// Test resolving already resolved dispute
	err = dr.ResolveDispute(disputeID, false)
	if err != ErrDisputeAlreadyResolved {
		t.Errorf("Expected ErrDisputeAlreadyResolved, got %v", err)
	}
	
	// Create another dispute and reject it
	disputeID2, err := dr.CreateDispute("test2", owner, disputer, reason)
	if err != nil {
		t.Fatalf("CreateDispute failed: %v", err)
	}
	
	err = dr.ResolveDispute(disputeID2, false)
	if err != nil {
		t.Fatalf("ResolveDispute failed: %v", err)
	}
	
	status2, err := dr.GetDisputeStatus(disputeID2)
	if err != nil {
		t.Fatalf("GetDisputeStatus failed: %v", err)
	}
	
	if status2 != DisputeRejected {
		t.Errorf("Expected status %v, got %v", DisputeRejected, status2)
	}
}

func TestDefaultDisputeResolver_GetDisputeStatus(t *testing.T) {
	dr := NewDefaultDisputeResolver()
	
	// Test non-existent dispute
	_, err := dr.GetDisputeStatus("nonexistent")
	if err != ErrDisputeNotFound {
		t.Errorf("Expected ErrDisputeNotFound, got %v", err)
	}
	
	// Create dispute and test status
	name := "admin"
	owner := std.Address("g1234567890123456789012345678901234567890")
	disputer := std.Address("g1234567890123456789012345678901234567891")
	
	disputeID, err := dr.CreateDispute(name, owner, disputer, "test")
	if err != nil {
		t.Fatalf("CreateDispute failed: %v", err)
	}
	
	status, err := dr.GetDisputeStatus(disputeID)
	if err != nil {
		t.Fatalf("GetDisputeStatus failed: %v", err)
	}
	
	if status != DisputePending {
		t.Errorf("Expected status %v, got %v", DisputePending, status)
	}
}

func TestDefaultDisputeResolver_GetActiveDisputes(t *testing.T) {
	dr := NewDefaultDisputeResolver()
	
	// Initially no active disputes
	active := dr.GetActiveDisputes()
	if len(active) != 0 {
		t.Errorf("Expected 0 active disputes, got %d", len(active))
	}
	
	// Create some disputes
	owner := std.Address("g1234567890123456789012345678901234567890")
	disputer := std.Address("g1234567890123456789012345678901234567891")
	
	disputeID1, _ := dr.CreateDispute("name1", owner, disputer, "reason1")
	disputeID2, _ := dr.CreateDispute("name2", owner, disputer, "reason2")
	
	// Check active disputes
	active = dr.GetActiveDisputes()
	if len(active) != 2 {
		t.Errorf("Expected 2 active disputes, got %d", len(active))
	}
	
	// Resolve one dispute
	dr.ResolveDispute(disputeID1, true)
	
	// Check active disputes again
	active = dr.GetActiveDisputes()
	if len(active) != 1 {
		t.Errorf("Expected 1 active dispute, got %d", len(active))
	}
	
	if active[0] != disputeID2 {
		t.Errorf("Expected active dispute %s, got %s", disputeID2, active[0])
	}
}

func TestDefaultDisputeResolver_GetDisputeStatistics(t *testing.T) {
	dr := NewDefaultDisputeResolver()
	
	// Initially no statistics
	stats := dr.GetDisputeStatistics()
	if stats["total"] != 0 {
		t.Errorf("Expected 0 total disputes, got %d", stats["total"])
	}
	
	// Create and resolve some disputes
	owner := std.Address("g1234567890123456789012345678901234567890")
	disputer := std.Address("g1234567890123456789012345678901234567891")
	
	disputeID1, _ := dr.CreateDispute("name1", owner, disputer, "reason1")
	disputeID2, _ := dr.CreateDispute("name2", owner, disputer, "reason2")
	disputeID3, _ := dr.CreateDispute("name3", owner, disputer, "reason3")
	
	// Resolve disputes with different outcomes
	dr.ResolveDispute(disputeID1, true)  // Approved
	dr.ResolveDispute(disputeID2, false) // Rejected
	// disputeID3 remains pending
	
	// Check statistics
	stats = dr.GetDisputeStatistics()
	
	if stats["total"] != 3 {
		t.Errorf("Expected 3 total disputes, got %d", stats["total"])
	}
	
	if stats["pending"] != 1 {
		t.Errorf("Expected 1 pending dispute, got %d", stats["pending"])
	}
	
	if stats["approved"] != 1 {
		t.Errorf("Expected 1 approved dispute, got %d", stats["approved"])
	}
	
	if stats["rejected"] != 1 {
		t.Errorf("Expected 1 rejected dispute, got %d", stats["rejected"])
	}
}

func TestDefaultDisputeResolver_Configuration(t *testing.T) {
	dr := NewDefaultDisputeResolver()
	
	// Test setting dispute fee
	originalFee := dr.GetDisputeFee()
	newFee := int64(2_000_000)
	
	dr.SetDisputeFee(newFee)
	if dr.GetDisputeFee() != newFee {
		t.Errorf("Expected dispute fee %d, got %d", newFee, dr.GetDisputeFee())
	}
	
	// Reset
	dr.SetDisputeFee(originalFee)
	
	// Test setting voting period
	originalPeriod := dr.GetVotingPeriod()
	newPeriod := 48 * time.Hour
	
	dr.SetVotingPeriod(newPeriod)
	if dr.GetVotingPeriod() != newPeriod {
		t.Errorf("Expected voting period %v, got %v", newPeriod, dr.GetVotingPeriod())
	}
	
	// Reset
	dr.SetVotingPeriod(originalPeriod)
	
	// Test setting required quorum
	originalQuorum := dr.GetRequiredQuorum()
	newQuorum := int64(10)
	
	dr.SetRequiredQuorum(newQuorum)
	if dr.GetRequiredQuorum() != newQuorum {
		t.Errorf("Expected required quorum %d, got %d", newQuorum, dr.GetRequiredQuorum())
	}
	
	// Reset
	dr.SetRequiredQuorum(originalQuorum)
}

func TestDefaultDisputeResolver_EdgeCases(t *testing.T) {
	dr := NewDefaultDisputeResolver()
	
	// Test empty inputs
	_, err := dr.CreateDispute("", std.Address("g1234567890123456789012345678901234567890"), 
		std.Address("g1234567890123456789012345678901234567891"), "reason")
	if err == nil {
		t.Error("Expected error for empty name")
	}
	
	_, err = dr.CreateDispute("name", "", 
		std.Address("g1234567890123456789012345678901234567891"), "reason")
	if err == nil {
		t.Error("Expected error for empty owner")
	}
	
	_, err = dr.CreateDispute("name", std.Address("g1234567890123456789012345678901234567890"), 
		"", "reason")
	if err == nil {
		t.Error("Expected error for empty disputer")
	}
	
	_, err = dr.CreateDispute("name", std.Address("g1234567890123456789012345678901234567890"), 
		std.Address("g1234567890123456789012345678901234567891"), "")
	if err == nil {
		t.Error("Expected error for empty reason")
	}
	
	// Test self-dispute
	addr := std.Address("g1234567890123456789012345678901234567890")
	_, err = dr.CreateDispute("name", addr, addr, "reason")
	if err == nil {
		t.Error("Expected error for self-dispute")
	}
	
	// Test resolving non-existent dispute
	err = dr.ResolveDispute("nonexistent", true)
	if err != ErrDisputeNotFound {
		t.Errorf("Expected ErrDisputeNotFound, got %v", err)
	}
}
