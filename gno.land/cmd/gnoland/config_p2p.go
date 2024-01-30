package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type configP2PCfg struct {
	commonEditCfg

	rootDir                 string
	listenAddress           string
	externalAddress         string
	seeds                   string
	persistentPeers         string
	upnp                    string // toggle
	maxNumInboundPeers      int
	maxNumOutboundPeers     int
	flushThrottleTimeout    time.Duration
	maxPacketMsgPayloadSize int
	sendRate                int64
	recvRate                int64
	pexReactor              string // toggle
	seedMode                string // toggle
	privatePeerIDs          string
	allowDuplicateIP        string // toggle
	handshakeTimeout        time.Duration
	dialTimeout             time.Duration
}

// newConfigP2PCmd creates the new config p2p command
func newConfigP2PCmd(io commands.IO) *commands.Command {
	cfg := &configP2PCfg{}

	cmd := commands.NewCommand(
		commands.Metadata{
			Name:       "p2p",
			ShortUsage: "config p2p [flags]",
			ShortHelp:  "Edits the Gno node's p2p configuration",
			LongHelp:   "Edits the Gno node's p2p configuration locally",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execConfigP2P(cfg, io)
		},
	)

	return cmd
}

func (c *configP2PCfg) RegisterFlags(fs *flag.FlagSet) {
	c.commonEditCfg.RegisterFlags(fs)

	fs.StringVar(
		&c.rootDir,
		"root-dir",
		"",
		"the root directory for all p2p data",
	)

	fs.StringVar(
		&c.listenAddress,
		"listen-address",
		"",
		"the address to listen on for incoming connections",
	)

	fs.StringVar(
		&c.externalAddress,
		"external-address",
		"",
		"the address to advertise to peers for them to dial",
	)

	fs.StringVar(
		&c.seeds,
		"seeds",
		"",
		"comma separated list of seed nodes to connect to",
	)

	fs.StringVar(
		&c.persistentPeers,
		"persistent-peers",
		"",
		"comma separated list of nodes to keep persistent connections to",
	)

	fs.StringVar(
		&c.upnp,
		"upnp",
		offValue,
		fmt.Sprintf(
			"toggle indicating if UPNP port forwarding should be enabled: %s | %s",
			onValue,
			offValue,
		),
	)

	fs.IntVar(
		&c.maxNumInboundPeers,
		"max-inbound-peers",
		-1,
		"maximum number of inbound peers",
	)

	fs.IntVar(
		&c.maxNumOutboundPeers,
		"max-outbound-peers",
		-1,
		"maximum number of outbound peers",
	)

	fs.DurationVar(
		&c.flushThrottleTimeout,
		"flush-throttle-timeout",
		time.Second*0,
		"time to wait before flushing messages out on the connection",
	)

	fs.IntVar(
		&c.maxPacketMsgPayloadSize,
		"max-message-payload",
		-1,
		"the maximum size of a message packet payload, in bytes",
	)

	fs.Int64Var(
		&c.sendRate,
		"send-rate",
		-1,
		"the rate at which packets can be sent, in bytes/second",
	)

	fs.Int64Var(
		&c.recvRate,
		"receive-rate",
		-1,
		"the rate at which packets can be received, in bytes/second",
	)

	fs.StringVar(
		&c.pexReactor,
		"pex-reactor",
		onValue,
		fmt.Sprintf(
			"value indicating if the peer-exchange reactor should be enabled: %s | %s",
			onValue,
			offValue,
		),
	)

	fs.StringVar(
		&c.seedMode,
		"seed-mode",
		offValue,
		fmt.Sprintf(
			"value indicating if the seed mode should be enabled: %s | %s",
			onValue,
			offValue,
		),
	)

	fs.StringVar(
		&c.privatePeerIDs,
		"private-peer-ids",
		"",
		"comma separated list of peer IDs to keep private (will not be gossiped to other peers)",
	)

	fs.StringVar(
		&c.allowDuplicateIP,
		"allow-duplicate-ip",
		offValue,
		fmt.Sprintf(
			"toggle to disable guard against peers connecting from the same ip: %s | %s",
			onValue,
			offValue,
		),
	)

	fs.DurationVar(
		&c.handshakeTimeout,
		"handshake-timeout",
		time.Second*0,
		"the handshake process timeout",
	)

	fs.DurationVar(
		&c.dialTimeout,
		"dial-timeout",
		time.Second*0,
		"the dial process timeout",
	)
}

func execConfigP2P(cfg *configP2PCfg, io commands.IO) error {
	// Load the config
	loadedCfg, err := config.LoadConfigFile(cfg.configPath)
	if err != nil {
		return fmt.Errorf("unable to load config, %w", err)
	}

	// Set the root dir, if any
	if cfg.rootDir != "" {
		loadedCfg.P2P.RootDir = cfg.rootDir
	}

	// Set the listen address, if any
	if cfg.listenAddress != "" {
		loadedCfg.P2P.ListenAddress = cfg.listenAddress
	}

	// Set the external address, if any
	if cfg.externalAddress != "" {
		loadedCfg.P2P.ExternalAddress = cfg.externalAddress
	}

	// Set the seeds, if any
	if cfg.seeds != "" {
		loadedCfg.P2P.Seeds = cfg.seeds
	}

	// Set the persistent peers, if any
	if cfg.persistentPeers != "" {
		loadedCfg.P2P.PersistentPeers = cfg.persistentPeers
	}

	// Set the upnp toggle, if any
	upnpVal, err := parseToggleValue(cfg.upnp)
	if err != nil {
		return err
	}

	if upnpVal != loadedCfg.P2P.UPNP {
		loadedCfg.P2P.UPNP = upnpVal
	}

	// Set the max inbound peers, if any
	if cfg.maxNumInboundPeers >= 0 {
		loadedCfg.P2P.MaxNumInboundPeers = cfg.maxNumInboundPeers
	}

	// Set the max outbound peers, if any
	if cfg.maxNumOutboundPeers >= 0 {
		loadedCfg.P2P.MaxNumOutboundPeers = cfg.maxNumOutboundPeers
	}

	// Set the flush throttle timeout, if any
	if cfg.flushThrottleTimeout >= 0 {
		loadedCfg.P2P.FlushThrottleTimeout = cfg.flushThrottleTimeout
	}

	// Set the max package payload size, if any
	if cfg.maxPacketMsgPayloadSize >= 0 {
		loadedCfg.P2P.MaxPacketMsgPayloadSize = cfg.maxPacketMsgPayloadSize
	}

	// Set the send rate, if any
	if cfg.sendRate >= 0 {
		loadedCfg.P2P.SendRate = cfg.sendRate
	}

	// Set the receive rate, if any
	if cfg.recvRate >= 0 {
		loadedCfg.P2P.RecvRate = cfg.recvRate
	}

	// Set the pex reactor toggle, if any
	pexVal, err := parseToggleValue(cfg.pexReactor)
	if err != nil {
		return err
	}

	if pexVal != loadedCfg.P2P.PexReactor {
		loadedCfg.P2P.PexReactor = pexVal
	}

	// Set the seed mode toggle, if any
	seedModeVal, err := parseToggleValue(cfg.seedMode)
	if err != nil {
		return err
	}

	if seedModeVal != loadedCfg.P2P.SeedMode {
		loadedCfg.P2P.SeedMode = seedModeVal
	}

	// Set the private peer IDs, if any
	if cfg.privatePeerIDs != "" {
		loadedCfg.P2P.PrivatePeerIDs = cfg.privatePeerIDs
	}

	// Set the allow duplicate IPs toggle, if any
	allowDuplicateIPVal, err := parseToggleValue(cfg.allowDuplicateIP)
	if err != nil {
		return err
	}

	if allowDuplicateIPVal != loadedCfg.P2P.AllowDuplicateIP {
		loadedCfg.P2P.AllowDuplicateIP = allowDuplicateIPVal
	}

	// Set the handshake timeout, if any
	if cfg.handshakeTimeout != time.Second*0 {
		loadedCfg.P2P.HandshakeTimeout = cfg.handshakeTimeout
	}

	// Set the handshake timeout, if any
	if cfg.dialTimeout != time.Second*0 {
		loadedCfg.P2P.DialTimeout = cfg.dialTimeout
	}

	// Make sure the config is now valid
	if err := loadedCfg.ValidateBasic(); err != nil {
		return fmt.Errorf("unable to validate config, %w", err)
	}

	// Save the config
	if err := config.WriteConfigFile(cfg.configPath, loadedCfg); err != nil {
		return fmt.Errorf("unable to save updated config, %w", err)
	}

	io.Printfln("Updated P2P configuration saved at %s", cfg.configPath)

	return nil
}
