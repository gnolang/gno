package remote

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// MaxMessageSize is the maximum size of a message that can be sent or received.
const MaxMessageSize = 1024

// RemoteSignerMessage is sent between Remote Signer clients and servers.
type RemoteSignerMessage interface{}

// PubKeyRequest requests the signing public key from the remote signer.
type PubKeyRequest struct{}

// PubKeyResponse is a response containing the public key or an error.
type PubKeyResponse struct {
	PubKey crypto.PubKey
	Error  error
}

// SignRequest is a request to sign arbitrary bytes.
type SignRequest struct {
	SignBytes []byte
}

// SignResponse is a response containing the signature or an error.
type SignResponse struct {
	Signature []byte
	Error     error
}

// PingRequest is a request to confirm that the connection is alive.
type PingRequest struct{}

// PingResponse is a response to confirm that the connection is alive.
type PingResponse struct{}
