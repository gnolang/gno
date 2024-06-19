package config

import (
	"time"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

// -----------------------------------------------------------------------------
// P2PConfig

const (
	// FuzzModeDrop is a mode in which we randomly drop reads/writes, connections or sleep
	FuzzModeDrop = iota
	// FuzzModeDelay is a mode in which we randomly sleep
	FuzzModeDelay
)

// P2PConfig defines the configuration options for the Tendermint peer-to-peer networking layer
type P2PConfig struct {
	RootDir string `toml:"home"`

	// Address to listen for incoming connections
	ListenAddress string `toml:"laddr" comment:"Address to listen for incoming connections"`

	// Address to advertise to peers for them to dial
	ExternalAddress string `toml:"external_address" comment:"Address to advertise to peers for them to dial\n If empty, will use the same port as the laddr,\n and will introspect on the listener or use UPnP\n to figure out the address."`

	// Comma separated list of seed nodes to connect to
	Seeds string `toml:"seeds" comment:"Comma separated list of seed nodes to connect to"`

	// Comma separated list of nodes to keep persistent connections to
	PersistentPeers string `toml:"persistent_peers" comment:"Comma separated list of nodes to keep persistent connections to"`

	// UPNP port forwarding
	UPNP bool `toml:"upnp" comment:"UPNP port forwarding"`

	// Maximum number of inbound peers
	MaxNumInboundPeers int `toml:"max_num_inbound_peers" comment:"Maximum number of inbound peers"`

	// Maximum number of outbound peers to connect to, excluding persistent peers
	MaxNumOutboundPeers int `toml:"max_num_outbound_peers" comment:"Maximum number of outbound peers to connect to, excluding persistent peers"`

	// Time to wait before flushing messages out on the connection
	FlushThrottleTimeout time.Duration `toml:"flush_throttle_timeout" comment:"Time to wait before flushing messages out on the connection"`

	// Maximum size of a message packet payload, in bytes
	MaxPacketMsgPayloadSize int `toml:"max_packet_msg_payload_size" comment:"Maximum size of a message packet payload, in bytes"`

	// Rate at which packets can be sent, in bytes/second
	SendRate int64 `toml:"send_rate" comment:"Rate at which packets can be sent, in bytes/second"`

	// Rate at which packets can be received, in bytes/second
	RecvRate int64 `toml:"recv_rate" comment:"Rate at which packets can be received, in bytes/second"`

	// Set true to enable the peer-exchange reactor
	PexReactor bool `toml:"pex" comment:"Set true to enable the peer-exchange reactor"`

	// Seed mode, in which node constantly crawls the network and looks for
	// peers. If another node asks it for addresses, it responds and disconnects.
	//
	// Does not work if the peer-exchange reactor is disabled.
	SeedMode bool `toml:"seed_mode" comment:"Seed mode, in which node constantly crawls the network and looks for\n peers. If another node asks it for addresses, it responds and disconnects.\n\n Does not work if the peer-exchange reactor is disabled."`

	// Comma separated list of peer IDs to keep private (will not be gossiped to
	// other peers)
	PrivatePeerIDs string `toml:"private_peer_ids" comment:"Comma separated list of peer IDs to keep private (will not be gossiped to other peers)"`

	// Toggle to disable guard against peers connecting from the same ip.
	AllowDuplicateIP bool `toml:"allow_duplicate_ip" comment:"Toggle to disable guard against peers connecting from the same ip."`

	// Peer connection configuration.
	HandshakeTimeout time.Duration `toml:"handshake_timeout" comment:"Peer connection configuration."`
	DialTimeout      time.Duration `toml:"dial_timeout"`

	// Testing params.
	// Force dial to fail
	TestDialFail bool `toml:"test_dial_fail"`
	// FUzz connection
	TestFuzz       bool            `toml:"test_fuzz"`
	TestFuzzConfig *FuzzConnConfig `toml:"test_fuzz_config"`
}

// DefaultP2PConfig returns a default configuration for the peer-to-peer layer
func DefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		ListenAddress:           "tcp://0.0.0.0:26656",
		ExternalAddress:         "",
		UPNP:                    false,
		MaxNumInboundPeers:      40,
		MaxNumOutboundPeers:     10,
		FlushThrottleTimeout:    100 * time.Millisecond,
		MaxPacketMsgPayloadSize: 1024,    // 1 kB
		SendRate:                5120000, // 5 mB/s
		RecvRate:                5120000, // 5 mB/s
		PexReactor:              true,
		SeedMode:                false,
		AllowDuplicateIP:        false,
		HandshakeTimeout:        20 * time.Second,
		DialTimeout:             3 * time.Second,
		TestDialFail:            false,
		TestFuzz:                false,
		TestFuzzConfig:          DefaultFuzzConnConfig(),
	}
}

// TestP2PConfig returns a configuration for testing the peer-to-peer layer
func TestP2PConfig() *P2PConfig {
	cfg := DefaultP2PConfig()
	cfg.ListenAddress = "tcp://0.0.0.0:26656"
	cfg.FlushThrottleTimeout = 10 * time.Millisecond
	cfg.AllowDuplicateIP = true
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

// FuzzConnConfig is a FuzzedConnection configuration.
type FuzzConnConfig struct {
	Mode         int
	MaxDelay     time.Duration
	ProbDropRW   float64
	ProbDropConn float64
	ProbSleep    float64
}

// DefaultFuzzConnConfig returns the default config.
func DefaultFuzzConnConfig() *FuzzConnConfig {
	return &FuzzConnConfig{
		Mode:         FuzzModeDrop,
		MaxDelay:     3 * time.Second,
		ProbDropRW:   0.2,
		ProbDropConn: 0.00,
		ProbSleep:    0.00,
	}
}
