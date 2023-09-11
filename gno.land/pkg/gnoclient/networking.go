package gnoclient

import "errors"

// Networking ...
type Networking interface {
	// write methods.

	// Broadcast sends an encoded and signed transaction message to the blockchain.
	// The default implementation connects on a node's Websocket interface.
	Broadcast(txbz []byte) error

	// TODO: asynchronous Broadcast
	// Broadcast(txbz []byte) (ch <-Event, error)

	// read methods.

	Query()
}

type WebsocketClient struct {
	// Opts
	Remote string // Remote RPC node
}

var _ Networking = (*WebsocketClient)(nil) // should implement Networking

func (wsc WebsocketClient) ValidateOpts() error {
	if wsc.Remote == "" {
		return errors.New("missing remote url")
	}

	return nil
}

func (wsc WebsocketClient) Broadcast(txbz []byte) error {
	return errors.New("not implemented")
}

func (wsc WebsocketClient) Query() {

}
