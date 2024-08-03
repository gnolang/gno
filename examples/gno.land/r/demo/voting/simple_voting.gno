package voting

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/bitmap"
	"gno.land/p/demo/nonce"
	"gno.land/p/demo/seqid"
)

// Proposal represents a voting proposal with associated metadata and voting state
type Proposal struct {
	owner           std.Address    // Owner is the address of the user who created the proposal
	title           string         // Title is the name or description of the proposal
	votes           uint64         // Votes is the total number of votes the proposal has received
	voters          *bitmap.Bitmap // Voters is a bitmap tracking whether users have voted
	userVoteCount   *nonce.Manager // UserVoteCount maps user addresses to the number of votes they have cast
	maxVotesPerUser uint64         // MaxVotesPerUser is the maximum number of votes a single user can cast
	isClosed        bool           // IsClosed indicates whether the proposal is closed for voting
}

var (
	voterIndex    seqid.ID // Sequential ID generator for voters
	proposalIndex seqid.ID // Sequential ID generator for proposals

	proposals        *avl.Tree // AVL tree to store proposals
	registeredVoters *avl.Tree // AVL tree to store registered voters
)

// init function initializes the voter and proposal indices and AVL trees
func Init() {
	voterIndex = 0
	proposalIndex = 0

	proposals = avl.NewTree()
	registeredVoters = avl.NewTree()
}

// CreateProposal creates a new proposal with the given title, number of participants, and max votes per user
func CreateProposal(title string, numParticipants uint64, maxVotesPerUser uint64) string {
	proposal := &Proposal{
		owner:           std.PrevRealm().Addr(),
		title:           title,
		votes:           0,
		voters:          bitmap.New(numParticipants),
		userVoteCount:   nonce.New(),
		maxVotesPerUser: maxVotesPerUser,
		isClosed:        false,
	}

	// Get the next proposal ID
	proposalID := proposalIndex.Next()

	// Store the proposal in the AVL tree
	proposals.Set(proposalID.String(), proposal)

	// Return the proposal ID as a string
	return proposalID.String()
}

// RegisterVoter registers a new voter and returns a voter ID
func RegisterVoter() string {
	// Get the caller's address
	caller := std.PrevRealm().Addr()

	// Check if the address is already registered
	if voter, exists := registeredVoters.Get(caller.String()); exists {
		voterId := voter.(string)
		// Return the existing voter ID
		return voterId
	}

	// Get the next voter ID
	voterID := voterIndex.Next()

	// Store the voter in the AVL tree
	registeredVoters.Set(caller.String(), voterID.String())

	// Return the new voter ID
	return voterID.String()
}

// Vote casts a vote for a proposal identified by proposalID and voterID
func Vote(proposalID, voterID string) error {
	// Get the caller's address
	caller := std.PrevRealm().Addr()

	var voter seqid.ID
	var exists bool

	// Check if the voter is registered
	if voter, exists := registeredVoters.Get(caller.String()); exists {
		if voter.(string) != voterID {
			// Voter ID does not match
			return errors.New("invalid voterID")
		}
	} else {
		// Voter is not registered
		return errors.New("must register as voter first")
	}

	// Get the proposal
	proposal, exists := proposals.Get(proposalID)
	if !exists {
		// Proposal does not exist
		return errors.New("proposal not found")
	}

	prop := proposal.(*Proposal)

	if prop.isClosed {
		// Proposal is closed for voting
		return errors.New("proposal is closed")
	}

	// Get the current vote count for the caller
	currentNonce := prop.userVoteCount.GetNonce(caller)
	// Validate nonce
	if currentNonce >= prop.maxVotesPerUser {
		// User has reached the vote limit
		return errors.New("reach limit vote per user")
	}

	// Mark the voter as having voted and increment vote count
	prop.voters.MustSet(uint64(voter))

	prop.votes++

	// Increment the user's vote count
	prop.userVoteCount.UseNonce(caller)

	return nil
}

func CloseProposal(proposalID string) error {
	// Get the caller's address
	caller := std.PrevRealm().Addr()

	// Get the proposal
	proposal, exists := proposals.Get(proposalID)
	if !exists {
		// Proposal does not exist
		return errors.New("proposal not found")
	}

	prop := proposal.(*Proposal)

	if caller != prop.owner {
		return errors.New("unauthorized")
	}

	prop.isClosed = true

	return nil
}

func GetOwner(proposalID string) (string, error) {
	// Get the proposal
	proposal, exists := proposals.Get(proposalID)
	if !exists {
		// Proposal does not exist
		return "", errors.New("proposal not found")
	}

	prop := proposal.(*Proposal)

	// Return the vote count
	return prop.owner.String(), nil
}

// GetVotes returns the current vote count for a proposal identified by proposalID
func GetVotes(proposalID string) (uint64, error) {
	// Get the proposal
	proposal, exists := proposals.Get(proposalID)
	if !exists {
		// Proposal does not exist
		return 0, errors.New("proposal not found")
	}

	prop := proposal.(*Proposal)

	// Return the vote count
	return prop.votes, nil
}

func Render(path string) string {
	// do something here
	return ""
}
