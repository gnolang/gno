package timestamp

import (
	"context"
	"flag"
	"fmt"
	"time"

	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

const (
	defaultRemoteAddress = "http://127.0.0.1:26657"
	defaultWebSocket     = true
	defaultCheckDuration = 30 * time.Second
	defaultCheckInterval = 50 * time.Millisecond
	defaultMaxDelta      = 10 * time.Second
	defaultVerbose       = false
)

type timestampCfg struct {
	remoteAddress string
	webSocket     bool
	checkDuration time.Duration
	checkInterval time.Duration
	maxDelta      time.Duration
	verbose       bool
}

// NewTimestampCmd creates the gnohealth timestamp subcommand
func NewTimestampCmd(io commands.IO) *commands.Command {
	cfg := &timestampCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "timestamp",
			ShortUsage: "timestamp [flags]",
			ShortHelp:  "check if block timestamps are drifting",
			LongHelp:   "This command checks if block timestamps are drifting on a blockchain by connecting to a specified node via RPC.",
		},
		cfg,
		func(_ context.Context, _ []string) error {
			return execTimestamp(cfg, io)
		},
	)
}

// RegisterFlags registers command-line flags for the timestamp command
func (c *timestampCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remoteAddress,
		"remote",
		defaultRemoteAddress,
		"the remote address of the node to connect to via RPC",
	)

	fs.BoolVar(
		&c.webSocket,
		"ws",
		defaultWebSocket,
		"flag indicating whether to use the WebSocket protocol for RPC",
	)

	fs.DurationVar(
		&c.checkDuration,
		"duration",
		defaultCheckDuration,
		"duration for which checks should be performed",
	)

	fs.DurationVar(
		&c.checkInterval,
		"interval",
		defaultCheckInterval,
		"interval between consecutive checks",
	)

	fs.DurationVar(
		&c.maxDelta,
		"max-delta",
		defaultMaxDelta,
		"maximum allowable time difference between the current time and the last block time",
	)

	fs.BoolVar(
		&c.verbose,
		"verbose",
		defaultVerbose,
		"flag indicating whether to enable verbose logging",
	)
}

func execTimestamp(cfg *timestampCfg, io commands.IO) error {
	var (
		client      *rpcClient.RPCClient
		err         error
		lastChecked int64
		count       uint64
		totalDelta  time.Duration
	)

	// Init RPC client
	if cfg.webSocket {
		if client, err = rpcClient.NewWSClient(cfg.remoteAddress); err != nil {
			return fmt.Errorf("unable to create WS client: %w", err)
		}
	} else {
		if client, err = rpcClient.NewHTTPClient(cfg.remoteAddress); err != nil {
			return fmt.Errorf("unable to create HTTP client: %w", err)
		}
	}

	// Create a ticker for check interval
	ticker := time.NewTicker(cfg.checkInterval)
	defer ticker.Stop()

	// Create a context that will stop this check when specified duration is elapsed
	ctx, cancel := context.WithTimeout(context.Background(), cfg.checkDuration)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			average := totalDelta / time.Duration(count)
			io.Printf("no timestamp drifted beyond the maximum delta (average %s)\n", average)
			return nil

		case <-ticker.C:
			// Fetch the latest block number from the chain
			status, err := client.Status(context.Background(), nil)
			if err != nil {
				return fmt.Errorf("unable to fetch latest block number: %w", err)
			}

			latest := status.SyncInfo.LatestBlockHeight

			// Check if there have been blocks since the last check
			if lastChecked == latest {
				continue
			}

			// Fetch the latest block from the chain
			lastBlock, err := client.Block(context.Background(), &latest)
			if err != nil {
				return fmt.Errorf("unable to fetch latest block content: %w", err)
			}

			// Check if the last block timestamp is not drifting
			delta := time.Until(lastBlock.Block.Time).Abs()
			if delta > cfg.maxDelta {
				return fmt.Errorf("block %d drifted of %s (max %s): KO", latest, delta, cfg.maxDelta)
			}

			// Increment counters to calculate average on exit
			count += 1
			totalDelta += delta

			// Update the last checked block number
			lastChecked = latest
			if cfg.verbose {
				io.Printf("block %d drifted of %s (max %s): OK\n", latest, delta, cfg.maxDelta)
			}
		}
	}
}
