package node

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/strings"
)

var (
	errInvalidID   = errors.New("invalid node ID")
	errInvalidRole = errors.New("invalid node role")
)

// Role is the specific node label, based on its function
type Role string

const (
	// Bootnode is a rendezvous node for the cluster (non-validator)
	Bootnode Role = "bootnode"

	// Validator is the consensus participant for the cluster
	Validator Role = "validator"

	// Peer is a non-validator node (like a sentry, RPC...)
	Peer Role = "peer"
)

// Config is the base single-node configuration.
// Configurations for node ID "*" are applied to all nodes, with
// the node's explicit configuration taking precedence
type Config struct {
	ID       string         `yaml:"id"`     // the unique ID of the node
	Role     Role           `yaml:"role"`   // the role of the node in the cluster
	TMConfig map[string]any `yaml:"config"` // tm2-specific config.toml values
}

func (c Config) Validate() error {
	// Make sure the ID is non-empty and contains only ascii
	if c.ID == "" || !strings.IsASCIIText(c.ID) {
		return fmt.Errorf("%w: %q", errInvalidID, c.ID)
	}

	// Make sure the role is valid
	if c.Role != Bootnode && c.Role != Validator && c.Role != Peer {
		return fmt.Errorf("%w: %q", errInvalidRole, c.Role)
	}

	return nil
}
