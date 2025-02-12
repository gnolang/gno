package remote

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// RemoteSignerClient implements types.Signer by connecting to a RemoteSignerServer.
type RemoteSignerClient struct {
	address         string
	conn            net.Conn
	maxDialAttempts int
	logger          *slog.Logger
}

// RemoteSignerClient type implements types.Signer.
var _ types.Signer = (*RemoteSignerClient)(nil)

// PubKey implements types.Signer.
func (rsc *RemoteSignerClient) PubKey() (crypto.PubKey, error) {
	response, err := rsc.send(&PubKeyRequest{})
	if err != nil {
		rsc.logger.Error("unable to send public key request", "err", err)
		return nil, fmt.Errorf("send request failed: %w", err)
	}

	pubKeyResponse, ok := response.(*PubKeyResponse)
	if !ok {
		rsc.logger.Error("wrong response type on public key request", "response", response)
		return nil, fmt.Errorf("wrong response type: %T", response)
	}

	if pubKeyResponse.Error != nil {
		rsc.logger.Error("server returned error on public key request", "err", pubKeyResponse.Error)
		return nil, fmt.Errorf("response contains error: %w", pubKeyResponse.Error)
	}

	return pubKeyResponse.PubKey, nil
}

// Sign implements types.Signer.
func (rsc *RemoteSignerClient) Sign(signBytes []byte) ([]byte, error) {
	response, err := rsc.send(&SignRequest{SignBytes: signBytes})
	if err != nil {
		rsc.logger.Error("unable to send sign request", "err", err)
		return nil, fmt.Errorf("send request failed: %w", err)
	}

	signResponse, ok := response.(*SignResponse)
	if !ok {
		rsc.logger.Error("wrong response type on sign request", "response", response)
		return nil, fmt.Errorf("wrong response type: %T", response)
	}

	if signResponse.Error != nil {
		rsc.logger.Error("server returned error on sign request", "err", signResponse.Error)
		return nil, fmt.Errorf("response contains error: %w", signResponse.Error)
	}

	return signResponse.Signature, nil
}

// Ping sends a ping request to the remote signer.
func (rsc *RemoteSignerClient) Ping() error {
	response, err := rsc.send(&PingRequest{})
	if err != nil {
		rsc.logger.Error("unable to send ping request", "err", err)
		return fmt.Errorf("send request failed: %w", err)
	}

	if _, ok := response.(*PingResponse); !ok {
		rsc.logger.Error("wrong response type on sign request", "response", response)
		return fmt.Errorf("wrong response type: %T", response)
	}

	return nil
}

func (rsc *RemoteSignerClient) send(request RemoteSignerMessage) (RemoteSignerMessage, error) {
	if _, err := amino.MarshalAnySized(request); err != nil {
		return nil, err
	}

	const maxResponseSize = 1024 * 10
	var response RemoteSignerMessage

	// rsc.conn.Read(b []byte)
	// if _, err := amino.UnmarshalSized(
	// 	return nil, err
	// }

	return response, nil
}

// Close closes the underlying connection.
func (rsc *RemoteSignerClient) Close() error {
	return rsc.conn.Close()
}

func NewRemoteSignerClient(address string) (*RemoteSignerClient, error) {
	return &RemoteSignerClient{address: address}, nil
}
