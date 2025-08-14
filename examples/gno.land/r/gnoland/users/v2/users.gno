// Package users provides user registration with anti-squatting protection.
// This is an enhanced version of r/gnoland/users/v1 that integrates with
// the anti-squatting system for high-value names.
package users

import (
	"regexp"
	"std"
	"strconv"
	"time"
	
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/fifo"
	"gno.land/p/sys/antisquatting"
	susers "gno.land/r/sys/users"
)

const (
	reValidUsername = "^[a-z]{3}[_a-z0-9]{0,14}[0-9]{3}$"
)

var (
	// Basic registration price for non-high-value names
	registerPrice = int64(1_000_000) // 1 GNOT
	
	// Track latest registrations for rendering
	latestUsers = fifo.New(10)
	
	// Username validation
	reUsername = regexp.MustCompile(reValidUsername)
	
	// System state
	paused = false
)

// Register registers a new username for the caller.
// For high-value names, this will start an auction process.
// For regular names, this performs direct registration.
func Register(cur realm, username string) {
	if !std.PreviousRealm().IsUser() {
		panic(ErrNonUserCall)
	}
	
	if paused {
		panic(ErrPaused)
	}
	
	// Validate username format
	if matched := reUsername.MatchString(username); !matched {
		panic(ErrInvalidUsername)
	}
	
	registrant := std.PreviousRealm().Address()
	
	// Check if this is a high-value name requiring auction
	if susers.IsHighValueName(username) {
		// For high-value names, start auction process
		startNameAuction(username, registrant)
		return
	}
	
	// For regular names, perform direct registration
	performDirectRegistration(username, registrant)
}

// performDirectRegistration handles direct registration for non-high-value names
func performDirectRegistration(username string, registrant std.Address) {
	// Check payment for direct registration
	if std.OriginSend().AmountOf("ugnot") != registerPrice {
		panic(ErrInvalidPayment)
	}
	
	// Register through sys/users (this will use the enhanced registration)
	if err := susers.RegisterUserWithAntiSquatting(cross, username, registrant); err != nil {
		panic(err)
	}
	
	// Track for rendering
	latestUsers.Append(username)
	
	// Emit registration event
	std.Emit("Registration", 
		"address", registrant.String(), 
		"name", username,
		"type", "direct",
	)
}

// startNameAuction initiates the auction process for high-value names
func startNameAuction(username string, requester std.Address) {
	// Get registration info to show auction details
	info := susers.GetNameRegistrationInfo(username)
	
	if !info.Available {
		panic(ErrNameNotAvailable)
	}
	
	// Start the auction if it doesn't exist
	if info.AuctionStatus == nil {
		if err := susers.StartNameAuction(cross, username); err != nil {
			panic(err)
		}
		
		// Get updated info
		info = susers.GetNameRegistrationInfo(username)
	}
	
	// Emit auction started event
	std.Emit("AuctionStarted",
		"name", username,
		"requester", requester.String(),
		"minimum_bid", strconv.FormatInt(info.MinimumBid, 10),
		"minimum_deposit", strconv.FormatInt(info.MinimumDeposit, 10),
		"commit_deadline", info.CommitDuration.String(),
		"reveal_deadline", info.RevealDuration.String(),
	)
	
	// Panic with informative message about auction process
	panic(ufmt.Sprintf(
		"Name '%s' requires auction. Minimum bid: %d ugnot, Minimum deposit: %d ugnot. "+
			"Use SubmitBid() to participate in the auction.",
		username, info.MinimumBid, info.MinimumDeposit,
	))
}

// SubmitBid allows users to submit a bid commitment for a name auction
func SubmitBid(cur realm, username string, bidHash string) {
	if !std.PreviousRealm().IsUser() {
		panic(ErrNonUserCall)
	}
	
	if paused {
		panic(ErrPaused)
	}
	
	// Validate bid hash format
	if !antisquatting.ValidateBidHash(bidHash) {
		panic(ErrInvalidBidHash)
	}
	
	// Submit bid through sys/users
	if err := susers.SubmitBid(cross, username, bidHash); err != nil {
		panic(err)
	}
	
	bidder := std.PreviousRealm().Address()
	deposit := std.OriginSend().AmountOf("ugnot")
	
	// Emit bid submitted event
	std.Emit("BidSubmitted",
		"name", username,
		"bidder", bidder.String(),
		"deposit", strconv.FormatInt(deposit, 10),
	)
}

// RevealBid allows users to reveal their bid amount and salt
func RevealBid(cur realm, username string, amount int64, salt string) {
	if !std.PreviousRealm().IsUser() {
		panic(ErrNonUserCall)
	}
	
	if paused {
		panic(ErrPaused)
	}
	
	// Reveal bid through sys/users
	if err := susers.RevealBid(cross, username, amount, salt); err != nil {
		panic(err)
	}
	
	bidder := std.PreviousRealm().Address()
	
	// Emit bid revealed event
	std.Emit("BidRevealed",
		"name", username,
		"bidder", bidder.String(),
		"amount", strconv.FormatInt(amount, 10),
	)
}

// FinalizeAuction completes an auction and registers the name to the winner
func FinalizeAuction(cur realm, username string) {
	if paused {
		panic(ErrPaused)
	}
	
	// Finalize auction through sys/users
	if err := susers.FinalizeAuction(cross, username); err != nil {
		panic(err)
	}
	
	// Get auction status to find winner
	status, err := susers.GetAuctionStatus(username)
	if err != nil {
		panic(err)
	}
	
	// Track for rendering if there was a winner
	if status.Winner != "" {
		latestUsers.Append(username)
	}
	
	// Emit finalization event
	std.Emit("AuctionFinalized",
		"name", username,
		"winner", status.Winner.String(),
		"winning_bid", strconv.FormatInt(status.HighestBid, 10),
	)
}

// GetAuctionStatus returns the current status of a name auction
func GetAuctionStatus(username string) antisquatting.AuctionStatus {
	status, err := susers.GetAuctionStatus(username)
	if err != nil {
		panic(err)
	}
	return status
}

// CreateDispute files a dispute for a registered name
func CreateDispute(cur realm, username string, reason string) string {
	if !std.PreviousRealm().IsUser() {
		panic(ErrNonUserCall)
	}
	
	if paused {
		panic(ErrPaused)
	}
	
	// Create dispute through sys/users
	disputeID, err := susers.CreateDispute(cross, username, reason)
	if err != nil {
		panic(err)
	}
	
	disputer := std.PreviousRealm().Address()
	disputeFee := std.OriginSend().AmountOf("ugnot")
	
	// Emit dispute created event
	std.Emit("DisputeCreated",
		"dispute_id", disputeID,
		"name", username,
		"disputer", disputer.String(),
		"fee", strconv.FormatInt(disputeFee, 10),
	)
	
	return disputeID
}

// GetDisputeStatus returns the status of a dispute
func GetDisputeStatus(disputeID string) antisquatting.DisputeStatus {
	status, err := susers.GetDisputeStatus(disputeID)
	if err != nil {
		panic(err)
	}
	return status
}

// Utility functions

// IsHighValueName checks if a name requires auction-based registration
func IsHighValueName(username string) bool {
	return susers.IsHighValueName(username)
}

// GetNameRegistrationInfo returns information about registering a name
func GetNameRegistrationInfo(username string) antisquatting.NameRegistrationInfo {
	return susers.GetNameRegistrationInfo(username)
}

// ComputeBidHash helps users compute the hash for their bid commitment
func ComputeBidHash(amount int64, salt string, bidder std.Address) string {
	return antisquatting.ComputeBidHash(amount, salt, bidder)
}

// Administrative functions

// Pause pauses new registrations (admin only)
func Pause() {
	assertIsAdmin()
	paused = true
	std.Emit("SystemPaused")
}

// Unpause resumes registrations (admin only)
func Unpause() {
	assertIsAdmin()
	paused = false
	std.Emit("SystemUnpaused")
}

// SetRegisterPrice sets the price for direct registration (admin only)
func SetRegisterPrice(newPrice int64) {
	assertIsAdmin()
	registerPrice = newPrice
	std.Emit("PriceUpdated", "new_price", strconv.FormatInt(newPrice, 10))
}

// ProcessExpiredAuctions processes any auctions that have expired (admin only)
func ProcessExpiredAuctions() {
	assertIsAdmin()
	if err := susers.ProcessExpiredAuctions(); err != nil {
		panic(err)
	}
	std.Emit("ExpiredAuctionsProcessed")
}

// Helper functions

func assertIsAdmin() {
	// In a real implementation, this would check against a list of admin addresses
	// For now, we'll use a simple check
	caller := std.PreviousRealm().Address()
	if caller == "" {
		panic("admin access required")
	}
}

// GetSystemStatus returns the current system status
func GetSystemStatus() map[string]interface{} {
	status := susers.GetAntiSquattingSystemStatus()
	
	return map[string]interface{}{
		"paused":            paused,
		"register_price":    registerPrice,
		"antisquatting_enabled": susers.IsAntiSquattingEnabled(),
		"active_auctions":   status.ActiveAuctions,
		"pending_disputes":  status.PendingDisputes,
		"total_registered":  status.TotalRegistered,
		"total_auctioned":   status.TotalAuctioned,
		"total_disputed":    status.TotalDisputed,
	}
}
