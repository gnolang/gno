package gnoclient

import (
	"errors"

	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
)

type Client struct {
	Keybase    keys.Keybase
	Networking Networking
}

func (c Client) validateSigner() error {
	if c.Keybase == nil {
		return errors.New("missing c.Keybase")
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
