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
}

type Node interface {
	ID() []byte

	Hash(proposal []byte) []byte

	BuildProposal(height uint64) []byte
}

type Broadcast interface {
	Broadcast(message *types.Message)
}
