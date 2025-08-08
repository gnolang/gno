package users

import (
	"std"
	"time"
	
	"gno.land/p/sys/antisquatting"
)

// Anti-squatting system integration for r/sys/users
var (
	// Global anti-squatting system instance
	antiSquattingSystem antisquatting.AntiSquattingSystem
	
	// Configuration
	antiSquattingEnabled = false
)

// SysUsersRegistrationHandler implements antisquatting.RegistrationHandler
// This bridges the anti-squatting system with the existing r/sys/users functionality
type SysUsersRegistrationHandler struct{}

// RegisterName registers a name to an owner after auction completion
func (h *SysUsersRegistrationHandler) RegisterName(name string, owner std.Address, amount int64) error {
	// Use the existing RegisterUser function but bypass anti-squatting checks
	return registerUserDirect(name, owner)
}

// IsNameAvailable checks if a name is available for registration
func (h *SysUsersRegistrationHandler) IsNameAvailable(name string) bool {
	// Check basic validation first
	if err := validateName(name); err != nil {
		return false
	}
	
	// Check if name is already taken
	return !nameStore.Has(name)
}

// GetNameOwner returns the current owner of a name (if any)
func (h *SysUsersRegistrationHandler) GetNameOwner(name string) (std.Address, bool) {
	raw, exists := nameStore.Get(name)
	if !exists {
		return "", false
	}
	
	userData := raw.(*UserData)
	if userData.IsDeleted() {
		return "", false
	}
	
	return userData.Addr(), true
}

// Enhanced RegisterUser function that uses anti-squatting system
func RegisterUserWithAntiSquatting(cur realm, name string, address std.Address) error {
	// Validate caller
	if !controllers.Has(std.PreviousRealm().Address()) {
		return NewErrNotWhitelisted()
	}
	
	// If anti-squatting is disabled, use original registration
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		return RegisterUser(cur, name, address)
	}
	
	// Use anti-squatting system for registration
	requiresAuction, err := antiSquattingSystem.RequestRegistration(name, address)
	if err != nil {
		return err
	}
	
	if requiresAuction {
		// Name requires auction - return specific error
		return ErrNameRequiresAuction
	}
	
	// Direct registration was completed by the anti-squatting system
	return nil
}

// registerUserDirect performs direct registration bypassing anti-squatting checks
// This is used internally by the anti-squatting system
func registerUserDirect(name string, address std.Address) error {
	// Validate name
	if err := validateName(name); err != nil {
		return err
	}
	
	// Validate address
	if !address.IsValid() {
		return ErrInvalidAddress
	}
	
	// Check if name is taken
	if nameStore.Has(name) {
		return ErrNameTaken
	}
	
	raw, ok := addressStore.Get(address.String())
	if ok {
		// Cannot re-register after deletion
		if raw.(*UserData).IsDeleted() {
			return ErrDeletedUser
		}
		
		// For a second name, use UpdateName
		return ErrAlreadyHasName
	}
	
	// Create UserData
	data := &UserData{
		addr:     address,
		username: name,
		deleted:  false,
	}
	
	// Set corresponding stores
	nameStore.Set(name, data)
	addressStore.Set(address.String(), data)
	
	std.Emit(RegisterUserEvent,
		"name", name,
		"address", address.String(),
	)
	return nil
}

// Auction management functions

// StartNameAuction starts an auction for a high-value name
func StartNameAuction(cur realm, name string) error {
	// Validate caller
	if !controllers.Has(std.PreviousRealm().Address()) {
		return NewErrNotWhitelisted()
	}
	
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		return ErrAntiSquattingDisabled
	}
	
	// Check if name is available
	handler := &SysUsersRegistrationHandler{}
	if !handler.IsNameAvailable(name) {
		return ErrNameNotAvailable
	}
	
	// Get classifier to determine auction parameters
	classifier := antiSquattingSystem.GetClassifier()
	if classifier == nil {
		return ErrSystemNotConfigured
	}
	
	// Start the auction
	return antiSquattingSystem.StartAuction(name, classifier)
}

// SubmitBid allows a user to submit a bid commitment for a name auction
func SubmitBid(cur realm, name string, bidHash string) error {
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		return ErrAntiSquattingDisabled
	}
	
	// Get the bidder address
	bidder := std.PreviousRealm().Address()
	if !bidder.IsValid() {
		return ErrInvalidAddress
	}
	
	// Get the required deposit amount
	classifier := antiSquattingSystem.GetClassifier()
	if classifier == nil {
		return ErrSystemNotConfigured
	}
	
	deposit := classifier.GetMinimumDeposit(name)
	
	// Check if user sent enough deposit
	sent := std.OriginSend()
	if sent.AmountOf("ugnot") < deposit {
		return ErrInsufficientDeposit
	}
	
	// Submit the bid
	return antiSquattingSystem.SubmitBid(name, bidder, bidHash, deposit)
}

// RevealBid allows a user to reveal their bid amount and salt
func RevealBid(cur realm, name string, amount int64, salt string) error {
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		return ErrAntiSquattingDisabled
	}
	
	// Get the bidder address
	bidder := std.PreviousRealm().Address()
	if !bidder.IsValid() {
		return ErrInvalidAddress
	}
	
	// Reveal the bid
	return antiSquattingSystem.RevealBid(name, bidder, amount, salt)
}

// FinalizeAuction completes an auction and registers the name to the winner
func FinalizeAuction(cur realm, name string) error {
	// Validate caller (allow anyone to finalize expired auctions)
	if !controllers.Has(std.PreviousRealm().Address()) {
		return NewErrNotWhitelisted()
	}
	
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		return ErrAntiSquattingDisabled
	}
	
	// Finalize the auction
	winner, amount, err := antiSquattingSystem.FinalizeAuction(name)
	if err != nil {
		return err
	}
	
	// Emit auction completion event
	std.Emit("AuctionFinalized",
		"name", name,
		"winner", winner.String(),
		"amount", string(amount),
	)
	
	return nil
}

// GetAuctionStatus returns the current status of a name auction
func GetAuctionStatus(name string) (antisquatting.AuctionStatus, error) {
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		return antisquatting.AuctionStatus{}, ErrAntiSquattingDisabled
	}
	
	return antiSquattingSystem.GetAuctionStatus(name)
}

// Dispute management functions

// CreateDispute files a dispute for a registered name
func CreateDispute(cur realm, name string, reason string) (string, error) {
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		return "", ErrAntiSquattingDisabled
	}
	
	// Get disputer address
	disputer := std.PreviousRealm().Address()
	if !disputer.IsValid() {
		return "", ErrInvalidAddress
	}
	
	// Get current owner
	handler := &SysUsersRegistrationHandler{}
	owner, exists := handler.GetNameOwner(name)
	if !exists {
		return "", ErrNameNotFound
	}
	
	// Check dispute fee
	disputeResolver := antiSquattingSystem.GetDisputeResolver()
	if dr, ok := disputeResolver.(*antisquatting.DefaultDisputeResolver); ok {
		disputeFee := dr.GetDisputeFee()
		sent := std.OriginSend()
		if sent.AmountOf("ugnot") < disputeFee {
			return "", ErrInsufficientDisputeFee
		}
	}
	
	// Create the dispute
	return antiSquattingSystem.CreateDispute(name, owner, disputer, reason)
}

// GetDisputeStatus returns the status of a dispute
func GetDisputeStatus(disputeID string) (antisquatting.DisputeStatus, error) {
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		return antisquatting.DisputePending, ErrAntiSquattingDisabled
	}
	
	return antiSquattingSystem.GetDisputeStatus(disputeID)
}

// Configuration and management functions

// InitializeAntiSquatting sets up the anti-squatting system
func InitializeAntiSquatting() error {
	// Only allow initialization by controllers
	if !controllers.Has(std.PreviousRealm().Address()) {
		return NewErrNotWhitelisted()
	}
	
	// Create the anti-squatting system
	system := antisquatting.NewDefaultAntiSquattingSystem()
	
	// Set up the registration handler
	handler := &SysUsersRegistrationHandler{}
	system.SetRegistrationHandler(handler)
	
	// Configure DAO integration for disputes
	daoResolver := antisquatting.NewDAODisputeResolver("gno.land/r/gov/dao")
	system.SetDisputeResolver(daoResolver)
	
	// Store the system globally
	antiSquattingSystem = system
	antiSquattingEnabled = true
	
	return nil
}

// EnableAntiSquatting enables or disables the anti-squatting system
func EnableAntiSquatting(enabled bool) error {
	// Only allow configuration by controllers
	if !controllers.Has(std.PreviousRealm().Address()) {
		return NewErrNotWhitelisted()
	}
	
	antiSquattingEnabled = enabled
	return nil
}

// IsAntiSquattingEnabled returns whether anti-squatting is enabled
func IsAntiSquattingEnabled() bool {
	return antiSquattingEnabled && antiSquattingSystem != nil
}

// GetAntiSquattingSystemStatus returns the current system status
func GetAntiSquattingSystemStatus() antisquatting.SystemStatus {
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		return antisquatting.SystemStatus{}
	}
	
	return antiSquattingSystem.GetSystemStatus()
}

// IsHighValueName checks if a name requires auction-based registration
func IsHighValueName(name string) bool {
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		return false
	}
	
	return antiSquattingSystem.IsHighValueName(name)
}

// GetNameRegistrationInfo returns information about registering a name
func GetNameRegistrationInfo(name string) antisquatting.NameRegistrationInfo {
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		// Return basic info for direct registration
		handler := &SysUsersRegistrationHandler{}
		return antisquatting.NameRegistrationInfo{
			Name:            name,
			Available:       handler.IsNameAvailable(name),
			RequiresAuction: false,
			DirectPrice:     0, // Free registration when anti-squatting is disabled
		}
	}
	
	return antiSquattingSystem.GetNameRegistrationInfo(name)
}

// ProcessExpiredAuctions processes any auctions that have expired
func ProcessExpiredAuctions() error {
	// Only allow processing by controllers
	if !controllers.Has(std.PreviousRealm().Address()) {
		return NewErrNotWhitelisted()
	}
	
	if !antiSquattingEnabled || antiSquattingSystem == nil {
		return ErrAntiSquattingDisabled
	}
	
	return antiSquattingSystem.ProcessExpiredAuctions()
}
