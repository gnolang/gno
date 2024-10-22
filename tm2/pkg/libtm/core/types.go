package core

import (
	"github.com/gnolang/libtm/messages/types"
)

// Signer is an abstraction over the signature manipulation process
type Signer interface {
	// Sign generates a signature for the given raw data
	Sign(data []byte) []byte

	// IsValidSignature verifies whether the signature matches the raw data
	IsValidSignature(data []byte, signature []byte) bool
}

// Verifier is an abstraction over the outer consensus calling context
// that has access to validator set information
type Verifier interface {
	// IsProposer checks if the given ID matches the proposer for the given height
	IsProposer(id []byte, height uint64, round uint64) bool

	// IsValidator checks if the given message sender ID belongs to a validator
	IsValidator(id []byte) bool

	// IsValidProposal checks if the given proposal is valid, for the given height
	IsValidProposal(proposal []byte, height uint64) bool

	// GetSumVotingPower returns the summed voting power from
	// the given unique message authors
	GetSumVotingPower(msgs []Message) uint64

	// GetTotalVotingPower returns the total voting power
	// of the entire validator set for the given height
	GetTotalVotingPower(height uint64) uint64
}

// Node interface is an abstraction over a single entity (current process) that runs
// the consensus algorithm
type Node interface {
	// ID returns the ID associated with the current process (validator)
	ID() []byte

	// Hash generates a hash of the given data.
	// It must not modify the slice proposal, even temporarily
	// and must not retain the data
	Hash(proposal []byte) []byte

	// BuildProposal generates a raw proposal for the given height
	BuildProposal(height uint64) []byte
}

// Broadcast is an abstraction over the networking / message sharing interface
// that enables message passing between validators
type Broadcast interface {
	// BroadcastPropose broadcasts a PROPOSAL message
	BroadcastPropose(message *types.ProposalMessage)

	// BroadcastPrevote broadcasts a PREVOTE message
	BroadcastPrevote(message *types.PrevoteMessage)

	// BroadcastPrecommit broadcasts a PRECOMMIT message
	BroadcastPrecommit(message *types.PrecommitMessage)
}

// Message is the content being passed around
// between consensus validators.
// Message types: PROPOSAL, PREVOTE, PRECOMMIT
type Message interface {
	// GetView fetches the message view
	GetView() *types.View

	// GetSender fetches the message sender
	GetSender() []byte

	// GetSignature fetches the message signature
	GetSignature() []byte

	// GetSignaturePayload fetches the signature payload (sign data)
	GetSignaturePayload() []byte

	// Verify verifies the message content is valid (base verification)
	Verify() error
}
