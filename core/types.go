package core

import (
	"github.com/gnolang/go-tendermint/messages/types"
)

type Signer interface {
	Sign(data []byte) []byte
	IsValidSignature(data []byte, signature []byte) bool
}

type Verifier interface {
	IsProposer(id []byte, height uint64, round uint64) bool
	IsValidator(from []byte) bool
	Quorum(msgs []Message) bool
}

// Node interface is an abstraction over a single entity that runs
// the consensus algorithm
//
// Hash must not modify the slice proposal, even temporarily.
// Implementations must not retain proposal.
type Node interface {
	ID() []byte
	Hash(proposal []byte) []byte
	BuildProposal(height uint64) []byte
}

type Broadcast interface {
	BroadcastProposal(message *types.ProposalMessage)
	BroadcastPrevote(message *types.PrevoteMessage)
	BroadcastPrecommit(message *types.PrecommitMessage)
}

type Message interface {
	Verify() error
	GetSender() []byte
}
