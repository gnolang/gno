package mempool

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

const Channel = byte(0x30)

const maxIDsPerMsg = 1024 // hard limit for message sizes

var (
	errNoTx         = errors.New("transaction is nil")
	errNoIDs        = errors.New("no transaction IDs in message")
	errExcessiveIDs = errors.New("too many IDs in message")
)

// Reactor handles mempool tx broadcasting amongst peers
type Reactor struct {
	p2p.BaseReactor

	gossipEnabled bool // flag indicating if txs should be propagated
	mempool       *CListMempool
}

// NewReactor returns a new Reactor with the given config and mempool.
func NewReactor(gossipEnabled bool, mempool *CListMempool) *Reactor {
	memR := &Reactor{
		gossipEnabled: gossipEnabled,
		mempool:       mempool,
	}

	memR.BaseReactor = *p2p.NewBaseReactor("Reactor", memR)

	return memR
}

// SetLogger sets the Logger on the reactor and the underlying mempool.
func (memR *Reactor) SetLogger(l *slog.Logger) {
	memR.Logger = l
	memR.mempool.SetLogger(l)
}

// OnStart implements p2p.BaseReactor.
func (memR *Reactor) OnStart() error {
	if !memR.gossipEnabled {
		memR.Logger.Info("Tx gossiping is disabled – receive‑only mode")
	}

	return nil
}

// GetChannels implements Reactor.
// It returns the list of channels for this reactor.
func (memR *Reactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:       Channel,
			Priority: 5,
		},
	}
}

// AddPeer implements Reactor.
// It starts a broadcast routine ensuring all txs are forwarded to the given peer.
func (memR *Reactor) AddPeer(peer p2p.PeerConn) {
	if !memR.gossipEnabled {
		return
	}

	go memR.broadcastTxRoutine(peer)
}

// RemovePeer implements Reactor.
func (memR *Reactor) RemovePeer(_ p2p.PeerConn, _ any) {}

// Receive implements Reactor.
// It adds any received transactions to the mempool.
func (memR *Reactor) Receive(chID byte, peer p2p.PeerConn, msgBytes []byte) {
	memR.Logger.Debug(
		"received message",
		"peerID", peer.ID(),
		"chID", chID,
	)

	// Unmarshal the message
	var msg Message

	if err := amino.UnmarshalAny(msgBytes, &msg); err != nil {
		memR.Logger.Error("unable to unmarshal mempool message", "err", err)

		return
	}

	// Validate the message
	if err := msg.ValidateBasic(); err != nil {
		memR.Logger.Warn("unable to validate mempool message", "err", err)

		return
	}

	switch msg := msg.(type) {
	case *TxMessage:
		if err := memR.handleTxRequest(msg.Tx); err != nil {
			memR.Logger.Warn("unable to handle tx request", "err", err)
		}
	case *IHaveMessage:
		if err := memR.handleHaveRequest(peer, msg.IDs); err != nil {
			memR.Logger.Warn("unable to handle have request", "err", err)
		}
	case *IWantMessage:
		if err := memR.handleWantRequest(peer, msg.IDs); err != nil {
			memR.Logger.Warn("unable to handle want request", "err", err)
		}
	default:
		memR.Logger.Warn("invalid message received", "type", reflect.TypeOf(msg))
	}
}

// handleTxRequest handles a transaction add request
func (memR *Reactor) handleTxRequest(tx types.Tx) error {
	if memR.hasTx(tx.Hash()) {
		return nil // duplicate
	}

	// Add the tx to the mempool
	return memR.mempool.CheckTx(tx, nil)
}

// handleHaveRequest requests unknown transactions from the peer, if any
func (memR *Reactor) handleHaveRequest(peer p2p.PeerConn, ids [][]byte) error {
	var unknownIDs [][]byte

	for _, id := range ids {
		if memR.hasTx(id) {
			continue
		}

		unknownIDs = append(unknownIDs, id)
	}

	if len(unknownIDs) == 0 {
		// No unknown transactions
		return nil
	}

	// Create the request, and marshal it to Amino binary
	req := &IWantMessage{
		IDs: unknownIDs,
	}

	preparedReq, err := amino.MarshalAny(req)
	if err != nil {
		return fmt.Errorf("unable to marshal want request: %w", err)
	}

	if !peer.Send(Channel, preparedReq) {
		return fmt.Errorf("unable to send want request to peer %s", peer.ID())
	}

	return nil
}

// handleWantRequest prepares the list of requested transactions, in case
// they are in the node's mempool
func (memR *Reactor) handleWantRequest(peer p2p.PeerConn, ids [][]byte) error {
	for _, id := range ids {
		if tx, ok := memR.getTx(id); ok {
			// Create the response, and marshal it to Amino binary
			resp := &TxMessage{
				Tx: tx,
			}

			preparedResp, err := amino.MarshalAny(resp)
			if err != nil {
				return fmt.Errorf("unable to marshal want response: %w", err)
			}

			if !peer.Send(Channel, preparedResp) {
				return fmt.Errorf("unable to send want response to peer %s", peer.ID())
			}
		}
	}

	return nil
}

// Send new mempool txs to peer.
func (memR *Reactor) broadcastTxRoutine(peer p2p.PeerConn) {
	for {
		if !memR.IsRunning() || !peer.IsRunning() {
			return
		}

		select {
		case <-memR.mempool.TxsWaitChan():
			if elem := memR.mempool.TxsFront(); elem != nil {
				memTx := elem.Value.(*mempoolTx)

				// Announce the tx to the peer
				// Create the request, and marshal it to Amino binary
				req := &IHaveMessage{
					IDs: [][]byte{memTx.tx.Hash()},
				}

				preparedReq, err := amino.MarshalAny(req)
				if err != nil {
					memR.Logger.Warn("unable to marshal have request", "err", err)

					continue
				}

				if !peer.Send(Channel, preparedReq) {
					memR.Logger.Warn("unable to send have request to peer", "peer", peer.ID())
				}
			}
		case <-peer.Quit():
			return
		case <-memR.Quit():
			return
		}
	}
}

// Message is the wrapper for the mempool message
type Message interface {
	ValidateBasic() error
}

type (
	TxMessage    struct{ Tx types.Tx }
	IHaveMessage struct{ IDs [][]byte }
	IWantMessage struct{ IDs [][]byte }
)

func (m *TxMessage) ValidateBasic() error {
	if m.Tx == nil {
		return errNoTx
	}

	return nil
}

func (m *TxMessage) String() string {
	return fmt.Sprintf("[TxMessage %X]", m.Tx)
}

func (m *IHaveMessage) ValidateBasic() error {
	if len(m.IDs) == 0 {
		return errNoIDs
	}

	if len(m.IDs) > maxIDsPerMsg {
		return errExcessiveIDs
	}

	return nil
}

func (m *IHaveMessage) String() string {
	return fmt.Sprintf("[IHaveMessage %d]", len(m.IDs))
}

func (m *IWantMessage) ValidateBasic() error {
	if len(m.IDs) == 0 {
		return errNoIDs
	}

	if len(m.IDs) > maxIDsPerMsg {
		return errExcessiveIDs
	}

	return nil
}

func (m *IWantMessage) String() string {
	return fmt.Sprintf("[IWantMessage %d]", len(m.IDs))
}
