package antisquatting

import (
	"std"
	"strconv"
	"strings"
	"time"
	
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

// DefaultDisputeResolver implements DisputeResolver with DAO integration
type DefaultDisputeResolver struct {
	disputes       *avl.Tree // disputeID -> *Dispute
	nameDisputes   *avl.Tree // name -> []disputeID
	disputeCounter seqid.ID
	
	// Configuration
	disputeFee     int64         // Cost to file a dispute
	votingPeriod   time.Duration // How long DAO has to vote
	requiredQuorum int64         // Minimum votes needed
	
	// DAO integration (would be injected in real implementation)
	daoProxyRealm string // Path to DAO proxy realm
}

// NewDefaultDisputeResolver creates a new dispute resolver
func NewDefaultDisputeResolver() *DefaultDisputeResolver {
	return &DefaultDisputeResolver{
		disputes:       avl.NewTree(),
		nameDisputes:   avl.NewTree(),
		disputeCounter: 1,
		
		// Default configuration
		disputeFee:     5_000_000, // 5 GNOT
		votingPeriod:   7 * 24 * time.Hour, // 1 week
		requiredQuorum: 3, // Minimum 3 votes
		
		// DAO integration
		daoProxyRealm: "gno.land/r/gov/dao",
	}
}

// CanDispute returns true if the disputer can file a dispute for this name
func (dr *DefaultDisputeResolver) CanDispute(name string, disputer std.Address) bool {
	// Basic validation
	if name == "" || disputer == "" {
		return false
	}
	
	// Check if there's already an active dispute for this name by this disputer
	if dr.hasActiveDispute(name, disputer) {
		return false
	}
	
	// Check dispute reasons - for now, allow anyone to dispute
	// In a real implementation, this might check:
	// - Trademark ownership
	// - Identity verification
	// - Reputation scores
	// - Time limits after registration
	
	return true
}

// CreateDispute files a new dispute with the given reason
func (dr *DefaultDisputeResolver) CreateDispute(name string, owner std.Address, disputer std.Address, reason string) (string, error) {
	// Validate inputs
	if name == "" {
		return "", ErrInvalidDisputer
	}
	if owner == "" || disputer == "" {
		return "", ErrInvalidDisputer
	}
	if reason == "" {
		return "", ErrInvalidDisputer
	}
	
	// Check if disputer can dispute this name
	if !dr.CanDispute(name, disputer) {
		return "", ErrCannotDispute
	}
	
	// Generate dispute ID
	disputeID := "dispute_" + strconv.FormatInt(int64(dr.disputeCounter), 10)
	dr.disputeCounter++
	
	// Create dispute
	dispute := &Dispute{
		ID:       disputeID,
		Name:     name,
		Owner:    owner,
		Disputer: disputer,
		Reason:   reason,
		Status:   DisputePending,
		Created:  time.Now(),
	}
	
	// Store dispute
	dr.disputes.Set(disputeID, dispute)
	
	// Track disputes by name
	dr.addNameDispute(name, disputeID)
	
	// In a real implementation, this would create a DAO proposal
	// For now, we just store the dispute
	
	return disputeID, nil
}

// GetDisputeStatus returns the current status of a dispute
func (dr *DefaultDisputeResolver) GetDisputeStatus(disputeID string) (DisputeStatus, error) {
	disputeVal, exists := dr.disputes.Get(disputeID)
	if !exists {
		return DisputePending, ErrDisputeNotFound
	}
	
	dispute := disputeVal.(*Dispute)
	
	// Check if dispute has expired
	if dispute.Status == DisputePending {
		if time.Now().After(dispute.Created.Add(dr.votingPeriod)) {
			dispute.Status = DisputeExpired
			dispute.Resolved = time.Now()
			dr.disputes.Set(disputeID, dispute)
		}
	}
	
	return dispute.Status, nil
}

// ResolveDispute executes the resolution of a dispute (called by DAO)
func (dr *DefaultDisputeResolver) ResolveDispute(disputeID string, approved bool) error {
	disputeVal, exists := dr.disputes.Get(disputeID)
	if !exists {
		return ErrDisputeNotFound
	}
	
	dispute := disputeVal.(*Dispute)
	
	// Can only resolve pending disputes
	if dispute.Status != DisputePending {
		return ErrInvalidPhase
	}
	
	// Update dispute status
	if approved {
		dispute.Status = DisputeApproved
	} else {
		dispute.Status = DisputeRejected
	}
	dispute.Resolved = time.Now()
	
	// Store updated dispute
	dr.disputes.Set(disputeID, dispute)
	
	return nil
}

// Helper methods

// hasActiveDispute checks if there's an active dispute for this name by this disputer
func (dr *DefaultDisputeResolver) hasActiveDispute(name string, disputer std.Address) bool {
	disputeIDsVal, exists := dr.nameDisputes.Get(name)
	if !exists {
		return false
	}
	
	disputeIDs := disputeIDsVal.([]string)
	
	for _, disputeID := range disputeIDs {
		disputeVal, exists := dr.disputes.Get(disputeID)
		if !exists {
			continue
		}
		
		dispute := disputeVal.(*Dispute)
		if dispute.Disputer == disputer && dispute.Status == DisputePending {
			return true
		}
	}
	
	return false
}

// addNameDispute adds a dispute ID to the list of disputes for a name
func (dr *DefaultDisputeResolver) addNameDispute(name string, disputeID string) {
	disputeIDsVal, exists := dr.nameDisputes.Get(name)
	var disputeIDs []string
	
	if exists {
		disputeIDs = disputeIDsVal.([]string)
	} else {
		disputeIDs = make([]string, 0)
	}
	
	disputeIDs = append(disputeIDs, disputeID)
	dr.nameDisputes.Set(name, disputeIDs)
}

// GetDispute returns the full dispute details
func (dr *DefaultDisputeResolver) GetDispute(disputeID string) (*Dispute, bool) {
	disputeVal, exists := dr.disputes.Get(disputeID)
	if !exists {
		return nil, false
	}
	return disputeVal.(*Dispute), true
}

// GetDisputesForName returns all disputes for a given name
func (dr *DefaultDisputeResolver) GetDisputesForName(name string) []*Dispute {
	disputeIDsVal, exists := dr.nameDisputes.Get(name)
	if !exists {
		return nil
	}
	
	disputeIDs := disputeIDsVal.([]string)
	disputes := make([]*Dispute, 0, len(disputeIDs))
	
	for _, disputeID := range disputeIDs {
		if dispute, exists := dr.GetDispute(disputeID); exists {
			disputes = append(disputes, dispute)
		}
	}
	
	return disputes
}

// GetActiveDisputes returns all pending disputes
func (dr *DefaultDisputeResolver) GetActiveDisputes() []*Dispute {
	var activeDisputes []*Dispute
	
	dr.disputes.Iterate("", "", func(key string, value interface{}) bool {
		dispute := value.(*Dispute)
		if dispute.Status == DisputePending {
			// Check if expired
			if time.Now().After(dispute.Created.Add(dr.votingPeriod)) {
				dispute.Status = DisputeExpired
				dispute.Resolved = time.Now()
				dr.disputes.Set(key, dispute)
			} else {
				activeDisputes = append(activeDisputes, dispute)
			}
		}
		return false // Continue iteration
	})
	
	return activeDisputes
}

// GetDisputesByDisputer returns all disputes filed by a specific disputer
func (dr *DefaultDisputeResolver) GetDisputesByDisputer(disputer std.Address) []*Dispute {
	var disputes []*Dispute
	
	dr.disputes.Iterate("", "", func(key string, value interface{}) bool {
		dispute := value.(*Dispute)
		if dispute.Disputer == disputer {
			disputes = append(disputes, dispute)
		}
		return false // Continue iteration
	})
	
	return disputes
}

// Configuration methods

func (dr *DefaultDisputeResolver) SetDisputeFee(fee int64) {
	dr.disputeFee = fee
}

func (dr *DefaultDisputeResolver) SetVotingPeriod(period time.Duration) {
	dr.votingPeriod = period
}

func (dr *DefaultDisputeResolver) SetRequiredQuorum(quorum int64) {
	dr.requiredQuorum = quorum
}

func (dr *DefaultDisputeResolver) SetDAOProxyRealm(realm string) {
	dr.daoProxyRealm = realm
}

// Getters for configuration

func (dr *DefaultDisputeResolver) GetDisputeFee() int64 {
	return dr.disputeFee
}

func (dr *DefaultDisputeResolver) GetVotingPeriod() time.Duration {
	return dr.votingPeriod
}

func (dr *DefaultDisputeResolver) GetRequiredQuorum() int64 {
	return dr.requiredQuorum
}

// Dispute validation helpers

// IsValidDisputeReason checks if a dispute reason is valid
func IsValidDisputeReason(reason string) bool {
	reason = strings.TrimSpace(reason)
	
	// Must have minimum length
	if len(reason) < 10 {
		return false
	}
	
	// Must not be too long
	if len(reason) > 1000 {
		return false
	}
	
	// Check for common dispute categories
	validCategories := []string{
		"trademark", "impersonation", "system", "abuse",
		"copyright", "brand", "identity", "fraud",
	}
	
	reasonLower := strings.ToLower(reason)
	for _, category := range validCategories {
		if strings.Contains(reasonLower, category) {
			return true
		}
	}
	
	// Allow generic disputes with sufficient detail
	return len(reason) >= 50
}

// GetDisputeStatistics returns statistics about disputes
func (dr *DefaultDisputeResolver) GetDisputeStatistics() map[string]int {
	stats := map[string]int{
		"total":    0,
		"pending":  0,
		"approved": 0,
		"rejected": 0,
		"expired":  0,
	}
	
	dr.disputes.Iterate("", "", func(key string, value interface{}) bool {
		dispute := value.(*Dispute)
		stats["total"]++
		
		switch dispute.Status {
		case DisputePending:
			stats["pending"]++
		case DisputeApproved:
			stats["approved"]++
		case DisputeRejected:
			stats["rejected"]++
		case DisputeExpired:
			stats["expired"]++
		}
		
		return false // Continue iteration
	})
	
	return stats
}
