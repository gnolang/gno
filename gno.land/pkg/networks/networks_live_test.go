package networks

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

// TestActiveNetworksReachable hits each active network's RPC /status endpoint
// and verifies the reported chain_id matches the registry. Detects a stale
// networks.json (renamed/retired testnet, wrong RPC URL, etc.).
//
// Skipped under -short to keep the default test run hermetic; CI can run it
// on a schedule by invoking `go test ./gno.land/pkg/networks/`.
func TestActiveNetworksReachable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network check in -short mode")
	}
	t.Parallel()

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
	resp, err := c.Get(rpcEndpoint + "/status")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
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
