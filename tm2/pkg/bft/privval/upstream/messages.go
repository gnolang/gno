package upstream

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// Privval socket-protocol messages, byte-compatible with upstream Tendermint
// v0.34's privval/types.proto. These are the messages exchanged between a
// validator and an external KMS (tmkms, Horcrux, etc.) over a SecretConnection.
//
// Each message is encoded as a oneof Message wrapper at the wire level; the
// Message struct below dispatches on the Sum field analogous to upstream's
// gogoproto-generated oneof.

// Message is the privval-protocol envelope. Exactly one of the Sum fields
// is set per wire message. Field numbers match upstream's
//
//	message Message {
//	  oneof sum {
//	    PubKeyRequest          pub_key_request           = 1;
//	    PubKeyResponse         pub_key_response          = 2;
//	    SignVoteRequest        sign_vote_request         = 3;
//	    SignedVoteResponse     signed_vote_response      = 4;
//	    SignProposalRequest    sign_proposal_request     = 5;
//	    SignedProposalResponse signed_proposal_response  = 6;
//	    PingRequest            ping_request              = 7;
//	    PingResponse           ping_response             = 8;
//	  }
//	}
type Message struct {
	PubKeyRequest          *PubKeyRequest
	PubKeyResponse         *PubKeyResponse
	SignVoteRequest        *SignVoteRequest
	SignedVoteResponse     *SignedVoteResponse
	SignProposalRequest    *SignProposalRequest
	SignedProposalResponse *SignedProposalResponse
	PingRequest            *PingRequest
	PingResponse           *PingResponse
}

// PubKeyRequest asks the signer for the validator's consensus pubkey.
// chain_id is included so the signer can verify it matches its
// configured chain (defense-in-depth against being fed signed messages
// from a different network).
//
//	message PubKeyRequest {
//	  string chain_id = 1;
//	}
type PubKeyRequest struct {
	ChainID string
}

// PubKeyResponse returns the consensus pubkey or an error.
//
//	message PubKeyResponse {
//	  PublicKey          pub_key = 1;
//	  RemoteSignerError  error   = 2;
//	}
//
// PubKey here uses tm2's crypto.PubKey interface; on the wire it carries
// upstream's Tendermint PublicKey oneof (sum of ed25519/secp256k1).
// Callers serializing to upstream-compat bytes must select the matching
// oneof variant.
type PubKeyResponse struct {
	PubKey crypto.PubKey
	Error  *RemoteSignerError
}

// SignVoteRequest carries a Vote (operationally-shaped, validator already
// filled all fields except Signature) and the chain_id. The signer canonicalizes
// using its CONFIGURED chain_id (which it MAY check against this field,
// or refuse if mismatched — tmkms takes the latter approach).
//
//	message SignVoteRequest {
//	  Vote   vote     = 1;
//	  string chain_id = 2;
//	}
type SignVoteRequest struct {
	Vote    *Vote
	ChainID string
}

// SignedVoteResponse echoes the Vote with .Signature populated, or an error.
//
//	message SignedVoteResponse {
//	  Vote              vote  = 1;
//	  RemoteSignerError error = 2;
//	}
type SignedVoteResponse struct {
	Vote  *Vote
	Error *RemoteSignerError
}

// SignProposalRequest mirrors SignVoteRequest for Proposal.
//
//	message SignProposalRequest {
//	  Proposal proposal = 1;
//	  string   chain_id = 2;
//	}
type SignProposalRequest struct {
	Proposal *Proposal
	ChainID  string
}

// SignedProposalResponse echoes the Proposal with .Signature populated.
//
//	message SignedProposalResponse {
//	  Proposal           proposal = 1;
//	  RemoteSignerError  error    = 2;
//	}
type SignedProposalResponse struct {
	Proposal *Proposal
	Error    *RemoteSignerError
}

// PingRequest/PingResponse are the application-level keepalive. Empty
// messages on the wire (just a tag).
//
//	message PingRequest  {}
//	message PingResponse {}
type (
	PingRequest  struct{}
	PingResponse struct{}
)

// RemoteSignerError is the structured error returned when the KMS refuses
// to sign or fails internally.
//
//	message RemoteSignerError {
//	  int32  code        = 1;
//	  string description = 2;
//	}
type RemoteSignerError struct {
	Code        int32 `binary:"varint"`
	Description string
}

// Error implements the error interface, allowing RemoteSignerError values
// to be returned through Go's error machinery while preserving the
// upstream-protocol structure on the wire.
func (e *RemoteSignerError) Error() string {
	if e == nil {
		return ""
	}
	return e.Description
}
