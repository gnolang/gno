package core

// CreateClient generates a new client identifier and invokes the associated
// light client module in order to initialize the client.
func CreateClient(clientType string, clientState, consensusState []byte) string {
	// TODO
	return "clientID"
}

// RegisterCounterparty will register the IBC v2 counterparty info for the
// given clientID. It must be called by the same relayer that called
// CreateClient.
func RegisterCounterparty(clientID string, counterpartyMerklePrefix [][]byte, CounterpartyClientID string) {
	// TODO
}

// UpdateClient will update the given IBC v2 light client with a new header.
// Can also be used to submit a misbehavior (clientMessage can be a header or
// a misbehavior, maybe split into 2 functions would make more sense here).
func UpdateClient(clientID string, clientMessage []byte) {
	// TODO
}

// UpdateClient upgrades the client to a new client state if this new client
// was committed to by the old client at the specified upgrade height.
func UpgradeClient(clientID string, clientState, consensusState []byte,
	proofUpgradeClient, proofUpgradeConsensusState []byte) {
	// TODO
}

// RecoverClient recovers a frozen or expired client.
// NOTE(tb): This must be a permissioned function called by a gov proposal.
func RecoverClient(clientID string) {
	// TODO
}
