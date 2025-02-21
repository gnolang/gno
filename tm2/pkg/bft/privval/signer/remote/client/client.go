package client

import (
	"fmt"
	"io"

	r "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// RemoteSignerClient type implements types.Signer.
var _ types.Signer = (*RemoteSignerClient)(nil)

// PubKey implements types.Signer.
func (rsc *RemoteSignerClient) PubKey() (crypto.PubKey, error) {
	response, err := rsc.send(&r.PubKeyRequest{})
	if err != nil {
		err := fmt.Errorf("%w: %w", ErrSendingRequestFailed, err)
		rsc.logger.Error("PubKey request failed", "error", err)
		return nil, err
	}

	pubKeyResponse, ok := response.(*r.PubKeyResponse)
	if !ok {
		err := fmt.Errorf("%w: %T", ErrInvalidResponseType, response)
		rsc.logger.Error("PubKey request failed", "error", err)
		return nil, err
	}

	if pubKeyResponse.Error != nil {
		err := fmt.Errorf("%w: %w", ErrResponseContainsError, pubKeyResponse.Error)
		rsc.logger.Error("PubKey request failed", "error", err)
		return nil, err
	}

	rsc.logger.Debug("PubKey request succeeded")

	return pubKeyResponse.PubKey, nil
}

// Sign implements types.Signer.
func (rsc *RemoteSignerClient) Sign(signBytes []byte) ([]byte, error) {
	response, err := rsc.send(&r.SignRequest{SignBytes: signBytes})
	if err != nil {
		err := fmt.Errorf("%w: %w", ErrSendingRequestFailed, err)
		rsc.logger.Error("Sign request failed", "error", err)
		return nil, err
	}

	signResponse, ok := response.(*r.SignResponse)
	if !ok {
		err := fmt.Errorf("%w: %T", ErrInvalidResponseType, response)
		rsc.logger.Error("Sign request failed", "error", err)
		return nil, err
	}

	if signResponse.Error != nil {
		err := fmt.Errorf("%w: %w", ErrResponseContainsError, signResponse.Error)
		rsc.logger.Error("Sign request failed", "error", err)
		return nil, err
	}

	rsc.logger.Debug("Sign request succeeded")

	return signResponse.Signature, nil
}

// Ping sends a ping request to the server.
func (rsc *RemoteSignerClient) Ping() error {
	response, err := rsc.send(&r.PingRequest{})
	if err != nil {
		err = fmt.Errorf("%w: %w", ErrSendingRequestFailed, err)
		rsc.logger.Error("Ping request failed", "error", err)
		return err
	}

	if _, ok := response.(*r.PingResponse); !ok {
		err = fmt.Errorf("%w: %T", ErrInvalidResponseType, response)
		rsc.logger.Error("Ping request failed", "error", err)
		return err
	}

	rsc.logger.Debug("Ping request succeeded")

	return nil
}

// RemoteSignerClient type implements fmt.Stringer.
var _ fmt.Stringer = (*RemoteSignerClient)(nil)

// String implements fmt.Stringer.
// Since this method requires a network request to get the server public key, it may
// be slow or fail. To mitigate this, the address is cached after the first request.
func (rsc *RemoteSignerClient) String() string {
	address := rsc.addrCache

	// If the address is not in the cache, get it from the server.
	if address == "" {
		address = "unknown"
		if pubKey, err := rsc.PubKey(); err == nil {
			address = pubKey.Address().String()
			rsc.addrCache = address // Save the address in the cache.
		}
	}

	return fmt.Sprintf("{Type: RemoteSigner, Addr: %s}", address)
}

// isClosed returns true if the client is closed.
func (rsc *RemoteSignerClient) isClosed() bool {
	return rsc.closed.Load()
}

// RemoteSignerClient type implements io.Closer.
var _ io.Closer = (*RemoteSignerClient)(nil)

// Close implements io.Closer.
func (rsc *RemoteSignerClient) Close() error {
	// Check if the client is already closed and set the closed state.
	if !rsc.closed.CompareAndSwap(false, true) {
		return ErrClientAlreadyClosed
	}

	// Close the connection.
	err := rsc.setConnection(nil)

	rsc.logger.Info("Client closed")

	return err
}
