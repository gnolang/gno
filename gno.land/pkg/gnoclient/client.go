package gnoclient

import (
	"errors"
)

type Client struct {
	Signer     Signer
	Networking Networking
}

func (c Client) validateSigner() error {
	if c.Signer == nil {
		return errors.New("missing c.Signer")
	}
	return nil
}

func (c Client) validateRPCClient() error {
	if c.Networking == nil {
		return errors.New("missing c.Networking")
	}
	return nil
}

// TODO: port existing code, i.e. faucet?
// TODO: create right now a tm2 generic go client and a gnovm generic go client?
// TODO: Command: Call
// TODO: Command: Send
// TODO: Command: AddPkg
// TODO: Command: Query
// TODO: Command: Eval
// TODO: Command: Exec
// TODO: Command: Package
// TODO: Command: QFile
// TODO: examples and unit tests
// TODO: Mock
// TODO: alternative configuration (pass existing websocket?)
// TODO: minimal go.mod to make it light to import
