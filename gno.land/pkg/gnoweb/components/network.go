package components

// NetworkKind tells the mainnet deployment apart from every other one. It is
// set by the operator (-network-kind); gnoweb does not guess it from the
// chain-id, so no chain-naming assumption is encoded here. The default is
// testnet: a mainnet that forgets the flag shows the alert chip, the safe
// direction, while a testnet can only present as mainnet by explicit
// misconfiguration.
type NetworkKind string

const (
	NetworkMainnet NetworkKind = "mainnet"
	NetworkTestnet NetworkKind = "testnet"
)

func (k NetworkKind) IsMainnet() bool { return k == NetworkMainnet }

// Valid rejects a bad -network-kind rather than treating a typo as a testnet.
func (k NetworkKind) Valid() bool {
	return k == NetworkMainnet || k == NetworkTestnet
}
