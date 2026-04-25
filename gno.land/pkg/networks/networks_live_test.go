package networks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestActiveNetworksReachable hits each active network's RPC /status endpoint
// and verifies the reported chain_id matches the registry. Detects a stale
// networks.json (renamed/retired testnet, wrong RPC URL, etc.).
//
// Opt-in: skipped unless GNO_NETWORKS_LIVE=1 is set. Keeps the default
// `go test ./...` (and `make test`) hermetic. The dedicated make target
// `_test.networks.live` and the scheduled CI workflow set it explicitly.
func TestActiveNetworksReachable(t *testing.T) {
	t.Parallel()
	if os.Getenv("GNO_NETWORKS_LIVE") != "1" {
		t.Skip("skipping live network check; set GNO_NETWORKS_LIVE=1 to run")
	}

	reg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	for _, n := range reg.Networks {
		if n.Status != StatusActive {
			continue
		}
		t.Run(n.ChainID, func(t *testing.T) {
			t.Parallel()
			got, err := fetchChainID(client, n.RPCEndpoint)
			if err != nil {
				t.Fatalf("%s: %v", n.RPCEndpoint, err)
			}
			if got != n.ChainID {
				t.Fatalf("registry chain_id=%q but %s reports %q (networks.json out of date?)",
					n.ChainID, n.RPCEndpoint, got)
			}
		})
	}
}

func fetchChainID(c *http.Client, rpcEndpoint string) (string, error) {
	url := rpcEndpoint + "/status"
	resp, err := c.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d from %s", resp.StatusCode, url)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var parsed struct {
		Result struct {
			NodeInfo struct {
				Network string `json:"network"`
			} `json:"node_info"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	return parsed.Result.NodeInfo.Network, nil
}
