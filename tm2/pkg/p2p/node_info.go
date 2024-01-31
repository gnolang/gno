package p2p

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore"
	"github.com/gnolang/gno/tm2/pkg/strings"
	"github.com/gnolang/gno/tm2/pkg/versionset"
)

const (
	maxNodeInfoSize = 10240 // 10KB
	maxNumChannels  = 16    // plenty of room for upgrades, for now
)

// Max size of the NodeInfo struct
func MaxNodeInfoSize() int {
	return maxNodeInfoSize
}

// -------------------------------------------------------------

// NodeInfo is the basic node information exchanged
// between two peers during the Tendermint P2P handshake.
type NodeInfo struct {
	// Set of protocol versions
	VersionSet versionset.VersionSet `json:"version_set"`

	// Authenticate
	NetAddress *NetAddress `json:"net_address"`

	// Check compatibility.
	// Channels are HexBytes so easier to read as JSON
	Network  string `json:"network"`  // network/chain ID
	Software string `json:"software"` // name of immediate software
	Version  string `json:"version"`  // software major.minor.revision
	Channels []byte `json:"channels"` // channels this node knows about

	// ASCIIText fields
	Moniker string        `json:"moniker"` // arbitrary moniker
	Other   NodeInfoOther `json:"other"`   // other application specific data
}

// NodeInfoOther is the misc. application specific data
type NodeInfoOther struct {
	TxIndex    string `json:"tx_index"`
	RPCAddress string `json:"rpc_address"`
}

// Validate checks the self-reported NodeInfo is safe.
// It returns an error if there
// are too many Channels, if there are any duplicate Channels,
// if the ListenAddr is malformed, or if the ListenAddr is a host name
// that can not be resolved to some IP.
// TODO: constraints for Moniker/Other? Or is that for the UI ?
// JAE: It needs to be done on the client, but to prevent ambiguous
// unicode characters, maybe it's worth sanitizing it here.
// In the future we might want to validate these, once we have a
// name-resolution system up.
// International clients could then use punycode (or we could use
// url-encoding), and we just need to be careful with how we handle that in our
// clients. (e.g. off by default).
func (info NodeInfo) Validate() error {
	// ID is already validated. TODO validate

	// Validate ListenAddr.
	if info.NetAddress == nil {
		return fmt.Errorf("info.NetAddress cannot be nil")
	}
	if err := info.NetAddress.ValidateLocal(); err != nil {
		return err
	}

	// Network is validated in CompatibleWith.

	// Validate Version
	if len(info.Version) > 0 &&
		(!strings.IsASCIIText(info.Version) || strings.ASCIITrim(info.Version) == "") {
		return fmt.Errorf("info.Version must be valid ASCII text without tabs, but got %v", info.Version)
	}

	// Validate Channels - ensure max and check for duplicates.
	if len(info.Channels) > maxNumChannels {
		return fmt.Errorf("info.Channels is too long (%v). Max is %v", len(info.Channels), maxNumChannels)
	}
	channels := make(map[byte]struct{})
	for _, ch := range info.Channels {
		_, ok := channels[ch]
		if ok {
			return fmt.Errorf("info.Channels contains duplicate channel id %v", ch)
		}
		channels[ch] = struct{}{}
	}

	// Validate Moniker.
	if !strings.IsASCIIText(info.Moniker) || strings.ASCIITrim(info.Moniker) == "" {
		return fmt.Errorf("info.Moniker must be valid non-empty ASCII text without tabs, but got %v", info.Moniker)
	}

	// Validate Other.
	other := info.Other
	txIndex := other.TxIndex
	switch txIndex {
	case "", eventstore.StatusOn, eventstore.StatusOff:
	default:
		return fmt.Errorf("info.Other.TxIndex should be either 'on', 'off', or empty string, got '%v'", txIndex)
	}
	// XXX: Should we be more strict about address formats?
	rpcAddr := other.RPCAddress
	if len(rpcAddr) > 0 && (!strings.IsASCIIText(rpcAddr) || strings.ASCIITrim(rpcAddr) == "") {
		return fmt.Errorf("info.Other.RPCAddress=%v must be valid ASCII text without tabs", rpcAddr)
	}

	return nil
}

func (info NodeInfo) ID() ID {
	return info.NetAddress.ID
}

// CompatibleWith checks if two NodeInfo are compatible with eachother.
// CONTRACT: two nodes are compatible if the Block version and network match
// and they have at least one channel in common.
func (info NodeInfo) CompatibleWith(other NodeInfo) error {
	// check protocol versions
	_, err := info.VersionSet.CompatibleWith(other.VersionSet)
	if err != nil {
		return err
	}

	// nodes must be on the same network
	if info.Network != other.Network {
		return fmt.Errorf("Peer is on a different network. Got %v, expected %v", other.Network, info.Network)
	}

	// if we have no channels, we're just testing
	if len(info.Channels) == 0 {
		return nil
	}

	// for each of our channels, check if they have it
	found := false
OUTER_LOOP:
	for _, ch1 := range info.Channels {
		for _, ch2 := range other.Channels {
			if ch1 == ch2 {
				found = true
				break OUTER_LOOP // only need one
			}
		}
	}
	if !found {
		return fmt.Errorf("Peer has no common channels. Our channels: %v ; Peer channels: %v", info.Channels, other.Channels)
	}
	return nil
}
