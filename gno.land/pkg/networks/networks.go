// Package networks is the canonical registry of known gno.land networks.
//
// The JSON file networks.json is the single source of truth used by gnoweb
// (served at /api/networks), the CLI tools, docs, and downstream consumers.
// Update that file — not a hardcoded list elsewhere — when testnets change.
package networks

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed networks.json
var rawJSON []byte

// Network status values.
const (
	StatusActive     = "active"
	StatusDeprecated = "deprecated"
	StatusOffline    = "offline"
)

// Network describes a single gno.land network entry. Status is one of
// "active", "deprecated", or "offline".
type Network struct {
	Name        string `json:"name"`
	ChainID     string `json:"chain_id"`
	RPCEndpoint string `json:"rpc_endpoint"`
	GnowebURL   string `json:"gnoweb_url,omitempty"`
	FaucetURL   string `json:"faucet_url,omitempty"`
	Status      string `json:"status"`
}

// Registry is the top-level structure of networks.json.
type Registry struct {
	Networks []Network `json:"networks"`
}

// Raw returns a copy of the embedded networks.json bytes.
func Raw() []byte {
	b := make([]byte, len(rawJSON))
	copy(b, rawJSON)
	return b
}

// Load parses the embedded networks.json into a Registry and validates that
// each entry has a known status and a unique chain_id.
func Load() (Registry, error) {
	var r Registry
	if err := json.Unmarshal(rawJSON, &r); err != nil {
		return Registry{}, fmt.Errorf("parse networks.json: %w", err)
	}
	seen := make(map[string]bool, len(r.Networks))
	for _, n := range r.Networks {
		switch n.Status {
		case StatusActive, StatusDeprecated, StatusOffline:
		default:
			return Registry{}, fmt.Errorf("network %q has invalid status %q", n.ChainID, n.Status)
		}
		if seen[n.ChainID] {
			return Registry{}, fmt.Errorf("duplicate chain_id %q", n.ChainID)
		}
		seen[n.ChainID] = true
	}
	return r, nil
}
