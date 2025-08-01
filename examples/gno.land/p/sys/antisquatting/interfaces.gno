// Package antisquatting provides a generic anti-squatting system for naming systems.
// It implements sealed-bid auctions and dispute resolution mechanisms to prevent
// valuable name squatting while remaining pluggable and reusable.
package antisquatting

import (
	"std"
	"time"
)

// NameClassifier determines if a name requires anti-squatting protection
// and provides auction parameters for high-value names.
type NameClassifier interface {
	// IsHighValue returns true if the name requires auction-based registration
	IsHighValue(name string) bool
	
	// GetMinimumBid returns the minimum bid amount for the name
	GetMinimumBid(name string) int64
	
	// GetAuctionDuration returns how long the commit phase should last
	GetAuctionDuration(name string) time.Duration
	
	// GetRevealDuration returns how long the reveal phase should last
	GetRevealDuration(name string) time.Duration
	
	// GetMinimumDeposit returns the minimum deposit required to participate
	GetMinimumDeposit(name string) int64
}

// DisputeResolver handles disputes for registered names through DAO governance.
type DisputeResolver interface {
	// CanDispute returns true if the disputer can file a dispute for this name
	CanDispute(name string, disputer std.Address) bool
	
	// CreateDispute files a new dispute with the given reason
	CreateDispute(name string, owner std.Address, disputer std.Address, reason string) (disputeID string, err error)
	
	// GetDisputeStatus returns the current status of a dispute
	GetDisputeStatus(disputeID string) (DisputeStatus, error)
	
	// ResolveDispute executes the resolution of a dispute (called by DAO)
	ResolveDispute(disputeID string, approved bool) error
}

// AuctionManager handles the sealed-bid auction process for high-value names.
type AuctionManager interface {
	// StartAuction begins a new auction for the given name
	StartAuction(name string, classifier NameClassifier) error
	
	// SubmitBid allows a user to submit a commitment for their bid
	SubmitBid(name string, bidder std.Address, bidHash string, deposit int64) error
	
	// RevealBid allows a user to reveal their actual bid amount and salt
	RevealBid(name string, bidder std.Address, amount int64, salt string) error
	
	// FinalizeAuction determines the winner and completes the auction
	FinalizeAuction(name string) (winner std.Address, winningBid int64, err error)
	
	// GetAuctionStatus returns the current status of an auction
	GetAuctionStatus(name string) (AuctionStatus, error)
	
	// RefundLosers returns deposits to non-winning bidders
	RefundLosers(name string) error
}

// RegistrationHandler integrates with the actual naming system to register names.
type RegistrationHandler interface {
	// RegisterName registers the name to the winner after auction completion
	RegisterName(name string, owner std.Address, amount int64) error
	
	// IsNameAvailable checks if a name is available for registration
	IsNameAvailable(name string) bool
	
	// GetNameOwner returns the current owner of a name (if any)
	GetNameOwner(name string) (std.Address, bool)
}

// AuctionPhase represents the current phase of an auction
type AuctionPhase int

const (
	PhaseNotStarted AuctionPhase = iota
	PhaseCommit                  // Users submit bid commitments
	PhaseReveal                  // Users reveal their actual bids
	PhaseFinalized               // Auction completed, winner determined
	PhaseCancelled               // Auction was cancelled
)

// AuctionStatus contains the current state of an auction
type AuctionStatus struct {
	Name           string
	Phase          AuctionPhase
	CommitDeadline time.Time
	RevealDeadline time.Time
	MinimumBid     int64
	MinimumDeposit int64
	BidCount       int
	HighestBid     int64 // Only visible after reveal phase
	Winner         std.Address
}

// DisputeStatus represents the current state of a dispute
type DisputeStatus int

const (
	DisputePending DisputeStatus = iota
	DisputeApproved
	DisputeRejected
	DisputeExpired
)

// Dispute contains information about a name dispute
type Dispute struct {
	ID       string
	Name     string
	Owner    std.Address
	Disputer std.Address
	Reason   string
	Status   DisputeStatus
	Created  time.Time
	Resolved time.Time
}

// BidCommitment represents a committed bid in the auction
type BidCommitment struct {
	Bidder    std.Address
	Hash      string    // hash(amount + salt + bidder)
	Deposit   int64     // Deposit paid with commitment
	Timestamp time.Time // When commitment was made
	Revealed  bool      // Whether bid has been revealed
	Amount    int64     // Actual bid amount (set during reveal)
	Salt      string    // Salt used for commitment (set during reveal)
}

// Auction represents an active auction for a name
type Auction struct {
	Name           string
	Phase          AuctionPhase
	CommitDeadline time.Time
	RevealDeadline time.Time
	MinimumBid     int64
	MinimumDeposit int64
	Commitments    []BidCommitment
	Winner         std.Address
	WinningBid     int64
	Created        time.Time
	Finalized      time.Time
}

// AntiSquattingSystem is the main interface that combines all functionality
type AntiSquattingSystem interface {
	// Core registration flow
	RequestRegistration(name string, requester std.Address) (requiresAuction bool, err error)
	
	// Auction management
	AuctionManager
	
	// Dispute handling
	DisputeResolver
	
	// Configuration
	SetNameClassifier(classifier NameClassifier)
	SetRegistrationHandler(handler RegistrationHandler)
	
	// Status queries
	GetSystemStatus() SystemStatus
}

// SystemStatus provides overall system statistics
type SystemStatus struct {
	ActiveAuctions   int
	PendingDisputes  int
	TotalRegistered  int
	TotalAuctioned   int
	TotalDisputed    int
}

// Error types for the anti-squatting system
type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrNameNotAvailable     Error = "name is not available"
	ErrAuctionNotFound      Error = "auction not found"
	ErrAuctionNotActive     Error = "auction is not active"
	ErrInvalidPhase         Error = "invalid auction phase"
	ErrBidTooLow           Error = "bid amount too low"
	ErrDepositTooLow       Error = "deposit amount too low"
	ErrInvalidCommitment   Error = "invalid bid commitment"
	ErrAlreadyCommitted    Error = "already committed to this auction"
	ErrNotCommitted        Error = "no commitment found"
	ErrRevealMismatch      Error = "revealed bid doesn't match commitment"
	ErrDisputeNotFound     Error = "dispute not found"
	ErrCannotDispute       Error = "cannot dispute this name"
	ErrInvalidDisputer     Error = "invalid disputer"
	ErrSystemNotConfigured Error = "anti-squatting system not properly configured"
)
