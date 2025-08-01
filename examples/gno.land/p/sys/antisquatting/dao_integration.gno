package antisquatting

import (
	"std"
	"strconv"
	"time"
	
	"gno.land/p/demo/ufmt"
)

// DAODisputeResolver integrates with the Gno DAO system for dispute resolution
type DAODisputeResolver struct {
	*DefaultDisputeResolver
	
	// DAO integration settings
	daoRealm        string // Path to DAO realm (e.g., "gno.land/r/gov/dao")
	proposalPrefix  string // Prefix for dispute proposals
	autoExecute     bool   // Whether to auto-execute approved disputes
}

// NewDAODisputeResolver creates a dispute resolver with DAO integration
func NewDAODisputeResolver(daoRealm string) *DAODisputeResolver {
	return &DAODisputeResolver{
		DefaultDisputeResolver: NewDefaultDisputeResolver(),
		daoRealm:              daoRealm,
		proposalPrefix:        "ANTISQUATTING_DISPUTE",
		autoExecute:           true,
	}
}

// CreateDispute files a new dispute and creates a DAO proposal
func (ddr *DAODisputeResolver) CreateDispute(name string, owner std.Address, disputer std.Address, reason string) (string, error) {
	// Create the base dispute
	disputeID, err := ddr.DefaultDisputeResolver.CreateDispute(name, owner, disputer, reason)
	if err != nil {
		return "", err
	}
	
	// Create DAO proposal for the dispute
	if err := ddr.createDAOProposal(disputeID, name, owner, disputer, reason); err != nil {
		// If DAO proposal creation fails, we should still keep the dispute
		// but mark it with an error status
		return disputeID, err
	}
	
	return disputeID, nil
}

// createDAOProposal creates a DAO proposal for dispute resolution
func (ddr *DAODisputeResolver) createDAOProposal(disputeID, name string, owner, disputer std.Address, reason string) error {
	// In a real implementation, this would call the DAO realm to create a proposal
	// For now, we simulate the proposal creation
	
	title := ufmt.Sprintf("%s: Dispute for name '%s'", ddr.proposalPrefix, name)
	description := ufmt.Sprintf(
		"Dispute Resolution Request\n\n"+
			"Name: %s\n"+
			"Current Owner: %s\n"+
			"Disputer: %s\n"+
			"Dispute ID: %s\n\n"+
			"Reason:\n%s\n\n"+
			"Please vote YES to approve the dispute (transfer name to disputer) or NO to reject it.",
		name, owner.String(), disputer.String(), disputeID, reason,
	)
	
	// Create the proposal executor that will handle the dispute resolution
	executor := &DisputeExecutor{
		disputeID: disputeID,
		resolver:  ddr,
	}
	
	// In a real implementation, this would call:
	// dao.CreateProposal(title, description, executor)
	
	// For now, we just log the proposal creation
	// The actual DAO integration would happen here
	
	return nil
}

// DisputeExecutor implements the DAO proposal executor interface
type DisputeExecutor struct {
	disputeID string
	resolver  *DAODisputeResolver
}

// Execute is called by the DAO when a proposal is approved
func (de *DisputeExecutor) Execute(cur realm) error {
	// This would be called by the DAO system when the proposal passes
	return de.resolver.ResolveDispute(de.disputeID, true)
}

// GetTitle returns the proposal title
func (de *DisputeExecutor) GetTitle() string {
	dispute, exists := de.resolver.GetDispute(de.disputeID)
	if !exists {
		return "Unknown Dispute"
	}
	return ufmt.Sprintf("Dispute Resolution: %s", dispute.Name)
}

// GetDescription returns the proposal description
func (de *DisputeExecutor) GetDescription() string {
	dispute, exists := de.resolver.GetDispute(de.disputeID)
	if !exists {
		return "Dispute not found"
	}
	
	return ufmt.Sprintf(
		"Resolve dispute for name '%s'\n"+
			"Disputer: %s\n"+
			"Reason: %s\n"+
			"Filed: %s",
		dispute.Name,
		dispute.Disputer.String(),
		dispute.Reason,
		dispute.Created.Format("2006-01-02 15:04:05"),
	)
}

// Enhanced dispute resolution with DAO callbacks

// OnProposalCreated is called when a DAO proposal is successfully created
func (ddr *DAODisputeResolver) OnProposalCreated(disputeID string, proposalID int64) error {
	dispute, exists := ddr.GetDispute(disputeID)
	if !exists {
		return ErrDisputeNotFound
	}
	
	// Store the proposal ID for tracking
	// In a real implementation, this would be stored in the dispute record
	
	return nil
}

// OnProposalVoted is called when someone votes on a dispute proposal
func (ddr *DAODisputeResolver) OnProposalVoted(disputeID string, voter std.Address, vote string) error {
	// This could be used for notifications or tracking
	return nil
}

// OnProposalExecuted is called when a dispute proposal is executed
func (ddr *DAODisputeResolver) OnProposalExecuted(disputeID string, approved bool) error {
	return ddr.ResolveDispute(disputeID, approved)
}

// Configuration and utility methods

func (ddr *DAODisputeResolver) SetDAORealm(realm string) {
	ddr.daoRealm = realm
}

func (ddr *DAODisputeResolver) SetProposalPrefix(prefix string) {
	ddr.proposalPrefix = prefix
}

func (ddr *DAODisputeResolver) SetAutoExecute(autoExecute bool) {
	ddr.autoExecute = autoExecute
}

func (ddr *DAODisputeResolver) GetDAORealm() string {
	return ddr.daoRealm
}

// DisputeProposalRequest represents a request to create a DAO proposal for a dispute
type DisputeProposalRequest struct {
	DisputeID   string
	Name        string
	Owner       std.Address
	Disputer    std.Address
	Reason      string
	Title       string
	Description string
	Executor    *DisputeExecutor
}

// CreateDisputeProposalRequest creates a standardized proposal request
func (ddr *DAODisputeResolver) CreateDisputeProposalRequest(disputeID string) (*DisputeProposalRequest, error) {
	dispute, exists := ddr.GetDispute(disputeID)
	if !exists {
		return nil, ErrDisputeNotFound
	}
	
	title := ufmt.Sprintf("%s: %s", ddr.proposalPrefix, dispute.Name)
	description := ufmt.Sprintf(
		"Name Dispute Resolution\n\n"+
			"Disputed Name: %s\n"+
			"Current Owner: %s\n"+
			"Disputer: %s\n"+
			"Filed: %s\n\n"+
			"Dispute Reason:\n%s\n\n"+
			"Voting Instructions:\n"+
			"- Vote YES to approve the dispute and transfer the name\n"+
			"- Vote NO to reject the dispute and maintain current ownership\n\n"+
			"This proposal will automatically execute if approved.",
		dispute.Name,
		dispute.Owner.String(),
		dispute.Disputer.String(),
		dispute.Created.Format("2006-01-02 15:04:05"),
		dispute.Reason,
	)
	
	executor := &DisputeExecutor{
		disputeID: disputeID,
		resolver:  ddr,
	}
	
	return &DisputeProposalRequest{
		DisputeID:   disputeID,
		Name:        dispute.Name,
		Owner:       dispute.Owner,
		Disputer:    dispute.Disputer,
		Reason:      dispute.Reason,
		Title:       title,
		Description: description,
		Executor:    executor,
	}, nil
}

// GetDisputeProposalStatus returns the status of a dispute's DAO proposal
func (ddr *DAODisputeResolver) GetDisputeProposalStatus(disputeID string) (string, error) {
	// In a real implementation, this would query the DAO system
	// to get the current status of the proposal
	
	dispute, exists := ddr.GetDispute(disputeID)
	if !exists {
		return "", ErrDisputeNotFound
	}
	
	switch dispute.Status {
	case DisputePending:
		return "voting", nil
	case DisputeApproved:
		return "approved", nil
	case DisputeRejected:
		return "rejected", nil
	case DisputeExpired:
		return "expired", nil
	default:
		return "unknown", nil
	}
}

// Helper functions for DAO integration

// IsDAOConfigured checks if the DAO integration is properly configured
func (ddr *DAODisputeResolver) IsDAOConfigured() bool {
	return ddr.daoRealm != ""
}

// GetRequiredDAOQuorum returns the minimum number of votes needed
func (ddr *DAODisputeResolver) GetRequiredDAOQuorum() int64 {
	return ddr.GetRequiredQuorum()
}

// GetDAOVotingPeriod returns how long DAO members have to vote
func (ddr *DAODisputeResolver) GetDAOVotingPeriod() time.Duration {
	return ddr.GetVotingPeriod()
}

// ValidateDAOIntegration checks if the DAO integration is working
func (ddr *DAODisputeResolver) ValidateDAOIntegration() error {
	if !ddr.IsDAOConfigured() {
		return ErrSystemNotConfigured
	}
	
	// In a real implementation, this would test the connection to the DAO realm
	// and verify that proposals can be created
	
	return nil
}

// GetDAOIntegrationInfo returns information about the DAO integration
func (ddr *DAODisputeResolver) GetDAOIntegrationInfo() map[string]interface{} {
	return map[string]interface{}{
		"dao_realm":        ddr.daoRealm,
		"proposal_prefix":  ddr.proposalPrefix,
		"auto_execute":     ddr.autoExecute,
		"voting_period":    ddr.GetVotingPeriod().String(),
		"required_quorum":  ddr.GetRequiredQuorum(),
		"dispute_fee":      ddr.GetDisputeFee(),
		"configured":       ddr.IsDAOConfigured(),
	}
}
