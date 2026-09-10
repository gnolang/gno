package components

// NetworkKind tells the mainnet deployment apart from every other one.
type NetworkKind string

const (
	NetworkMainnet NetworkKind = "mainnet"
	NetworkTestnet NetworkKind = "testnet"
)

// MainnetChainID is the mainnet chain-id. Note the hyphen: `gnoland1` is the
// betanet, a different chain (docs/resources/gnoland-networks.md).
const MainnetChainID = "gnoland-1"

// NetworkKindFromChainID derives the kind from a chain-id. Anything
// unrecognised is a testnet: a testnet passing for mainnet is the dangerous
// direction, so an unknown chain gets the cautious answer.
func NetworkKindFromChainID(chainID string) NetworkKind {
	if chainID == MainnetChainID {
		return NetworkMainnet
	}
	return NetworkTestnet
}

func (k NetworkKind) IsMainnet() bool { return k == NetworkMainnet }

// Valid rejects a bad -network-kind rather than treating a typo as a testnet.
func (k NetworkKind) Valid() bool {
	return k == NetworkMainnet || k == NetworkTestnet
}
