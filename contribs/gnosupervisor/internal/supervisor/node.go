package supervisor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/pebble"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/db/pebbledb"
)

// stateKey is where tm2 stores the consensus state in the "state" database
// (tm2/pkg/bft/state/state.go). TestCommittedHeight writes the state through
// sm.SaveState, so a rename there fails the test rather than this silently
// reading nothing.
var stateKey = []byte("stateKey")

// Node is the gnoland data directory the supervisor runs a node on.
type Node struct {
	DataDir string
	cfg     *config.Config
}

// OpenNode reads the node configuration under dataDir and refuses the
// settings a supervised replay cannot run with. A data directory with no
// config yet (gnoland start --lazy writes one) gets tm2's defaults.
func OpenNode(dataDir string) (*Node, error) {
	abs, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, err
	}
	cfg, err := config.LoadConfig(abs)
	if err != nil {
		if _, statErr := os.Stat(filepath.Join(abs, "config", "config.toml")); !errors.Is(statErr, os.ErrNotExist) {
			return nil, fmt.Errorf("load node config: %w", err)
		}
		cfg = config.DefaultConfig()
		cfg.SetRootDir(abs)
	}

	// halt_height in config.toml stops the node at a height the ledger knows
	// nothing about, and skip_upgrade_height switches off the startup checks
	// that refuse the wrong binary on either side of a governance halt. A
	// replay that follows the ledger needs neither, and with either set the
	// supervisor could not tell a ledger halt from something else.
	if cfg.HaltHeight != 0 {
		return nil, fmt.Errorf("halt_height is %d in %s; set it to 0 to run under the supervisor", cfg.HaltHeight, configPath(abs))
	}
	if cfg.SkipUpgradeHeight != 0 {
		return nil, fmt.Errorf("skip_upgrade_height is %d in %s; set it to 0 to run under the supervisor", cfg.SkipUpgradeHeight, configPath(abs))
	}
	if cfg.DBBackend != pebbledbBackend {
		return nil, fmt.Errorf("db_backend is %q; the supervisor reads the committed height of pebbledb nodes only", cfg.DBBackend)
	}
	return &Node{DataDir: abs, cfg: cfg}, nil
}

const pebbledbBackend = "pebbledb"

func configPath(root string) string { return filepath.Join(root, "config", "config.toml") }

// CommittedHeight is the last block the node committed, 0 for a node that
// has not started yet. It must only be called while no gnoland process runs
// on the data directory.
//
// It reads the consensus state record, which the next start's handshake
// resumes from. At a governance halt it holds the halt height: the halt
// panics in the next block's BeginBlock, before that block's state is
// saved (its block may already be in the block store). Released binaries
// save the state with a synced write every block; on this tree, block sync
// defers its writes but takes a flush point at the halt height
// (tm2/adr/pr6334_deferred_block_sync_writes.md), and after a stop between
// flush points the record holds the last flush point, where gnoland resumes
// too. The whole sm.State is decoded with this tree's amino types, so a
// future change to that record must keep older ones decodable here.
// The database is opened read-only: the supervisor never writes node data.
func (n *Node) CommittedHeight() (int64, error) {
	st, err := n.loadState()
	return st.LastBlockHeight, err
}

// CheckChain refuses a data directory whose state belongs to another chain.
// gnoland reads the genesis from its database once it has one, so checking
// the genesis file alone would not catch a data directory of another chain.
func (n *Node) CheckChain(chainID string) error {
	st, err := n.loadState()
	if err != nil {
		return err
	}
	if st.ChainID != "" && st.ChainID != chainID {
		return fmt.Errorf("the node's state in %s is for chain %q, the ledger is for %q", n.DataDir, st.ChainID, chainID)
	}
	return nil
}

// loadState returns the consensus state record, or a zero State for a node
// that has not started yet.
func (n *Node) loadState() (sm.State, error) {
	var st sm.State
	dir := n.cfg.DBDir()
	if _, err := os.Stat(filepath.Join(dir, "state.db")); errors.Is(err, os.ErrNotExist) {
		return st, nil
	}
	db, err := pebbledb.NewPebbleDBWithOpts("state", dir, &pebble.Options{ReadOnly: true})
	if err != nil {
		return st, fmt.Errorf("open state db: %w", err)
	}
	defer db.Close()

	buf, err := db.Get(stateKey)
	if err != nil {
		return st, fmt.Errorf("read state: %w", err)
	}
	if len(buf) == 0 {
		return st, nil
	}
	// sm.LoadState exits the process on a decode error; decode here instead.
	if err := amino.Unmarshal(buf, &st); err != nil {
		return sm.State{}, fmt.Errorf("decode state: %w", err)
	}
	return st, nil
}

// LiveHeight returns a reader of the running node's latest block height over
// its RPC (/status), or nil if the RPC does not listen on TCP.
func (n *Node) LiveHeight() LiveHeight {
	base := n.rpcBase()
	if base == "" {
		return nil
	}
	url := base + "/status"
	client := &http.Client{Timeout: 5 * time.Second}
	return func(ctx context.Context) (int64, error) {
		return rpcHeight(ctx, client, url)
	}
}

// rpcBase is the node's RPC as an http URL, asked locally, or "" if the RPC
// does not listen on TCP.
func (n *Node) rpcBase() string {
	addr, ok := strings.CutPrefix(n.cfg.RPC.ListenAddress, "tcp://")
	if !ok {
		return ""
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	if ip := net.ParseIP(host); host == "" || (ip != nil && ip.IsUnspecified()) {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port)
}

// HaltParam returns a reader of the chain's node:p:halt_height over the
// running node's RPC, or nil if the RPC does not listen on TCP.
func (n *Node) HaltParam() LiveHeight {
	base := n.rpcBase()
	if base == "" {
		return nil
	}
	url := base + `/abci_query?path=%22params/node:p:halt_height%22`
	client := &http.Client{Timeout: 5 * time.Second}
	return func(ctx context.Context) (int64, error) {
		var r struct {
			Result struct {
				Response struct {
					ResponseBase struct {
						Data []byte `json:"Data"`
					} `json:"ResponseBase"`
				} `json:"response"`
			} `json:"result"`
		}
		if err := getJSON(ctx, client, url, &r); err != nil {
			return 0, err
		}
		data := r.Result.Response.ResponseBase.Data
		if len(data) == 0 {
			return 0, nil // never set
		}
		var v string // the param is an int64, encoded as a JSON string
		if err := json.Unmarshal(data, &v); err != nil {
			return 0, fmt.Errorf("halt_height param %q: %w", data, err)
		}
		return strconv.ParseInt(v, 10, 64)
	}
}

func getJSON(ctx context.Context, client *http.Client, url string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(v)
}

func rpcHeight(ctx context.Context, client *http.Client, url string) (int64, error) {
	var st struct {
		Result struct {
			SyncInfo struct {
				Height string `json:"latest_block_height"`
			} `json:"sync_info"`
		} `json:"result"`
	}
	if err := getJSON(ctx, client, url, &st); err != nil {
		return 0, err
	}
	return strconv.ParseInt(st.Result.SyncInfo.Height, 10, 64)
}
