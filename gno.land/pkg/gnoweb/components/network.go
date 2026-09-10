package components

import "regexp"

// NetworkKind tells the mainnet deployment apart from every other one.
type NetworkKind string

const (
	NetworkMainnet NetworkKind = "mainnet"
	NetworkTestnet NetworkKind = "testnet"
)

// mainnetChainID matches the mainnet series: gnoland-1 today, gnoland-2,
// gnoland-3 if the chain is ever restarted. Note the hyphen: `gnoland1` is
// the betanet, a different chain (docs/resources/gnoland-networks.md).
var mainnetChainID = regexp.MustCompile(`^gnoland-[1-9][0-9]*$`)

// NetworkKindFromChainID derives the kind from a chain-id. Anything
// unrecognised is a testnet: a testnet passing for mainnet is the dangerous
// direction, so an unknown chain gets the cautious answer.
func NetworkKindFromChainID(chainID string) NetworkKind {
	if mainnetChainID.MatchString(chainID) {
		return NetworkMainnet
	}
	return NetworkTestnet
}

func (k NetworkKind) IsMainnet() bool { return k == NetworkMainnet }

// Valid rejects a bad -network-kind rather than treating a typo as a testnet.
func (k NetworkKind) Valid() bool {
	return k == NetworkMainnet || k == NetworkTestnet
}
