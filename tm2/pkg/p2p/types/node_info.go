package types

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/strings"
	"github.com/gnolang/gno/tm2/pkg/versionset"
)

const (
	MaxNodeInfoSize = int64(10240) // 10KB
	maxNumChannels  = 16           // plenty of room for upgrades, for now
)

var (
	ErrInvalidPeerID        = errors.New("invalid peer ID")
	ErrInvalidVersion       = errors.New("invalid node version")
	ErrInvalidMoniker       = errors.New("invalid node moniker")
	ErrInvalidRPCAddress    = errors.New("invalid node RPC address")
	ErrExcessiveChannels    = errors.New("excessive node channels")
	ErrDuplicateChannels    = errors.New("duplicate node channels")
	ErrIncompatibleNetworks = errors.New("incompatible networks")
	ErrNoCommonChannels     = errors.New("no common channels")
)

// NodeInfo is the basic node information exchanged
// between two peers during the Tendermint P2P handshake.
type NodeInfo struct {
	// Set of protocol versions
	VersionSet versionset.VersionSet `json:"version_set"`

	// Unique peer identifier
	PeerID ID `json:"id"`

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
// that can not be resolved to some IP
func (info NodeInfo) Validate() error {
	// Validate the ID
	if err := info.PeerID.Validate(); err != nil {
		return fmt.Errorf("%w, %w", ErrInvalidPeerID, err)
	}

	// Validate Version
	if len(info.Version) > 0 &&
		(!strings.IsASCIIText(info.Version) ||
			strings.ASCIITrim(info.Version) == "") {
		return ErrInvalidVersion
	}

	// Validate Channels - ensure max and check for duplicates.
	if len(info.Channels) > maxNumChannels {
		return ErrExcessiveChannels
	}

	channelMap := make(map[byte]struct{}, len(info.Channels))
	for _, ch := range info.Channels {
		if _, ok := channelMap[ch]; ok {
			return ErrDuplicateChannels
		}

		// Mark the channel as present
		channelMap[ch] = struct{}{}
	}

	// Validate Moniker.
	if !strings.IsASCIIText(info.Moniker) || strings.ASCIITrim(info.Moniker) == "" {
		return ErrInvalidMoniker
	}

	// XXX: Should we be more strict about address formats?
	rpcAddr := info.Other.RPCAddress
	if len(rpcAddr) > 0 && (!strings.IsASCIIText(rpcAddr) || strings.ASCIITrim(rpcAddr) == "") {
		return ErrInvalidRPCAddress
	}

	return nil
}

// ID returns the local node ID
func (info NodeInfo) ID() ID {
	return info.PeerID
}

// CompatibleWith checks if two NodeInfo are compatible with each other.
// CONTRACT: two nodes are compatible if the Block version and networks match,
// and they have at least one channel in common
func (info NodeInfo) CompatibleWith(other NodeInfo) error {
	// Validate the protocol versions
	if _, err := info.VersionSet.CompatibleWith(other.VersionSet); err != nil {
		return fmt.Errorf("incompatible version sets, %w", err)
	}

	// Make sure nodes are on the same network
	if info.Network != other.Network {
		return ErrIncompatibleNetworks
	}

	// Make sure there is at least 1 channel in common
	commonFound := false
	for _, ch1 := range info.Channels {
		for _, ch2 := range other.Channels {
			if ch1 == ch2 {
				commonFound = true

				break
			}
		}

		if commonFound {
			break
		}
	}

	if !commonFound {
		return ErrNoCommonChannels
	}

	return nil
}
