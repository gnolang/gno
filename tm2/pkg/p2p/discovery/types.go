package discovery

import (
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

var errNoPeers = errors.New("no peers received")

// Message is the wrapper for the discovery message
type Message interface {
	ValidateBasic() error
}

// Request is the peer discovery request.
// It is empty by design, since it's used as
// a notification type
type Request struct{}

func (r *Request) ValidateBasic() error {
	return nil
}

// Response is the peer discovery response
type Response struct {
	Peers []*types.NetAddress // the peer set returned by the peer
}

func (r *Response) ValidateBasic() error {
	// Make sure at least some peers were received
	if len(r.Peers) == 0 {
		return errNoPeers
	}

	// Make sure the returned peer dial
	// addresses are valid
	for _, peer := range r.Peers {
		if err := peer.Validate(); err != nil {
			return err
		}
	}

	return nil
}
