package types

import (
	"errors"
	"fmt"
	"slices"

	"github.com/gnolang/gno/tm2/pkg/strings"
	"github.com/gnolang/gno/tm2/pkg/versionset"
)

const (
	MaxNodeInfoSize = int64(10240) // 10KB
	maxNumChannels  = 16           // plenty of room for upgrades, for now
)

var (
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

	// The advertised net address of the peer
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
// if the NetAddress is malformed, or if the NetAddress is a host name
// that can not be resolved to some IP
func (info NodeInfo) Validate() error {
	// There are a few checks that need to be performed when validating
	// the node info's net address:
	// - the ID needs to be valid
	// - the FORMAT of the net address needs to be valid
	//
	// The key nuance here is that the net address is not being validated
	// for its "dialability", but whether it's of the correct format.
	//
	// Unspecified IPs are tolerated (ex. 0.0.0.0 or ::),
	// because of legacy logic that assumes node info
	// can have unspecified IPs (ex. no external address is set, use
	// the listen address which is bound to 0.0.0.0).
	//
	// These types of IPs are caught during the
	// real peer info sharing process, since they are undialable
	_, err := NewNetAddressFromString(NetAddressString(info.NetAddress.ID, info.NetAddress.DialString()))
	if err != nil {
		return fmt.Errorf("invalid net address in node info, %w", err)
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
	return info.NetAddress.ID
}

// DialAddress is the advertised peer dial address (share-able)
func (info NodeInfo) DialAddress() *NetAddress {
	return info.NetAddress
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
		if slices.Contains(other.Channels, ch1) {
			commonFound = true
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
