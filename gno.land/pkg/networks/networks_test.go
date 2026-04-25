package networks

import (
	"encoding/json"
	"testing"
)

func TestLoad(t *testing.T) {
	reg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(reg.Networks) == 0 {
		t.Fatal("expected at least one network")
	}

	seen := map[string]bool{}
	for _, n := range reg.Networks {
		if n.Name == "" {
			t.Errorf("network missing name: %+v", n)
		}
		if n.ChainID == "" {
			t.Errorf("network %q missing chain_id", n.Name)
		}
		if n.RPCEndpoint == "" {
			t.Errorf("network %q missing rpc_endpoint", n.Name)
		}
		if n.Status == "" {
			t.Errorf("network %q missing status", n.Name)
		}
		if seen[n.ChainID] {
			t.Errorf("duplicate chain_id: %q", n.ChainID)
		}
		seen[n.ChainID] = true
	}
}

func TestRawIsValidJSON(t *testing.T) {
	var v any
	if err := json.Unmarshal(Raw(), &v); err != nil {
		t.Fatalf("networks.json is not valid JSON: %v", err)
	}
}

func TestRegistryContainsBetanet(t *testing.T) {
	reg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, n := range reg.Networks {
		if n.ChainID == "gnoland1" {
			return
		}
	}
	t.Error("chain_id gnoland1 not found in registry")
}
