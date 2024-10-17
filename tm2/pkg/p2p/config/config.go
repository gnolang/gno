package config

import (
	"time"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

// P2PConfig defines the configuration options for the Tendermint peer-to-peer networking layer
type P2PConfig struct {
	RootDir string `json:"rpc" toml:"home"`

	// Address to listen for incoming connections
	ListenAddress string `json:"laddr" toml:"laddr" comment:"Address to listen for incoming connections"`

	// Address to advertise to peers for them to dial
	ExternalAddress string `json:"external_address" toml:"external_address" comment:"Address to advertise to peers for them to dial\n If empty, will use the same port as the laddr,\n and will introspect on the listener or use UPnP\n to figure out the address."`

	// Comma separated list of seed nodes to connect to
	Seeds string `json:"seeds" toml:"seeds" comment:"Comma separated list of seed nodes to connect to"`

	// Comma separated list of nodes to keep persistent connections to
	PersistentPeers string `json:"persistent_peers" toml:"persistent_peers" comment:"Comma separated list of nodes to keep persistent connections to"`

	// Maximum number of inbound peers
	MaxNumInboundPeers uint64 `json:"max_num_inbound_peers" toml:"max_num_inbound_peers" comment:"Maximum number of inbound peers"`

	// Maximum number of outbound peers to connect to, excluding persistent peers
	MaxNumOutboundPeers uint64 `json:"max_num_outbound_peers" toml:"max_num_outbound_peers" comment:"Maximum number of outbound peers to connect to, excluding persistent peers"`

	// Time to wait before flushing messages out on the connection
	FlushThrottleTimeout time.Duration `json:"flush_throttle_timeout" toml:"flush_throttle_timeout" comment:"Time to wait before flushing messages out on the connection"`

	// Maximum size of a message packet payload, in bytes
	MaxPacketMsgPayloadSize int `json:"max_packet_msg_payload_size" toml:"max_packet_msg_payload_size" comment:"Maximum size of a message packet payload, in bytes"`

	// Rate at which packets can be sent, in bytes/second
	SendRate int64 `json:"send_rate" toml:"send_rate" comment:"Rate at which packets can be sent, in bytes/second"`

	// Rate at which packets can be received, in bytes/second
	RecvRate int64 `json:"recv_rate" toml:"recv_rate" comment:"Rate at which packets can be received, in bytes/second"`

	// Set true to enable the peer-exchange reactor
	PeerExchange bool `json:"pex" toml:"pex" comment:"Set true to enable the peer-exchange reactor"`

	// Comma separated list of peer IDs to keep private (will not be gossiped to other peers)
	PrivatePeerIDs string `json:"private_peer_ids" toml:"private_peer_ids" comment:"Comma separated list of peer IDs to keep private (will not be gossiped to other peers)"`
}

// DefaultP2PConfig returns a default configuration for the peer-to-peer layer
func DefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		ListenAddress:           "tcp://0.0.0.0:26656",
		ExternalAddress:         "",
		MaxNumInboundPeers:      40,
		MaxNumOutboundPeers:     10,
		FlushThrottleTimeout:    100 * time.Millisecond,
		MaxPacketMsgPayloadSize: 1024,    // 1 kB
		SendRate:                5120000, // 5 mB/s
		RecvRate:                5120000, // 5 mB/s
		PeerExchange:            true,
	}
}

// TestP2PConfig returns a configuration for testing the peer-to-peer layer
func TestP2PConfig() *P2PConfig {
	cfg := DefaultP2PConfig()
	cfg.ListenAddress = "tcp://0.0.0.0:26656"
	cfg.FlushThrottleTimeout = 10 * time.Millisecond

	return cfg
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *P2PConfig) ValidateBasic() error {
	if cfg.MaxNumInboundPeers < 0 {
		return errors.New("max_num_inbound_peers can't be negative")
	}
	if cfg.MaxNumOutboundPeers < 0 {
		return errors.New("max_num_outbound_peers can't be negative")
	}
	if cfg.FlushThrottleTimeout < 0 {
		return errors.New("flush_throttle_timeout can't be negative")
	}
	if cfg.MaxPacketMsgPayloadSize < 0 {
		return errors.New("max_packet_msg_payload_size can't be negative")
	}
	if cfg.SendRate < 0 {
		return errors.New("send_rate can't be negative")
	}
	if cfg.RecvRate < 0 {
		return errors.New("recv_rate can't be negative")
	}
	return nil
}
