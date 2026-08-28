package config

import (
	"errors"
	"path/filepath"
	"time"
)

var (
	ErrInvalidFlushThrottleTimeout = errors.New("invalid flush throttle timeout")
	ErrInvalidMaxPayloadSize       = errors.New("invalid message payload size")
	ErrInvalidSendRate             = errors.New("invalid packet send rate")
	ErrInvalidReceiveRate          = errors.New("invalid packet receive rate")
	ErrInvalidRecvAssemblyTimeout  = errors.New("invalid receive assembly timeout")
	ErrInvalidMaxRecvBufferBytes   = errors.New("invalid max receive buffer size")
)

// defaultAddrBookPath is the default relative path for the persisted peer address book
var defaultAddrBookPath = "config/addrbook.json"

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

	// Maximum time a single incoming message may spend being assembled from partial packets
	RecvAssemblyTimeout time.Duration `json:"recv_assembly_timeout" toml:"recv_assembly_timeout" comment:"Maximum time a single incoming message may spend being assembled from partial packets.\n The deadline is anchored to the first partial packet of a message and is not extended by later ones,\n so a peer cannot hold buffer space indefinitely by dribbling packets. 0 disables the deadline."`

	// Maximum total bytes buffered for incomplete messages, across all of a connection's channels
	MaxRecvBufferBytes int `json:"max_recv_buffer_bytes" toml:"max_recv_buffer_bytes" comment:"Maximum total bytes buffered for incomplete incoming messages, across all of a connection's channels.\n This is the per-connection bound on memory a peer can pin, so the node's worst case is this times the peer limits.\n 0 disables the budget, leaving only each channel's own capacity."`

	// Set true to enable the peer-exchange reactor
	PeerExchange bool `json:"pex" toml:"pex" comment:"Set true to enable the peer-exchange reactor"`

	// Comma separated list of peer IDs to keep private (will not be gossiped to other peers)
	PrivatePeerIDs string `json:"private_peer_ids" toml:"private_peer_ids" comment:"Comma separated list of peer IDs to keep private (will not be gossiped to other peers)"`

	// Toggle to disable guard against peers connecting from the same ip
	AllowDuplicateIP bool `json:"allow_duplicate_ip" toml:"allow_duplicate_ip" comment:"Toggle to disable guard against peers connecting from the same ip"`

	// Path to the address book file used to persist discovered peers across restarts.
	// When empty, a default path relative to the root directory is used.
	AddrBook string `json:"addr_book_file" toml:"addr_book_file" comment:"Path to the address book file used to persist discovered peers across restarts"`
}

// DefaultP2PConfig returns a default configuration for the peer-to-peer layer
func DefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		ListenAddress:           "tcp://0.0.0.0:26656",
		ExternalAddress:         "", // nothing is advertised differently
		MaxNumInboundPeers:      40,
		MaxNumOutboundPeers:     10,
		FlushThrottleTimeout:    100 * time.Millisecond,
		MaxPacketMsgPayloadSize: 1024,    // 1 kB
		SendRate:                5120000, // 5 mB/s
		RecvRate:                5120000, // 5 mB/s
		RecvAssemblyTimeout:     30 * time.Second,
		MaxRecvBufferBytes:      20 << 20, // 20MB
		PeerExchange:            true,
		AllowDuplicateIP:        false,
		AddrBook:                defaultAddrBookPath,
	}
}

// ValidateBasic performs basic validation (checking param bounds, etc.) and
// returns an error if any check fails.
func (cfg *P2PConfig) ValidateBasic() error {
	if cfg.FlushThrottleTimeout < 0 {
		return ErrInvalidFlushThrottleTimeout
	}

	if cfg.MaxPacketMsgPayloadSize < 0 {
		return ErrInvalidMaxPayloadSize
	}

	if cfg.SendRate < 0 {
		return ErrInvalidSendRate
	}

	if cfg.RecvRate < 0 {
		return ErrInvalidReceiveRate
	}

	if cfg.RecvAssemblyTimeout < 0 {
		return ErrInvalidRecvAssemblyTimeout
	}

	if cfg.MaxRecvBufferBytes < 0 {
		return ErrInvalidMaxRecvBufferBytes
	}

	return nil
}

// AddrBookFile returns the absolute path to the address book file.
// When AddrBook is a relative path, it is resolved against RootDir.
// When AddrBook is empty, the default path is used.
func (cfg *P2PConfig) AddrBookFile() string {
	path := cfg.AddrBook
	if path == "" {
		path = defaultAddrBookPath
	}

	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(cfg.RootDir, path)
}
