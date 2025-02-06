package remote

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	p2pconn "github.com/gnolang/gno/tm2/pkg/p2p/conn"
)

var ErrInvalidRequest = errors.New("invalid request received")

type RemoteSignerServer struct {
	signer         types.Signer
	logger         *slog.Logger
	serverKey      crypto.PrivKey
	authorizedKeys []crypto.PubKey

	listeners []net.Listener
	conns     []net.Conn
}

func NewRemoteSignerServer(
	listenerAddresses []string,
	logger *slog.Logger,
	signer types.Signer,
) (*RemoteSignerServer, error) {
	rss := &RemoteSignerServer{
		signer: signer,
		logger: logger,
	}

	if len(listenerAddresses) == 0 {
		return nil, errors.New("no listen address provided")
	}

	rss.listeners = make([]net.Listener, len(listenerAddresses))

	for i := range listenerAddresses {
		protocol, address := osm.ProtocolAndAddress(listenerAddresses[i])
		listener, err := net.Listen(protocol, address)
		if err != nil {
			return nil, err
		}
		rss.listeners[i] = listener
	}

	return rss, nil
}

func (rss *RemoteSignerServer) Close() error {
	var errors []string

	for _, listener := range rss.listeners {
		if err := listener.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("%s://%s: %v", listener.Addr().Network(), listener.Addr().String(), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, ", "))
	}

	return nil
}

func (rss *RemoteSignerServer) Start() {
	for _, listener := range rss.listeners {
		go func(listener net.Listener) {
			for {
				conn, err := listener.Accept()
				if err != nil {
					rss.logger.Error("failed to accept connection", "err", err)
					continue
				}

				if listener.Addr().Network() != "unix" {
					conn, err = p2pconn.MakeSecretConnection(conn, rss.serverKey)
					if err != nil {
						rss.logger.Error("failed to make connection secret", "err", err)
						continue
					}
				}

				go func(conn net.Conn) {
					fmt.Println("conn", conn)
				}(conn)
			}
		}(listener)
	}
}

func (rss *RemoteSignerServer) handle(request RemoteSignerMessage) RemoteSignerMessage {
	switch r := request.(type) {
	case *PubKeyRequest:
		if pubKey, err := rss.signer.PubKey(); err != nil {
			return &PubKeyResponse{nil, &RemoteSignerError{0, err.Error()}}
		} else {
			return &PubKeyResponse{pubKey, nil}
		}

	case *SignRequest:
		if signature, err := rss.signer.Sign(r.SignBytes); err != nil {
			return &SignResponse{nil, &RemoteSignerError{0, err.Error()}}
		} else {
			return &SignResponse{signature, nil}
		}

	case *PingRequest:
		return &PingResponse{}
	}

	return nil
}
