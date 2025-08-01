package antisquatting

import (
	"std"
	"testing"
)

// MockRegistrationHandler for testing
type MockRegistrationHandler struct {
	registeredNames map[string]std.Address
	availableNames  map[string]bool
}

func NewMockRegistrationHandler() *MockRegistrationHandler {
	return &MockRegistrationHandler{
		registeredNames: make(map[string]std.Address),
		availableNames:  make(map[string]bool),
	}
}

func (m *MockRegistrationHandler) RegisterName(name string, owner std.Address, amount int64) error {
	if !m.IsNameAvailable(name) {
		return ErrNameNotAvailable
	}
	m.registeredNames[name] = owner
	m.availableNames[name] = false
	return nil
}

func (m *MockRegistrationHandler) IsNameAvailable(name string) bool {
	if available, exists := m.availableNames[name]; exists {
		return available
	}
	return true // Default to available
}

func (m *MockRegistrationHandler) GetNameOwner(name string) (std.Address, bool) {
	owner, exists := m.registeredNames[name]
	return owner, exists
}

func (m *MockRegistrationHandler) SetNameAvailable(name string, available bool) {
	m.availableNames[name] = available
}

func TestDefaultAntiSquattingSystem_Integration(t *testing.T) {
	// Create system with mock handler
	system := NewDefaultAntiSquattingSystem()
	mockHandler := NewMockRegistrationHandler()
	system.SetRegistrationHandler(mockHandler)
	
	// Test high-value name registration (should start auction)
	requester := std.Address("g1234567890123456789012345678901234567890")
	
	requiresAuction, err := system.RequestRegistration("admin", requester)
	if err != nil {
		t.Fatalf("RequestRegistration failed: %v", err)
	}
	
	if !requiresAuction {
		t.Error("High-value name should require auction")
	}
	
	// Verify auction was started
	status, err := system.GetAuctionStatus("admin")
	if err != nil {
		t.Fatalf("GetAuctionStatus failed: %v", err)
	}
	
	if status.Name != "admin" {
		t.Errorf("Expected auction for 'admin', got '%s'", status.Name)
	}
	
	if status.Phase != PhaseCommit {
		t.Errorf("Expected phase %v, got %v", PhaseCommit, status.Phase)
	}
}

func TestDefaultAntiSquattingSystem_DirectRegistration(t *testing.T) {
	// Create system with mock handler
	system := NewDefaultAntiSquattingSystem()
	mockHandler := NewMockRegistrationHandler()
	system.SetRegistrationHandler(mockHandler)
	
	// Test regular name registration (should register directly)
	requester := std.Address("g1234567890123456789012345678901234567890")
	
	requiresAuction, err := system.RequestRegistration("regularname", requester)
	if err != nil {
		t.Fatalf("RequestRegistration failed: %v", err)
	}
	
	if requiresAuction {
		t.Error("Regular name should not require auction")
	}
	
	// Verify name was registered
	owner, exists := mockHandler.GetNameOwner("regularname")
	if !exists {
		t.Error("Name should be registered")
	}
	
	if owner != requester {
		t.Errorf("Expected owner %s, got %s", requester, owner)
	}
}

func TestDefaultAntiSquattingSystem_CompleteAuctionFlow(t *testing.T) {
	// Create system with mock handler
	system := NewDefaultAntiSquattingSystem()
	mockHandler := NewMockRegistrationHandler()
	system.SetRegistrationHandler(mockHandler)
	
	name := "admin"
	
	// Start auction
	classifier := system.GetClassifier()
	err := system.StartAuction(name, classifier)
	if err != nil {
		t.Fatalf("StartAuction failed: %v", err)
	}
	
	// Add bidders
	bidder1 := std.Address("g1234567890123456789012345678901234567890")
	bidder2 := std.Address("g1234567890123456789012345678901234567891")
	bidder3 := std.Address("g1234567890123456789012345678901234567892")
	
	deposit := classifier.GetMinimumDeposit(name)
	
	// Submit bids
	hash1 := ComputeBidHash(3_000_000, "salt1", bidder1)
	hash2 := ComputeBidHash(5_000_000, "salt2", bidder2)
	hash3 := ComputeBidHash(7_000_000, "salt3", bidder3) // Winner
	
	err = system.SubmitBid(name, bidder1, hash1, deposit)
	if err != nil {
		t.Fatalf("SubmitBid failed: %v", err)
	}
	
	err = system.SubmitBid(name, bidder2, hash2, deposit)
	if err != nil {
		t.Fatalf("SubmitBid failed: %v", err)
	}
	
	err = system.SubmitBid(name, bidder3, hash3, deposit)
	if err != nil {
		t.Fatalf("SubmitBid failed: %v", err)
	}
	
	// Move to reveal phase (simulate time passing)
	auctionManager := system.GetAuctionManager()
	if am, ok := auctionManager.(*DefaultAuctionManager); ok {
		auction, _ := am.GetAuction(name)
		auction.Phase = PhaseReveal
		am.auctions.Set(name, auction)
	}
	
	// Reveal bids
	err = system.RevealBid(name, bidder1, 3_000_000, "salt1")
	if err != nil {
		t.Fatalf("RevealBid failed: %v", err)
	}
	
	err = system.RevealBid(name, bidder2, 5_000_000, "salt2")
	if err != nil {
		t.Fatalf("RevealBid failed: %v", err)
	}
	
	err = system.RevealBid(name, bidder3, 7_000_000, "salt3")
	if err != nil {
		t.Fatalf("RevealBid failed: %v", err)
	}
	
	// Finalize auction
	winner, winningBid, err := system.FinalizeAuction(name)
	if err != nil {
		t.Fatalf("FinalizeAuction failed: %v", err)
	}
	
	// Verify results
	if winner != bidder3 {
		t.Errorf("Expected winner %s, got %s", bidder3, winner)
	}
	
	if winningBid != 7_000_000 {
		t.Errorf("Expected winning bid 7000000, got %d", winningBid)
	}
	
	// Verify name was registered to winner
	owner, exists := mockHandler.GetNameOwner(name)
	if !exists {
		t.Error("Name should be registered after auction")
	}
	
	if owner != bidder3 {
		t.Errorf("Expected owner %s, got %s", bidder3, owner)
	}
}

func TestDefaultAntiSquattingSystem_DisputeFlow(t *testing.T) {
	// Create system with mock handler
	system := NewDefaultAntiSquattingSystem()
	mockHandler := NewMockRegistrationHandler()
	system.SetRegistrationHandler(mockHandler)
	
	// Register a name first
	name := "trademark"
	owner := std.Address("g1234567890123456789012345678901234567890")
	mockHandler.RegisterName(name, owner, 1_000_000)
	
	// Create dispute
	disputer := std.Address("g1234567890123456789012345678901234567891")
	reason := "This name violates my trademark"
	
	disputeID, err := system.CreateDispute(name, owner, disputer, reason)
	if err != nil {
		t.Fatalf("CreateDispute failed: %v", err)
	}
	
	if disputeID == "" {
		t.Error("Expected non-empty dispute ID")
	}
	
	// Check dispute status
	status, err := system.GetDisputeStatus(disputeID)
	if err != nil {
		t.Fatalf("GetDisputeStatus failed: %v", err)
	}
	
	if status != DisputePending {
		t.Errorf("Expected status %v, got %v", DisputePending, status)
	}
	
	// Resolve dispute (approve)
	err = system.ResolveDispute(disputeID, true)
	if err != nil {
		t.Fatalf("ResolveDispute failed: %v", err)
	}
	
	// Check final status
	status, err = system.GetDisputeStatus(disputeID)
	if err != nil {
		t.Fatalf("GetDisputeStatus failed: %v", err)
	}
	
	if status != DisputeApproved {
		t.Errorf("Expected status %v, got %v", DisputeApproved, status)
	}
}

func TestDefaultAntiSquattingSystem_GetNameRegistrationInfo(t *testing.T) {
	// Create system with mock handler
	system := NewDefaultAntiSquattingSystem()
	mockHandler := NewMockRegistrationHandler()
	system.SetRegistrationHandler(mockHandler)
	
	// Test high-value name info
	info := system.GetNameRegistrationInfo("admin")
	
	if !info.Available {
		t.Error("Name should be available initially")
	}
	
	if !info.RequiresAuction {
		t.Error("High-value name should require auction")
	}
	
	if info.MinimumBid <= 0 {
		t.Error("Minimum bid should be positive")
	}
	
	if info.MinimumDeposit <= 0 {
		t.Error("Minimum deposit should be positive")
	}
	
	// Test regular name info
	info2 := system.GetNameRegistrationInfo("regularname")
	
	if !info2.Available {
		t.Error("Name should be available initially")
	}
	
	if info2.RequiresAuction {
		t.Error("Regular name should not require auction")
	}
	
	if info2.DirectPrice <= 0 {
		t.Error("Direct price should be positive")
	}
	
	// Test unavailable name
	mockHandler.SetNameAvailable("taken", false)
	mockHandler.RegisterName("taken", std.Address("g1234567890123456789012345678901234567890"), 1000)
	
	info3 := system.GetNameRegistrationInfo("taken")
	
	if info3.Available {
		t.Error("Taken name should not be available")
	}
	
	if info3.CurrentOwner == "" {
		t.Error("Should have current owner for taken name")
	}
}

func TestDefaultAntiSquattingSystem_SystemStatus(t *testing.T) {
	// Create system with mock handler
	system := NewDefaultAntiSquattingSystem()
	mockHandler := NewMockRegistrationHandler()
	system.SetRegistrationHandler(mockHandler)
	
	// Initial status
	status := system.GetSystemStatus()
	
	if status.ActiveAuctions != 0 {
		t.Errorf("Expected 0 active auctions, got %d", status.ActiveAuctions)
	}
	
	if status.PendingDisputes != 0 {
		t.Errorf("Expected 0 pending disputes, got %d", status.PendingDisputes)
	}
	
	// Start an auction
	classifier := system.GetClassifier()
	system.StartAuction("admin", classifier)
	
	// Create a dispute
	owner := std.Address("g1234567890123456789012345678901234567890")
	disputer := std.Address("g1234567890123456789012345678901234567891")
	mockHandler.RegisterName("disputed", owner, 1000)
	system.CreateDispute("disputed", owner, disputer, "test")
	
	// Check updated status
	status = system.GetSystemStatus()
	
	if status.ActiveAuctions != 1 {
		t.Errorf("Expected 1 active auction, got %d", status.ActiveAuctions)
	}
	
	if status.PendingDisputes != 1 {
		t.Errorf("Expected 1 pending dispute, got %d", status.PendingDisputes)
	}
}

func TestDefaultAntiSquattingSystem_Configuration(t *testing.T) {
	system := NewDefaultAntiSquattingSystem()
	
	// Test enabling/disabling
	if !system.IsEnabled() {
		t.Error("System should be enabled by default")
	}
	
	system.SetEnabled(false)
	if system.IsEnabled() {
		t.Error("System should be disabled after SetEnabled(false)")
	}
	
	system.SetEnabled(true)
	if !system.IsEnabled() {
		t.Error("System should be enabled after SetEnabled(true)")
	}
	
	// Test component access
	if system.GetClassifier() == nil {
		t.Error("Should have classifier")
	}
	
	if system.GetAuctionManager() == nil {
		t.Error("Should have auction manager")
	}
	
	if system.GetDisputeResolver() == nil {
		t.Error("Should have dispute resolver")
	}
	
	// Test validation
	err := system.ValidateConfiguration()
	if err != nil {
		t.Errorf("Configuration should be valid: %v", err)
	}
	
	// Test with missing registration handler
	system.SetRegistrationHandler(nil)
	err = system.ValidateConfiguration()
	if err == nil {
		t.Error("Should fail validation without registration handler")
	}
}
