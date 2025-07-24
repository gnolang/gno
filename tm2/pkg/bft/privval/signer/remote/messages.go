package remote

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// MaxMessageSize is the maximum size in bytes of a message that can be sent or received.
const MaxMessageSize = 10 * 1024

// RemoteSignerError is an error returned by the remote signer.
// Necessary because golang errors are not serializable (private fields).
type RemoteSignerError struct {
	Err string
}

// RemoteSignerError type implements error.
var _ error = (*RemoteSignerError)(nil)

// Error implements error.
func (rse *RemoteSignerError) Error() string {
	return rse.Err
}

// RemoteSignerMessage is sent between Remote Signer clients and servers.
type RemoteSignerMessage interface{}

// PubKeyRequest requests the signing public key from the remote signer.
type PubKeyRequest struct{}

// PubKeyResponse is a response containing the public key.
type PubKeyResponse struct {
	PubKey crypto.PubKey
}

// SignRequest is a request to sign arbitrary bytes.
type SignRequest struct {
	SignBytes []byte
}

// SignResponse is a response containing the signature or an error.
type SignResponse struct {
	Signature []byte
	Error     *RemoteSignerError
}

// PingRequest is a request to confirm that the connection is alive.
type PingRequest struct{}

// PingResponse is a response to confirm that the connection is alive.
type PingResponse struct{}
