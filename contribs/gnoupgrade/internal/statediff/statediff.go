package statediff

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type statediffCfg struct {
	remote string
	output string
}

func NewStateDiffCmd(io commands.IO) *commands.Command {
	cfg := &statediffCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "statediff",
			ShortUsage: "statediff [flags]",
			ShortHelp:  "capture chain state snapshot in a diff-friendly format",
			LongHelp: `Connects to a running gnoland node and captures a snapshot of the
chain state in a canonical, sorted, diff-friendly JSON format. Run this
against the old and new binary to compare what changed after an upgrade.

Output includes:
  - Chain metadata (chain ID, height, app hash)
  - Realm/package listing (sorted by path)

Usage:
  # Snapshot before upgrade
  gnoupgrade statediff --remote http://old-node:26657 --output before.json

  # Snapshot after upgrade
  gnoupgrade statediff --remote http://new-node:26657 --output after.json

  # Compare
  diff before.json after.json`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execStateDiff(ctx, cfg, io)
		},
	)
}

func (c *statediffCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		"http://127.0.0.1:26657",
		"RPC address of the gnoland node",
	)

	fs.StringVar(
		&c.output,
		"output",
		"",
		"output file path (default: stdout)",
	)
}

// StateSnapshot is the canonical representation of chain state.
type StateSnapshot struct {
	ChainID  string          `json:"chain_id"`
	Height   int64           `json:"height"`
	AppHash  string          `json:"app_hash"`
	Realms   []RealmSnapshot `json:"realms,omitempty"`
	Packages []RealmSnapshot `json:"packages,omitempty"`
}

type RealmSnapshot struct {
	Path    string `json:"path"`
	Creator string `json:"creator,omitempty"`
}

func execStateDiff(ctx context.Context, cfg *statediffCfg, io commands.IO) error {
	client, err := rpcClient.NewHTTPClient(cfg.remote)
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}

	io.ErrPrintfln("Capturing state snapshot from %s...", cfg.remote)

	snapshot, err := captureSnapshot(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to capture snapshot: %w", err)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	if cfg.output != "" {
		if err := os.WriteFile(cfg.output, data, 0o644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		io.ErrPrintfln("Snapshot written to %s", cfg.output)
	} else {
		io.Println(string(data))
	}

	return nil
}

func captureSnapshot(ctx context.Context, client *rpcClient.RPCClient) (*StateSnapshot, error) {
	status, err := client.Status(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	snapshot := &StateSnapshot{
		ChainID: status.NodeInfo.Network,
		Height:  status.SyncInfo.LatestBlockHeight,
		AppHash: fmt.Sprintf("%X", status.SyncInfo.LatestAppHash),
	}

	// Query realms and packages via ABCI
	realms, err := queryPackageList(ctx, client, "gno.land/r/")
	if err == nil {
		snapshot.Realms = realms
	}

	packages, err := queryPackageList(ctx, client, "gno.land/p/")
	if err == nil {
		snapshot.Packages = packages
	}

	return snapshot, nil
}

func queryPackageList(ctx context.Context, client *rpcClient.RPCClient, prefix string) ([]RealmSnapshot, error) {
	qres, err := client.ABCIQuery(ctx, "vm/qpackages", nil)
	if err != nil {
		return nil, err
	}

	if qres.Response.IsErr() {
		return nil, errors.New(qres.Response.Log)
	}

	// Parse the response — each line is "path creator"
	var results []RealmSnapshot
	lines := strings.Split(string(qres.Response.Data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		parts := strings.Fields(line)
		entry := RealmSnapshot{Path: parts[0]}
		if len(parts) > 1 {
			entry.Creator = parts[1]
		}
		results = append(results, entry)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})

	return results, nil
}
