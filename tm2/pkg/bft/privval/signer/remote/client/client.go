package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	r "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/remote"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

// RemoteSignerClient implements types.Signer by connecting to a RemoteSignerServer.
type RemoteSignerClient struct {
	// Required config.
	ctx      context.Context
	protocol string
	address  string
	logger   *slog.Logger

	// Optional connection config.
	dialMaxRetries    int // If -1, retry indefinitely.
	dialRetryInterval time.Duration
	dialTimeout       time.Duration // If 0, no timeout is set.
	keepAlivePeriod   time.Duration // If 0, keep alive is disabled.
	requestTimeout    time.Duration // If 0, no timeout is set.

	// Optional authentication config.
	clientPrivKey  ed25519.PrivKeyEd25519  // Default is a random key.
	authorizedKeys []ed25519.PubKeyEd25519 // If empty, all keys are authorized.

	// Internal.
	conn         net.Conn
	connLock     sync.RWMutex
	dialer       net.Dialer
	cachedPubKey crypto.PubKey
	cancelCtx    context.CancelFunc
}

// RemoteSignerClient type implements types.Signer.
var _ types.Signer = (*RemoteSignerClient)(nil)

// PubKey implements types.Signer.
func (rsc *RemoteSignerClient) PubKey() crypto.PubKey {
	return rsc.cachedPubKey
}

// cachePubKey sends a PubKey request to the server and caches the response.
// This method is called only once when the client is created.
func (rsc *RemoteSignerClient) cachePubKey() error {
	response, err := rsc.send(&r.PubKeyRequest{})
	if err != nil {
		err := fmt.Errorf("%w: %w", ErrSendingRequestFailed, err)
		if !errors.Is(err, ErrClientAlreadyClosed) {
			rsc.logger.Error("PubKey request failed", "error", err)
		}
		return err
	}

	pubKeyResponse, ok := response.(*r.PubKeyResponse)
	if !ok {
		err := fmt.Errorf("%w: %T", ErrInvalidResponseType, response)
		rsc.logger.Error("PubKey request failed", "error", err)
		return err
	}

	// Save the address in the cache for the String method.
	rsc.cachedPubKey = pubKeyResponse.PubKey

	return nil
}

// Sign implements types.Signer.
func (rsc *RemoteSignerClient) Sign(signBytes []byte) ([]byte, error) {
	response, err := rsc.send(&r.SignRequest{SignBytes: signBytes})
	if err != nil {
		err := fmt.Errorf("%w: %w", ErrSendingRequestFailed, err)
		if !errors.Is(err, ErrClientAlreadyClosed) {
			rsc.logger.Error("Sign request failed", "error", err)
		}
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

// Close implements type.Signer.
func (rsc *RemoteSignerClient) Close() error {
	// Check if the client is already closed and set the closed state.
	if rsc.ctx.Err() != nil {
		return ErrClientAlreadyClosed
	}

	// Cancel the dial context.
	rsc.cancelCtx()

	// Close the connection.
	err := rsc.setConnection(nil)

	rsc.logger.Info("Client closed")

	return err
}

// RemoteSignerClient type implements fmt.Stringer.
var _ fmt.Stringer = (*RemoteSignerClient)(nil)

// String implements fmt.Stringer.
func (rsc *RemoteSignerClient) String() string {
	return fmt.Sprintf("{Type: RemoteSigner, Addr: %s}", rsc.cachedPubKey.Address())
}
