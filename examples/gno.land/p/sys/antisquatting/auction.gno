package antisquatting

import (
	"crypto/sha256"
	"encoding/hex"
	"std"
	"strconv"
	"time"
	
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

// DefaultAuctionManager implements AuctionManager with commit-reveal auctions
type DefaultAuctionManager struct {
	auctions *avl.Tree // name -> *Auction
	deposits *avl.Tree // bidder_name -> int64 (total deposits)
}

// NewDefaultAuctionManager creates a new auction manager
func NewDefaultAuctionManager() *DefaultAuctionManager {
	return &DefaultAuctionManager{
		auctions: avl.NewTree(),
		deposits: avl.NewTree(),
	}
}

// StartAuction begins a new auction for the given name
func (am *DefaultAuctionManager) StartAuction(name string, classifier NameClassifier) error {
	// Check if auction already exists
	if _, exists := am.auctions.Get(name); exists {
		return ErrAuctionNotActive // Auction already exists
	}
	
	// Get auction parameters from classifier
	commitDuration := classifier.GetAuctionDuration(name)
	revealDuration := classifier.GetRevealDuration(name)
	minimumBid := classifier.GetMinimumBid(name)
	minimumDeposit := classifier.GetMinimumDeposit(name)
	
	now := time.Now()
	auction := &Auction{
		Name:           name,
		Phase:          PhaseCommit,
		CommitDeadline: now.Add(commitDuration),
		RevealDeadline: now.Add(commitDuration + revealDuration),
		MinimumBid:     minimumBid,
		MinimumDeposit: minimumDeposit,
		Commitments:    make([]BidCommitment, 0),
		Created:        now,
	}
	
	am.auctions.Set(name, auction)
	return nil
}

// SubmitBid allows a user to submit a commitment for their bid
func (am *DefaultAuctionManager) SubmitBid(name string, bidder std.Address, bidHash string, deposit int64) error {
	// Get auction
	auctionVal, exists := am.auctions.Get(name)
	if !exists {
		return ErrAuctionNotFound
	}
	
	auction := auctionVal.(*Auction)
	
	// Check auction phase
	if auction.Phase != PhaseCommit {
		return ErrInvalidPhase
	}
	
	// Check if commit deadline has passed
	if time.Now().After(auction.CommitDeadline) {
		return ErrInvalidPhase
	}
	
	// Check minimum deposit
	if deposit < auction.MinimumDeposit {
		return ErrDepositTooLow
	}
	
	// Check if bidder already committed
	for _, commitment := range auction.Commitments {
		if commitment.Bidder == bidder {
			return ErrAlreadyCommitted
		}
	}
	
	// Validate hash format (should be hex string)
	if len(bidHash) != 64 { // SHA256 produces 64 hex characters
		return ErrInvalidCommitment
	}
	if _, err := hex.DecodeString(bidHash); err != nil {
		return ErrInvalidCommitment
	}
	
	// Create commitment
	commitment := BidCommitment{
		Bidder:    bidder,
		Hash:      bidHash,
		Deposit:   deposit,
		Timestamp: time.Now(),
		Revealed:  false,
	}
	
	// Add commitment to auction
	auction.Commitments = append(auction.Commitments, commitment)
	
	// Track total deposits for this bidder
	depositKey := bidder.String() + "_" + name
	am.deposits.Set(depositKey, deposit)
	
	// Update auction in storage
	am.auctions.Set(name, auction)
	
	return nil
}

// RevealBid allows a user to reveal their actual bid amount and salt
func (am *DefaultAuctionManager) RevealBid(name string, bidder std.Address, amount int64, salt string) error {
	// Get auction
	auctionVal, exists := am.auctions.Get(name)
	if !exists {
		return ErrAuctionNotFound
	}
	
	auction := auctionVal.(*Auction)
	
	// Check auction phase (can reveal during reveal phase or after commit phase ends)
	now := time.Now()
	if auction.Phase == PhaseCommit && now.Before(auction.CommitDeadline) {
		return ErrInvalidPhase
	}
	
	// Transition to reveal phase if commit phase just ended
	if auction.Phase == PhaseCommit && now.After(auction.CommitDeadline) {
		auction.Phase = PhaseReveal
	}
	
	// Check if reveal deadline has passed
	if now.After(auction.RevealDeadline) {
		return ErrInvalidPhase
	}
	
	// Find bidder's commitment
	var commitmentIndex = -1
	for i, commitment := range auction.Commitments {
		if commitment.Bidder == bidder {
			commitmentIndex = i
			break
		}
	}
	
	if commitmentIndex == -1 {
		return ErrNotCommitted
	}
	
	commitment := &auction.Commitments[commitmentIndex]
	
	// Check if already revealed
	if commitment.Revealed {
		return ErrAlreadyCommitted // Already revealed
	}
	
	// Check minimum bid
	if amount < auction.MinimumBid {
		return ErrBidTooLow
	}
	
	// Verify commitment hash
	expectedHash := am.computeBidHash(amount, salt, bidder)
	if expectedHash != commitment.Hash {
		return ErrRevealMismatch
	}
	
	// Update commitment with revealed information
	commitment.Amount = amount
	commitment.Salt = salt
	commitment.Revealed = true
	
	// Update auction in storage
	am.auctions.Set(name, auction)
	
	return nil
}

// FinalizeAuction determines the winner and completes the auction
func (am *DefaultAuctionManager) FinalizeAuction(name string) (std.Address, int64, error) {
	// Get auction
	auctionVal, exists := am.auctions.Get(name)
	if !exists {
		return "", 0, ErrAuctionNotFound
	}
	
	auction := auctionVal.(*Auction)
	
	// Check if reveal phase has ended
	if time.Now().Before(auction.RevealDeadline) {
		return "", 0, ErrInvalidPhase
	}
	
	// Check if already finalized
	if auction.Phase == PhaseFinalized {
		return auction.Winner, auction.WinningBid, nil
	}
	
	// Find highest valid bid
	var winner std.Address
	var winningBid int64 = 0
	
	for _, commitment := range auction.Commitments {
		if commitment.Revealed && commitment.Amount > winningBid {
			winner = commitment.Bidder
			winningBid = commitment.Amount
		}
	}
	
	// Update auction status
	auction.Phase = PhaseFinalized
	auction.Winner = winner
	auction.WinningBid = winningBid
	auction.Finalized = time.Now()
	
	// Update auction in storage
	am.auctions.Set(name, auction)
	
	return winner, winningBid, nil
}

// GetAuctionStatus returns the current status of an auction
func (am *DefaultAuctionManager) GetAuctionStatus(name string) (AuctionStatus, error) {
	auctionVal, exists := am.auctions.Get(name)
	if !exists {
		return AuctionStatus{}, ErrAuctionNotFound
	}
	
	auction := auctionVal.(*Auction)
	
	// Count valid bids and find highest (if in reveal phase or later)
	bidCount := len(auction.Commitments)
	var highestBid int64 = 0
	
	if auction.Phase == PhaseReveal || auction.Phase == PhaseFinalized {
		for _, commitment := range auction.Commitments {
			if commitment.Revealed && commitment.Amount > highestBid {
				highestBid = commitment.Amount
			}
		}
	}
	
	return AuctionStatus{
		Name:           auction.Name,
		Phase:          auction.Phase,
		CommitDeadline: auction.CommitDeadline,
		RevealDeadline: auction.RevealDeadline,
		MinimumBid:     auction.MinimumBid,
		MinimumDeposit: auction.MinimumDeposit,
		BidCount:       bidCount,
		HighestBid:     highestBid,
		Winner:         auction.Winner,
	}, nil
}

// RefundLosers returns deposits to non-winning bidders
func (am *DefaultAuctionManager) RefundLosers(name string) error {
	auctionVal, exists := am.auctions.Get(name)
	if !exists {
		return ErrAuctionNotFound
	}
	
	auction := auctionVal.(*Auction)
	
	// Can only refund after auction is finalized
	if auction.Phase != PhaseFinalized {
		return ErrInvalidPhase
	}
	
	// Process refunds for all non-winning bidders
	for _, commitment := range auction.Commitments {
		// Skip the winner
		if commitment.Bidder == auction.Winner {
			continue
		}
		
		// Refund deposit (this would integrate with the banking system)
		depositKey := commitment.Bidder.String() + "_" + name
		am.deposits.Remove(depositKey)
		
		// In a real implementation, this would transfer the deposit back
		// For now, we just remove it from our tracking
	}
	
	return nil
}

// Helper methods

// computeBidHash computes the hash for a bid commitment
func (am *DefaultAuctionManager) computeBidHash(amount int64, salt string, bidder std.Address) string {
	// Create commitment string: amount + salt + bidder
	commitment := strconv.FormatInt(amount, 10) + salt + bidder.String()
	
	// Compute SHA256 hash
	hash := sha256.Sum256([]byte(commitment))
	
	// Return as hex string
	return hex.EncodeToString(hash[:])
}

// GetActiveAuctions returns all active auctions
func (am *DefaultAuctionManager) GetActiveAuctions() []string {
	var activeAuctions []string
	
	am.auctions.Iterate("", "", func(key string, value interface{}) bool {
		auction := value.(*Auction)
		if auction.Phase == PhaseCommit || auction.Phase == PhaseReveal {
			activeAuctions = append(activeAuctions, key)
		}
		return false // Continue iteration
	})
	
	return activeAuctions
}

// GetAuction returns the full auction details (for internal use)
func (am *DefaultAuctionManager) GetAuction(name string) (*Auction, bool) {
	auctionVal, exists := am.auctions.Get(name)
	if !exists {
		return nil, false
	}
	return auctionVal.(*Auction), true
}

// CancelAuction cancels an auction (admin function)
func (am *DefaultAuctionManager) CancelAuction(name string) error {
	auctionVal, exists := am.auctions.Get(name)
	if !exists {
		return ErrAuctionNotFound
	}
	
	auction := auctionVal.(*Auction)
	
	// Can only cancel if not finalized
	if auction.Phase == PhaseFinalized {
		return ErrInvalidPhase
	}
	
	// Mark as cancelled
	auction.Phase = PhaseCancelled
	am.auctions.Set(name, auction)
	
	// Refund all deposits
	for _, commitment := range auction.Commitments {
		depositKey := commitment.Bidder.String() + "_" + name
		am.deposits.Remove(depositKey)
	}
	
	return nil
}

// GetBidderCommitment returns a bidder's commitment for an auction
func (am *DefaultAuctionManager) GetBidderCommitment(name string, bidder std.Address) (*BidCommitment, bool) {
	auctionVal, exists := am.auctions.Get(name)
	if !exists {
		return nil, false
	}
	
	auction := auctionVal.(*Auction)
	
	for _, commitment := range auction.Commitments {
		if commitment.Bidder == bidder {
			return &commitment, true
		}
	}
	
	return nil, false
}

// ValidateBidHash validates that a hash is properly formatted
func ValidateBidHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	_, err := hex.DecodeString(hash)
	return err == nil
}

// ComputeBidHash is a utility function for clients to compute bid hashes
func ComputeBidHash(amount int64, salt string, bidder std.Address) string {
	commitment := strconv.FormatInt(amount, 10) + salt + bidder.String()
	hash := sha256.Sum256([]byte(commitment))
	return hex.EncodeToString(hash[:])
}
