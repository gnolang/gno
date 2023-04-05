package privval

import (
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// SignerMessage is sent between Signer Clients and Servers.
type SignerMessage interface{}

// TODO: Add ChainIDRequest

// PubKeyRequest requests the consensus public key from the remote signer.
type PubKeyRequest struct{}

// PubKeyResponse is a response message containing the public key.
type PubKeyResponse struct {
	PubKey crypto.PubKey
	Error  *RemoteSignerError
}

// SignVoteRequest is a request to sign a vote
type SignVoteRequest struct {
	Vote *types.Vote
}

// SignedVoteResponse is a response containing a signed vote or an error
type SignedVoteResponse struct {
	Vote  *types.Vote
	Error *RemoteSignerError
}

// SignProposalRequest is a request to sign a proposal
type SignProposalRequest struct {
	Proposal *types.Proposal
}

// SignedProposalResponse is response containing a signed proposal or an error
type SignedProposalResponse struct {
	Proposal *types.Proposal
	Error    *RemoteSignerError
}

// PingRequest is a request to confirm that the connection is alive.
type PingRequest struct{}

// PingResponse is a response to confirm that the connection is alive.
type PingResponse struct{}
