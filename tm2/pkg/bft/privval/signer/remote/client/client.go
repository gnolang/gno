package client

import (
	"fmt"

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

// isClosed returns true if the client is closed.
func (rsc *RemoteSignerClient) isClosed() bool {
	return rsc.closed.Load()
}

// Close closes the connection to the server and stops the client.
func (rsc *RemoteSignerClient) Close() error {
	// Check if the client is already closed and set the closed state.
	if !rsc.closed.CompareAndSwap(false, true) {
		return ErrClientAlreadyClosed
	}

	// Close the connection.
	return rsc.setConnection(nil)
}
