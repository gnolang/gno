package antisquatting

import (
	"std"
	"time"
)

// DefaultAntiSquattingSystem implements AntiSquattingSystem
type DefaultAntiSquattingSystem struct {
	// Core components
	classifier         NameClassifier
	auctionManager     AuctionManager
	disputeResolver    DisputeResolver
	registrationHandler RegistrationHandler
	
	// Configuration
	enabled bool
}

// NewDefaultAntiSquattingSystem creates a new anti-squatting system
func NewDefaultAntiSquattingSystem() *DefaultAntiSquattingSystem {
	return &DefaultAntiSquattingSystem{
		classifier:      NewDefaultNameClassifier(),
		auctionManager:  NewDefaultAuctionManager(),
		disputeResolver: NewDefaultDisputeResolver(),
		enabled:         true,
	}
}

// RequestRegistration handles a name registration request
func (ass *DefaultAntiSquattingSystem) RequestRegistration(name string, requester std.Address) (bool, error) {
	// Check if system is enabled
	if !ass.enabled {
		return false, ErrSystemNotConfigured
	}
	
	// Validate components are configured
	if ass.classifier == nil || ass.registrationHandler == nil {
		return false, ErrSystemNotConfigured
	}
	
	// Check if name is available
	if !ass.registrationHandler.IsNameAvailable(name) {
		return false, ErrNameNotAvailable
	}
	
	// Check if name requires auction
	if ass.classifier.IsHighValue(name) {
		// Start auction if one doesn't exist
		if err := ass.auctionManager.StartAuction(name, ass.classifier); err != nil {
			// If auction already exists, that's fine
			if err != ErrAuctionNotActive {
				return false, err
			}
		}
		return true, nil // Requires auction
	}
	
	// For non-high-value names, register directly
	// This would typically involve payment processing
	if err := ass.registrationHandler.RegisterName(name, requester, ass.classifier.GetMinimumBid(name)); err != nil {
		return false, err
	}
	
	return false, nil // Direct registration completed
}

// AuctionManager interface implementation

func (ass *DefaultAntiSquattingSystem) StartAuction(name string, classifier NameClassifier) error {
	return ass.auctionManager.StartAuction(name, classifier)
}

func (ass *DefaultAntiSquattingSystem) SubmitBid(name string, bidder std.Address, bidHash string, deposit int64) error {
	return ass.auctionManager.SubmitBid(name, bidder, bidHash, deposit)
}

func (ass *DefaultAntiSquattingSystem) RevealBid(name string, bidder std.Address, amount int64, salt string) error {
	return ass.auctionManager.RevealBid(name, bidder, amount, salt)
}

func (ass *DefaultAntiSquattingSystem) FinalizeAuction(name string) (std.Address, int64, error) {
	winner, amount, err := ass.auctionManager.FinalizeAuction(name)
	if err != nil {
		return "", 0, err
	}
	
	// Register the name to the winner
	if winner != "" && ass.registrationHandler != nil {
		if err := ass.registrationHandler.RegisterName(name, winner, amount); err != nil {
			return "", 0, err
		}
	}
	
	// Refund losing bidders
	if err := ass.auctionManager.RefundLosers(name); err != nil {
		// Log error but don't fail the finalization
		// In a real implementation, this would be handled more robustly
	}
	
	return winner, amount, nil
}

func (ass *DefaultAntiSquattingSystem) GetAuctionStatus(name string) (AuctionStatus, error) {
	return ass.auctionManager.GetAuctionStatus(name)
}

func (ass *DefaultAntiSquattingSystem) RefundLosers(name string) error {
	return ass.auctionManager.RefundLosers(name)
}

// DisputeResolver interface implementation

func (ass *DefaultAntiSquattingSystem) CanDispute(name string, disputer std.Address) bool {
	if ass.disputeResolver == nil {
		return false
	}
	return ass.disputeResolver.CanDispute(name, disputer)
}

func (ass *DefaultAntiSquattingSystem) CreateDispute(name string, owner std.Address, disputer std.Address, reason string) (string, error) {
	if ass.disputeResolver == nil {
		return "", ErrSystemNotConfigured
	}
	return ass.disputeResolver.CreateDispute(name, owner, disputer, reason)
}

func (ass *DefaultAntiSquattingSystem) GetDisputeStatus(disputeID string) (DisputeStatus, error) {
	if ass.disputeResolver == nil {
		return DisputePending, ErrSystemNotConfigured
	}
	return ass.disputeResolver.GetDisputeStatus(disputeID)
}

func (ass *DefaultAntiSquattingSystem) ResolveDispute(disputeID string, approved bool) error {
	if ass.disputeResolver == nil {
		return ErrSystemNotConfigured
	}
	return ass.disputeResolver.ResolveDispute(disputeID, approved)
}

// Configuration methods

func (ass *DefaultAntiSquattingSystem) SetNameClassifier(classifier NameClassifier) {
	ass.classifier = classifier
}

func (ass *DefaultAntiSquattingSystem) SetRegistrationHandler(handler RegistrationHandler) {
	ass.registrationHandler = handler
}

func (ass *DefaultAntiSquattingSystem) SetAuctionManager(manager AuctionManager) {
	ass.auctionManager = manager
}

func (ass *DefaultAntiSquattingSystem) SetDisputeResolver(resolver DisputeResolver) {
	ass.disputeResolver = resolver
}

func (ass *DefaultAntiSquattingSystem) SetEnabled(enabled bool) {
	ass.enabled = enabled
}

// Status and information methods

func (ass *DefaultAntiSquattingSystem) GetSystemStatus() SystemStatus {
	status := SystemStatus{}
	
	// Count active auctions
	if am, ok := ass.auctionManager.(*DefaultAuctionManager); ok {
		activeAuctions := am.GetActiveAuctions()
		status.ActiveAuctions = len(activeAuctions)
	}
	
	// Count pending disputes
	if dr, ok := ass.disputeResolver.(*DefaultDisputeResolver); ok {
		activeDisputes := dr.GetActiveDisputes()
		status.PendingDisputes = len(activeDisputes)
		
		// Get dispute statistics
		stats := dr.GetDisputeStatistics()
		status.TotalDisputed = stats["total"]
	}
	
	return status
}

func (ass *DefaultAntiSquattingSystem) IsEnabled() bool {
	return ass.enabled
}

func (ass *DefaultAntiSquattingSystem) GetClassifier() NameClassifier {
	return ass.classifier
}

func (ass *DefaultAntiSquattingSystem) GetAuctionManager() AuctionManager {
	return ass.auctionManager
}

func (ass *DefaultAntiSquattingSystem) GetDisputeResolver() DisputeResolver {
	return ass.disputeResolver
}

func (ass *DefaultAntiSquattingSystem) GetRegistrationHandler() RegistrationHandler {
	return ass.registrationHandler
}

// Utility methods for external use

// IsHighValueName checks if a name requires auction-based registration
func (ass *DefaultAntiSquattingSystem) IsHighValueName(name string) bool {
	if ass.classifier == nil {
		return false
	}
	return ass.classifier.IsHighValue(name)
}

// GetNameRegistrationInfo returns information about how to register a name
func (ass *DefaultAntiSquattingSystem) GetNameRegistrationInfo(name string) NameRegistrationInfo {
	info := NameRegistrationInfo{
		Name:      name,
		Available: false,
	}
	
	// Check if registration handler is configured
	if ass.registrationHandler == nil {
		info.Error = "System not configured"
		return info
	}
	
	// Check availability
	info.Available = ass.registrationHandler.IsNameAvailable(name)
	if !info.Available {
		if owner, exists := ass.registrationHandler.GetNameOwner(name); exists {
			info.CurrentOwner = owner
		}
		return info
	}
	
	// Check if high-value (requires auction)
	if ass.classifier != nil && ass.classifier.IsHighValue(name) {
		info.RequiresAuction = true
		info.MinimumBid = ass.classifier.GetMinimumBid(name)
		info.MinimumDeposit = ass.classifier.GetMinimumDeposit(name)
		info.CommitDuration = ass.classifier.GetAuctionDuration(name)
		info.RevealDuration = ass.classifier.GetRevealDuration(name)
		
		// Check if auction already exists
		if status, err := ass.auctionManager.GetAuctionStatus(name); err == nil {
			info.AuctionStatus = &status
		}
	} else {
		info.RequiresAuction = false
		if ass.classifier != nil {
			info.DirectPrice = ass.classifier.GetMinimumBid(name)
		}
	}
	
	return info
}

// NameRegistrationInfo contains information about registering a name
type NameRegistrationInfo struct {
	Name             string
	Available        bool
	CurrentOwner     std.Address
	RequiresAuction  bool
	MinimumBid       int64
	MinimumDeposit   int64
	DirectPrice      int64
	CommitDuration   time.Duration
	RevealDuration   time.Duration
	AuctionStatus    *AuctionStatus
	Error            string
}

// ValidateConfiguration checks if the system is properly configured
func (ass *DefaultAntiSquattingSystem) ValidateConfiguration() error {
	if ass.classifier == nil {
		return ErrSystemNotConfigured
	}
	if ass.auctionManager == nil {
		return ErrSystemNotConfigured
	}
	if ass.disputeResolver == nil {
		return ErrSystemNotConfigured
	}
	if ass.registrationHandler == nil {
		return ErrSystemNotConfigured
	}
	return nil
}

// ProcessExpiredAuctions processes any auctions that have expired
func (ass *DefaultAntiSquattingSystem) ProcessExpiredAuctions() error {
	if am, ok := ass.auctionManager.(*DefaultAuctionManager); ok {
		activeAuctions := am.GetActiveAuctions()
		
		for _, name := range activeAuctions {
			status, err := am.GetAuctionStatus(name)
			if err != nil {
				continue
			}
			
			// Check if reveal phase has ended
			if status.Phase == PhaseReveal && time.Now().After(status.RevealDeadline) {
				// Finalize the auction
				_, _, err := ass.FinalizeAuction(name)
				if err != nil {
					// Log error but continue processing other auctions
					continue
				}
			}
		}
	}
	
	return nil
}
